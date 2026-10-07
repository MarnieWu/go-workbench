# 08 — 从 PostgreSQL 查询当前 Owner 的 Task

**目标结果：** `GET /v1/tasks` 不再读取本地 fixture，而是从 PostgreSQL 查询当前 Owner 的 Task。Owner A 永远看不到 Owner B 的数据。`status` query 必须真正影响查询结果，而不是只在 handler 层校验。

**Blocked by:** 07 完成并通过对应门禁。

**Status:** ready-for-agent

## 前置条件

- 07 已经建立可用的 PostgreSQL migration 和隔离测试库 helper。
- 本地测试只使用 `TEST_DATABASE_URL` 指向的安全测试库。不要读取真实 `.env`、密钥或生产配置。
- `owners` 和 `tasks` 表已经能通过 migration 创建。
- 本任务只做 Task list 的 PostgreSQL read path。不实现创建、更新、删除、认证、Project 页面或 Candidate 流程。

## 本次必须做的契约决定

1. `GET /v1/tasks?status=` 保持现有 OpenAPI 语义：无 `status` 返回当前 Owner 的全部未归档 Task；有 `status` 只返回该状态 Task。
2. handler 继续负责 HTTP 参数校验和错误映射，但不能包含 SQL。
3. service/repository 接口必须能接收 status filter。可以使用 `ListTasksFilter`、`ListOptions` 或同等清晰命名。
4. PostgreSQL repository 必须放在数据访问层，例如 `internal/postgres`。不要把 SQL 写进 `cmd/api`、handler 或 service。
5. `cmd/api` 要从本地 fixture repository 切到 PostgreSQL repository，或显式保留 fixture 模式并提供清楚的启动参数。浏览器验证必须能看到 PostgreSQL 数据。

## 预期修改文件

- `internal/task/service.go`
  - 修改 `Repository` 接口，让 `List` 接收 owner 和 status filter。
  - 修改 `Service.List`，把 filter 继续传给 repository。
  - 如有需要，新增 `ListTasksFilter` 或同等类型。这个类型应放在 task 包，因为 handler、service、repository 都会使用它。
- `internal/task/service_test.go`
  - 更新 stub repository 的方法签名。
  - 增加或调整测试，证明 service 会把 ownerID 和 status filter 原样传给 repository。
- `internal/httpapi/task_handler.go`
  - 保留当前 `status` query 校验。
  - 把合法 status 转成 task 层 filter，传给 `service.List`。
  - 不在这里写 SQL，不在这里做数据库查询。
- `internal/httpapi/router_test.go`
  - 更新 `stubTaskRepository.List` 的签名。
  - 增加断言：请求 `/v1/tasks?status=backlog` 时，stub repository 收到 `backlog` filter。
  - 保留现有错误路径测试：无效 status、缺少 owner、repository error、repository panic。
- `cmd/api/local_repository.go`
  - 如果本任务仍保留 fixture 模式，更新 `LocalRepository.List` 签名并支持 status filter。
  - 如果本任务切换到 PostgreSQL repository，确保旧 fixture 不再是默认运行路径。
- `cmd/api/local_repository_test.go`
  - 如果保留 `LocalRepository`，更新调用签名，并补 status filter 的最小测试。
- `internal/postgres/task_repository.go`
  - 新增 PostgreSQL Task repository 实现。
  - 查询 `tasks` 表，显式使用 `owner_id` 条件。
  - 可选 status filter 只能通过 SQL 参数绑定实现。
- `internal/postgres/task_repository_test.go`
  - 新增集成测试，使用隔离 schema 和 business migration。
  - 插入两个 Owner 和多条 Task fixture，覆盖 owner 隔离、空列表、status filter。
- `cmd/api/main.go`
  - 连接 PostgreSQL repository 到 HTTP API，或明确保留 fixture 模式的启动参数。
  - 不读取真实 `.env`。需要配置时使用环境变量，并让用户提供脱敏值或本地测试值。
- `openapi/openapi.yaml`
  - 本任务原则上不改 OpenAPI，因为 `status` query 已存在。
  - 只有当 HTTP 响应字段发生变化时才修改，并同步生成 TypeScript client。

