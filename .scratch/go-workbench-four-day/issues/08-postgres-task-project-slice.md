# 08 — 让 Task 列表从 PostgreSQL 读取

**要做的事：** 把 `GET /v1/tasks` 从本地 fixture 数据切到 PostgreSQL 数据。浏览器打开 Task 页面时，请求仍然走现有 HTTP API，但 API 返回的数据来自 `tasks` 表，并且只返回当前 Owner 的 Task。`status` query 要从浏览器请求一路传到 SQL 查询。

**Blocked by:** 07 — PostgreSQL migration 和隔离测试库 helper 已可用。

**Status:** ready-for-agent

## 08 的主流程

用户打开 Task 页面。前端调用 `GET /v1/tasks`，可以不带 `status`，也可以带一个状态，例如 `GET /v1/tasks?status=backlog`。

HTTP handler 先读取当前请求中的 `ownerID`。没有 `ownerID` 时返回 `401`。handler 再读取 `status` query。没有 `status` 时表示不过滤状态；有合法 `status` 时生成 task 层的 filter；有非法 `status` 时返回 `400`，并且不能调用 repository。

handler 调用 `task.Service.List(ctx, ownerID, filter)`。service 不做 SQL，也不重新解释 HTTP 参数，只把 owner 和 filter 传给 repository。

PostgreSQL Task repository 接收 owner 和 filter。它查询 `tasks` 表。SQL 必须包含 `owner_id` 条件。如果 filter 里有 `Status`，SQL 还要按 `status` 过滤。repository 使用 `rows.Scan` 把数据库行转换成 `task.Task`，返回给 service。

handler 把 `[]task.Task` 转成当前已有的 JSON 响应。Web 页面继续读取 `items` 并显示列表。Network 中的 `/v1/tasks` 响应不能包含其他 Owner 的 Task。

## 需要改出的代码形状

`internal/task/service.go`

- `Status` 保留当前枚举值：`backlog`、`in_progress`、`blocked`、`done`。
- 新增或保留 `ListTasksFilter`，字段使用 `Status *Status`。`nil` 表示没有 status filter；非 nil 表示只查一个状态。
- `Repository.List` 接收 `ctx`、`ownerID`、`ListTasksFilter`。
- `Service.List` 接收同样的 filter，并把它传给 repository。

`internal/httpapi/task_handler.go`

- 继续读取 `ownerIDKey`。
- 继续校验 `status` query。
- 合法 `status` 转成 `task.Status`，写入 `task.ListTasksFilter{Status: &status}`。
- 空 `status` 使用零值 `task.ListTasksFilter{}`。
- 调用 `service.List(c.Request.Context(), ownerID, filter)`。
- 不写 SQL，不访问 PostgreSQL。

`internal/httpapi/router_test.go`

- `stubTaskRepository.List` 的签名要跟 `task.Repository` 一致。
- stub 记录收到的 `ownerID` 和 filter。
- 增加或补强测试：请求 `/v1/tasks?status=backlog` 后，stub 收到 `StatusBacklog`。
- 保留现有错误路径测试：非法 status 返回 `400`，缺少 owner 返回 `401`，repository error 和 panic 返回稳定 `500`。

`cmd/api/local_repository.go`

- 如果还保留本地 fixture，它的 `List` 签名也要接收 filter。
- success fixture 至少支持 `StatusBacklog`：请求 backlog 返回当前 success task，请求其他 status 返回空列表。
- 这个文件只用于本地 fixture，不代表 08 的最终数据来源。

`cmd/api/local_repository_test.go`

- 更新调用签名。
- 覆盖 success、empty、error。
- 补一个最小 status filter 行为：success fixture 在非 backlog filter 下返回空列表。

`internal/postgres/task_repository.go`

- 新增 PostgreSQL Task repository。
- 提供构造函数，供 `cmd/api/main.go` 注入到 `task.NewService`。
- 查询 `tasks` 表。
- SQL 用参数绑定 owner 和 status。不要拼接用户输入。
- 每次查询都使用传入的 `ctx`。
- 关闭 `Rows`，并检查 `rows.Err()`。
- 映射到 `task.Task`：至少包括 `id`、`owner_id`、`title`、`status`、`priority`、`labels`、`version`、`created_at`、`updated_at`。

`internal/postgres/task_repository_test.go`

- 新增查询集成测试。
- 使用 `newTestDatabase` 和 `applyBusinessMigration` 创建隔离 schema。
- 这个测试验证 Task repository 查询代码，不重复验证 migration runner 或 schema 约束。
- 插入 Owner A、Owner B。
- 插入 A/backlog、A/done、B/backlog 或 B/done。
- 调用 `TaskRepository.List` 验证三件事：Owner A 查不到 Owner B；无 Task 的 Owner 返回空 slice；`StatusBacklog` 和 `StatusDone` filter 返回对应状态。

`cmd/api/main.go`

- 默认运行路径要使用 PostgreSQL Task repository。
- 如果保留 `-fixture`，它只能作为显式开发模式。
- PostgreSQL 连接信息从环境变量读取。不要读取真实 `.env` 文件，不打印连接字符串。

`web/app/task-list.tsx`

- 原则上不改。
- 只有 JSON 字段变化导致页面无法显示时才改。08 不应主动改 UI。

`openapi/openapi.yaml`

- 原则上不改。
- 当前 OpenAPI 已有 `GET /v1/tasks` 和单个 `status` query。08 不新增多个 status。

## 实现顺序

### 1. 先让 task 层能表达查询条件

修改 `internal/task/service.go` 和 `internal/task/service_test.go`。

先确认 `ListTasksFilter` 的形状：

```go
type ListTasksFilter struct {
	Status *Status
}
```

然后让 `Repository.List` 和 `Service.List` 都接收这个 filter。`service_test.go` 的 stub repository 要记录收到的 filter。测试要证明 service 没有丢掉 ownerID，也没有丢掉 status filter。

验证命令：

```bash
go test ./internal/task -count=1
```

通过结果：`internal/task` 通过。其他包此时可能因为接口签名变化暂时编译失败，下一步修复。

### 2. 再让 HTTP status query 进入 task filter

修改 `internal/httpapi/task_handler.go` 和 `internal/httpapi/router_test.go`。

handler 的心智流程是：

1. 没有 owner，返回 `401`。
2. 没有 status，创建空 filter。
3. status 是 `backlog`、`in_progress`、`blocked`、`done`，创建带 Status 指针的 filter。
4. status 是其他值，返回 `400`，不调用 repository。
5. 调用 `service.List(ctx, ownerID, filter)`。

router 测试要看见这个传递结果。`stubTaskRepository` 记录 filter。新增或补强一个测试：请求 `/v1/tasks?status=backlog` 后，`repository.gotFilter.Status` 不是 nil，并且值是 `task.StatusBacklog`。

验证命令：

```bash
go test ./internal/httpapi -count=1
```

通过结果：HTTP 测试通过；非法 status、缺少 owner、repository error、repository panic 的响应保持原行为。

### 3. 处理 cmd/api 里的本地 fixture

修改 `cmd/api/local_repository.go` 和 `cmd/api/local_repository_test.go`。

这一步只解决接口签名和本地开发 fixture。`LocalRepository.List` 要接收 filter。success fixture 可以只返回 backlog task；如果 filter 是其他 status，返回空 slice。

验证命令：

```bash
go test ./cmd/api -count=1
```

通过结果：`cmd/api` 编译通过；本地 fixture 测试不会再因为 `Repository.List` 签名变化失败。

### 4. 写 PostgreSQL Task repository 的 RED 测试

新增 `internal/postgres/task_repository_test.go`。

测试准备流程：

1. 用 `newTestDatabase` 创建隔离 schema。
2. 用 `applyBusinessMigration` 建表。
3. 插入 Owner A 和 Owner B。
4. 插入 A/backlog、A/done、B/backlog 或 B/done。
5. 创建 `TaskRepository`。
6. 调用 `List(ctx, ownerA, ListTasksFilter{})`，断言只返回 A 的 Task。
7. 调用 `List(ctx, emptyOwner, ListTasksFilter{})`，断言返回空 slice。
8. 调用 `List(ctx, ownerA, ListTasksFilter{Status: &StatusBacklog})`，断言只返回 A 的 backlog。
9. 调用 `List(ctx, ownerA, ListTasksFilter{Status: &StatusDone})`，断言只返回 A 的 done。

验证命令：

```bash
go test ./internal/postgres -run 'Test.*Task.*Repository.*List' -count=1
```

RED 结果：测试应因 `TaskRepository` 尚未实现而失败。失败不能来自数据库没启动、migration 失败、fixture 写错或测试连接到错误数据库。

### 5. 实现 PostgreSQL Task repository

新增或补全 `internal/postgres/task_repository.go`。

实现流程：