## Agent 支援边界

- Agent 可以补测试骨架、fixture helper、接口接线、router 断言和文档。
- 你写 repository 核心查询：参数化 SQL、`rows.Scan`、错误处理、context 传递和 rows 关闭。
- Agent 不直接代写核心 SQL 实现，除非你明确要求我接手。

## 步骤与验证

1. 先改 task 层查询契约。
   - 修改文件：`internal/task/service.go`。
   - 修改文件：`internal/task/service_test.go`。
   - 在 `task` 包新增 `ListTasksFilter` 或同等类型，至少包含可选 `Status`。
   - 修改 `Repository.List` 签名，让它接收 `ownerID` 和 filter。
   - 修改 `Service.List` 签名，让它接收相同 filter，并原样传给 repository。
   - 更新 `service_test.go` 的 stub repository，断言 service 把 ownerID 和 filter 传下去。
   - 验证命令：`go test ./internal/task -count=1`。
   - 通过标准：task 包测试通过；如果其他包暂时因为接口签名变化编译失败，进入下一步修复，不把它当成业务失败。

2. 再改 HTTP 层，让 status filter 进入 service。
   - 修改文件：`internal/httpapi/task_handler.go`。
   - 修改文件：`internal/httpapi/router_test.go`。
   - 更新 `router_test.go` 的 `stubTaskRepository.List` 签名。
   - 在 router 测试中增加断言：请求 `/v1/tasks?status=backlog` 时，stub repository 收到 `backlog` filter。
   - 保留当前 `status` query 校验：合法值进入 service；非法值返回稳定 `400`，且不能调用 repository。
   - 保留缺少 owner、repository error、repository panic 测试。
   - 验证命令：`go test ./internal/httpapi -count=1`。
   - 通过标准：HTTP 层证明 status 被传下去，错误响应行为保持不变。

3. 更新本地 fixture repository，消除接口签名漂移。
   - 修改文件：`cmd/api/local_repository.go`。
   - 修改文件：`cmd/api/local_repository_test.go`。
   - 如果 08 仍保留 fixture 模式，给 `LocalRepository.List` 增加 filter 参数，并让 success fixture 支持 status filter 的最小行为。
   - 如果 08 决定默认切换 PostgreSQL，可以先保留 `LocalRepository` 作为测试 fixture，但不能让它继续伪装成最终运行路径。
   - 验证命令：`go test ./cmd/api -count=1`。
   - 通过标准：`cmd/api` 编译通过；fixture 测试不会掩盖 PostgreSQL repository 未实现。

4. 写 TaskRepository 查询集成测试。
   - 修改文件：新增 `internal/postgres/task_repository_test.go`。
   - 复用文件：`internal/postgres/test_helpers_test.go` 中的 `newTestDatabase` 和 `applyBusinessMigration`。
   - 这个测试不重复验证 migration runner 或 schema 约束。那些已经由 `internal/postgres/migrate_test.go`、`schema_test.go` 和 runner 测试覆盖。
   - 这个测试验证将要写的 PostgreSQL Task repository：SQL 是否带 owner scope，status filter 是否真正进入查询，`rows.Scan` 是否把数据库行正确映射成 `task.Task`。
   - 在隔离 schema 中运行 migration，只是为了给 repository 提供真实表结构。
   - 插入 Owner A、Owner B，作为查询隔离 fixture。
   - 插入至少 3 条 Task：A/backlog、A/done、B/backlog 或 B/done，作为 owner scope 和 status filter 的输入数据。
   - 调用 `TaskRepository.List`，断言 Owner A 查询永远不返回 Owner B 数据。
   - 调用 `TaskRepository.List`，断言不存在 Task 的 Owner 返回空列表，不返回 `nil` 或错误。
   - 调用 `TaskRepository.List`，断言 `status=backlog` 和 `status=done` 只返回对应状态。
   - 验证命令：`go test ./internal/postgres -run 'Test.*Task.*Repository.*List' -count=1`。
   - RED 合格标准：失败原因是 PostgreSQL Task repository 尚未实现，不是数据库未启动、migration 失败或 fixture 写错。