1. repository 持有 pgx 连接或当前项目接受的 pgx 查询接口。
2. `NewTaskRepository` 返回可注入 `task.NewService` 的 repository。
3. `List` 根据 filter 组织查询。
4. 无 status filter 时查询当前 owner 的 Task。
5. 有 status filter 时查询当前 owner 且 status 匹配的 Task。
6. `rows.Scan` 映射到 `task.Task`。
7. 返回空结果时使用空 slice。
8. 数据库错误向上返回，交给 HTTP 层映射成 `500`。

实现检查点：

- SQL 中必须出现 `owner_id` 条件。
- owner 和 status 都通过参数传入。
- 不拼接用户输入。
- `Rows` 被关闭。
- `rows.Err()` 被检查。
- `ctx` 传给 pgx 查询。

验证命令：

```bash
go test ./internal/postgres -run 'Test.*Task.*Repository.*List' -count=1
```

通过结果：Task repository 查询测试通过。

### 6. 把 API 默认运行路径接到 PostgreSQL

修改 `cmd/api/main.go`。

启动流程要变成：

1. 读取 PostgreSQL 连接信息的环境变量。
2. 创建 PostgreSQL 连接。
3. 创建 `postgres.TaskRepository`。
4. 创建 `task.Service`。
5. 创建 router。
6. 启动 API。

如果继续保留 fixture 模式，要求显式传 `-fixture`。不传 `-fixture` 时，默认使用 PostgreSQL。

验证命令：

```bash
make test-cmd-api
```

通过结果：`cmd/api` 测试通过，默认运行路径不再依赖 `LocalRepository`。

### 7. 跑包级回归测试

先跑已经被 08 改动影响的包：

```bash
go test ./internal/task ./internal/httpapi ./internal/postgres ./cmd/api -count=1
```

通过结果：

- task 层能传递 filter。
- HTTP 层能把 status query 转成 filter。
- PostgreSQL repository 能按 owner 和 status 查询。
- cmd/api 能编译并接入 repository。

### 8. 跑完整 Go 测试

```bash
make test
```

通过结果：全部 Go 测试通过。`make test` 使用 `TEST_DATABASE_URL`。如果失败原因是本地 PostgreSQL 没启动，记录为环境阻塞。不要把未运行或环境失败写成通过。

### 9. 浏览器验证

准备本地 PostgreSQL，运行 migration，插入浏览器验证数据：

```bash
make test-db-up
make run-migrate-test-db
make run-seed-test-db
```

这些 `run-*-test-db` 命令使用调用者显式导出的 `TEST_DATABASE_URL`。`make run-seed-test-db` 会把 `TEST_DATABASE_URL` 注入给 `cmd/seed` 的 `DATABASE_URL`。`cmd/seed` 会插入 Owner A、Owner B、A 的 backlog/done Task、B 的 backlog Task。API 本地 owner 使用 Owner A 的固定 UUID，所以浏览器页面应该只看到 Owner A 的 Task。

启动 API。确认 API 使用 PostgreSQL repository。启动 Web。打开 Task 页面，查看浏览器 Network 中 `/v1/tasks` 的响应。

```bash
make run-api-test-db
```

当前 Web 页面只展示 Task 列表，不提供 status filter 控件。08 的浏览器验证只确认页面能显示 PostgreSQL 数据和 owner 隔离。`status` filter 通过 API 请求验证，例如直接请求 `/v1/tasks?status=backlog`，或在浏览器 Network/API 客户端中确认响应只包含 backlog Task。前端筛选控件放到后续 UI 切片。

通过结果：

- 页面显示 PostgreSQL 中的 Task。
- 响应中没有 Owner B 的 Task。
- 直接请求 `/v1/tasks?status=backlog` 时，响应只包含 backlog Task。
- Web 不需要为了 08 主动改 UI。

## 最终验收

- [ ] `GET /v1/tasks` 默认数据源是 PostgreSQL。
- [ ] `GET /v1/tasks?status=backlog` 的 status 从 HTTP handler 传到 repository。
- [ ] SQL 查询显式使用 owner scope。
- [ ] Owner A 查不到 Owner B 的 Task。
- [ ] 无 Task 的 Owner 返回空 slice，并最终表现为 `items: []`。
- [ ] status filter 只支持单个 status，不实现多个 status。
- [ ] PostgreSQL repository 查询测试通过。
- [ ] HTTP 回归测试通过。
- [ ] `make test` 已运行并通过，或明确记录环境阻塞。
- [ ] 浏览器 Network 验证过 PostgreSQL 数据、owner 隔离和 status filter。