5. 你实现 PostgreSQL Task repository。
   - 修改文件：新增 `internal/postgres/task_repository.go`。
   - 参考 schema：`migrations/000001_initial.up.sql` 中的 `owners` 和 `tasks` 表。
   - 推荐结构：`type TaskRepository struct { conn *pgx.Conn }` 或等价结构。若后续要支持连接池，也可以使用当前项目接受的 pgx 接口抽象。
   - 推荐构造函数：`NewTaskRepository(...)`，供 `cmd/api/main.go` 使用。
   - SQL 必须包含显式 `owner_id = $...` 条件。
   - status filter 必须通过参数绑定实现，不能字符串拼接。
   - 查询必须传入请求 context。
   - `Rows` 必须关闭，并检查 `rows.Err()`。
   - `labels`、`version`、`created_at`、`updated_at` 必须映射到 `task.Task`。
   - 如果本次不映射 `project_id`、`description`、`due_at`、`archived_at`，需要在代码或后续 issue 中保持明确，不要让 OpenAPI 和实际响应继续无说明漂移。
   - 验证命令：`go test ./internal/postgres -run 'Test.*Task.*Repository.*List' -count=1`。
   - 通过标准：TaskRepository 查询集成测试通过。

6. 把 API 默认运行路径接到 PostgreSQL repository。
   - 修改文件：`cmd/api/main.go`。
   - 如有必要，新增配置读取代码，但只读取环境变量，不读取真实 `.env` 文件。
   - 启动 API 时创建 PostgreSQL 连接，并把 `internal/postgres.TaskRepository` 注入 `task.NewService`。
   - 如果保留 fixture 模式，只能作为显式开发参数存在，不能继续作为默认路径。
   - 验证命令：`go test ./cmd/api -count=1`。
   - 通过标准：`cmd/api` 编译通过；默认运行路径不再依赖本地 fixture。

7. 跑 HTTP 回归测试。
   - 修改文件：如步骤 2 已完成，此步原则上只运行测试。若测试失败，只修改 `internal/httpapi/task_handler.go` 或 `internal/httpapi/router_test.go` 中与失败直接相关的内容。
   - `GET /v1/tasks` 响应结构保持当前 Web 可用。
   - `GET /v1/tasks?status=unknown` 仍返回稳定 `400`。
   - 缺少 owner 仍返回稳定 `401`。
   - repository 错误和 panic 仍返回不泄露内部信息的 `500`。
   - 验证命令：`go test ./internal/httpapi -count=1`。

8. 跑完整 Go 测试。
   - 验证命令：`go test ./... -count=1`。
   - 通过标准：无 FAIL；若数据库未启动导致失败，明确标记为环境阻塞，不写成业务通过。

9. 做浏览器验证。
   - 可能修改文件：`web/app/task-list.tsx` 只在响应结构变化导致页面无法显示时修改。本任务不应主动改 UI。
   - 启动本地测试 PostgreSQL。
   - 运行 migration。
   - 插入与 Repository 测试等价的 Owner/Task fixture，或使用明确的开发 seed。
   - 启动 API，并确认它使用 PostgreSQL repository。
   - 启动 Web。
   - 打开 Task 页面，查看 Network 中 `/v1/tasks`。
   - 通过标准：页面显示 PostgreSQL 数据；Network 响应不包含 Owner B 数据；带 `status` 的请求与 Repository 测试结果一致。

## 最终通过

- [ ] `GET /v1/tasks` 的运行数据来自 PostgreSQL，而不是本地 fixture。
- [ ] service/repository 接口支持 status filter，handler 已向下传递。
- [ ] SQL 显式 owner scope，Owner A 查询不会返回 Owner B 数据。
- [ ] 两个 Owner 隔离测试通过。
- [ ] 空列表、status filter、无效 status、缺少 owner、repository 错误和 panic 测试通过。
- [ ] 参数化 SQL、`Rows` 关闭、`rows.Err()`、context 传递都已覆盖或人工核对。
- [ ] 没有把 SQL 写进 handler。
- [ ] `go test ./... -count=1` 已运行并通过，或明确记录环境阻塞原因。
- [ ] 浏览器 Network 验证过 PostgreSQL 数据和 owner 隔离。
