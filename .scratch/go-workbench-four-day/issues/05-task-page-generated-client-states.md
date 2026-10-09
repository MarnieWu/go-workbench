# 05 — 用 OpenAPI 生成类型显示 Task 列表

## 这一步要做什么

把前面已经完成的 Go Task 查询接口接到首页，形成第一条可以在浏览器中实际运行的只读链路：

```text
首页 Task 列表
  → OpenAPI 生成客户端
  → GET /v1/tasks
  → Gin Handler
  → Task Service
  → 本地内存 Repository
```

具体要完成四件事：

1. 建立本地运行链路，让浏览器能够通过 Next.js 转发访问 Gin API。
2. 对齐 Go Handler、OpenAPI 文档和生成的 TypeScript 类型，确保三者描述的是同一份接口契约。
3. 在首页显示 Task 列表，并处理 loading、empty、success、HTTP error 和 network error 五种状态。
4. 用真实可复现的本地 fixture 和构建命令验证链路，不在页面中增加只为演示状态而存在的 mock 分支。

## 为什么要做

这一步的重点不是“画一个列表页面”，而是验证前后端能否依赖同一份 OpenAPI 契约协作。完成后，接口字段或响应状态发生变化时，重新生成的 TypeScript 类型应在编译阶段暴露不一致，避免前端长期维护一套手写且容易漂移的 API 类型。

同时，这一步把第一天前面分散完成的 Handler、错误响应和 request ID 串成一条真实请求链路。只有真实复现成功、空数据、服务端错误和网络错误，才能证明页面处理的是运行时结果，而不只是 TypeScript 编译通过。

## 完成结果

首页的 Task 列表区域通过 `web/lib/api/client.ts` 导出的 `openapi-fetch` 客户端调用 `GET /v1/tasks`，并根据请求结果显示 loading、empty、success、HTTP error/retry 或 network error/retry。页面使用 `web/lib/api/generated/schema.d.ts` 中的生成类型，不手写第二套 Task API 类型。

## 范围边界

- 本任务只完成 Task 列表的读取和展示，不实现创建、编辑、删除、筛选或分页。
- 本任务使用本地内存 Repository 复现状态，不接 PostgreSQL，也不写入数据。
- 本任务使用固定的本地 `ownerID` 打通开发链路，不把它当作认证；真实身份验证留到 OIDC 任务。
- 本任务不配置 CORS。浏览器请求同源 `/api`，由 Next.js 转发到 Gin。
- Go 接口契约和生成类型由你核对；Agent 可以完成 UI、交互、样式和必要的连接代码。

## 完成后你应该能解释

- 为什么前端要复用 OpenAPI 生成类型，而不是再手写一套 Task 类型。
- 浏览器请求如何经过 Next.js、Gin、Handler、Service，最后到达 Repository。
- HTTP error 与 network error 的区别，以及为什么后者没有服务端 request ID。
- 为什么 TypeScript 编译通过不能替代 Go router 测试和真实浏览器联调。

**Blocked by:** 03、04。本任务的第 0 步定义并实现本地运行时链路；只有 03、04 和第 0 步同时验收通过，才能开始 Task 页面的真实联调。

**Status:** blocked。只有在 03、04 和第 0 步通过验收后，才可改为 `ready-for-agent`。

## 开始本任务时的基线

- `cmd/api/main.go` 当前不启动 HTTP server，因此现在无法从浏览器发起真实的 Task 请求。
- `web/lib/api/client.ts` 当前默认请求 `http://localhost:8080`，但仓库未定义浏览器到 Go API 的转发链路。
- `GET /v1/tasks` 的 Go Handler 可返回 `200`、`400`、`401` 和 `500`。在每次修改契约后，必须重新生成 TypeScript 类型，并核对生成结果仍包含四种响应。
- 本任务不通过页面 URL 参数、前端环境变量或页面内的 mock 开关伪造状态。状态复现仅使用第 0 步定义的本地 Go fixture。

## 第 0 步 — 本地运行时链路

**已选择的开发方案：** 浏览器只请求同源的 `/api` 路径，Next.js 将请求转发到本地 Gin 服务。本阶段不配置 CORS、OIDC、数据库或生产部署。

```text
Browser: http://localhost:3000/api/v1/tasks
  → Next.js rewrite
  → Gin: http://127.0.0.1:8080/v1/tasks
```

1. `GET /v1/tasks` 的 Handler 在执行前必须从 Gin Context 读取 `ownerID`；没有该值时会返回 `401`。本地阶段没有登录系统，因此在 `internal/httpapi/router.go` 将现有的 `LocalOwnerMiddleware` 改为接收一个固定 ID，并在进入 Handler 前写入 Context：
   ```go
   func LocalOwnerMiddleware(ownerID string) gin.HandlerFunc {
       return func(c *gin.Context) {
           c.Set(ownerIDKey, ownerID)
           c.Next()
       }
   }
   ```
   它的目的只是让本地 `GET /v1/tasks` 能以固定用户 `owner-local` 查询数据；它不是认证功能。不得从客户端 Header、Query 或 Cookie 读取此 ID，否则任何调用方都能冒充其他用户。接入 OIDC 后，由验证通过的身份声明写入 `ownerID`，并且不注册此固定 Owner middleware。
2. 在 `cmd/api/local_repository.go` 写一个**给本地 API 提供固定测试数据的组件**。现在还没有接数据库，但 Service 查询任务时仍需要一个地方回答“这个用户有哪些任务”。这个文件负责提供这个回答：Service 调用它的 `List(ctx, ownerID)` 方法，它返回任务列表或错误。这里的 Repository 就是“任务数据来源”；`local` 表示这个实现只用于本地开发。

   例如，浏览器打开首页后，请求经过 Handler 和 Service，最终来到这个组件。启动 API 时选择的 `-fixture` 决定它如何回答。`fixture` 在这里表示预先设置好的测试场景：

   | 启动时选择 | `List` 返回什么 | 浏览器应该看到什么 |
   | --- | --- | --- |
   | `-fixture=success` | 一条固定的示例任务，例如“学习 Go 接口”，以及 `nil` error | 正常的任务列表 |
   | `-fixture=empty` | 非 `nil` 的空任务切片，以及 `nil` error | “暂无任务” |
   | `-fixture=error` | 一个内部 error | Handler 返回 `500 INTERNAL_ERROR`，页面显示错误和重试按钮 |

   这样你可以切换启动参数，稳定检查页面在三种结果下的表现，不必先安装数据库、插入数据或故意破坏数据库。请求仍然经过真实的 Go Handler 和 Service；这个组件只控制数据查询的结果。它不返回 HTTP 状态码或 JSON，这些由 Handler 负责。

   **具体要写的是一个结构体和它的 `List` 方法。** `internal/task/service.go` 已经定义了 `task.Repository` 接口，不需要再定义 `LocalRepository interface`。接口只规定方法签名；本步骤要求你写出查询时实际执行的代码。按下面顺序完成：

   **2.1 设置文件所属 package 和导入。** `cmd/api/local_repository.go` 与 `main.go` 同属 `package main`。导入 `context` 和 `go-workbench/internal/task`；按实现需要导入 `errors`、`time`。任务类型来自 `task` package，因此此文件的方法返回值要写 `[]task.Task`。

   **2.2 定义结构体。** 在这个文件中写：
   ```go
   type LocalRepository struct {
       fixture string
   }
   ```
   `fixture` 保存启动时选择的场景。例如 `LocalRepository{fixture: "success"}` 表示这个实例查询时返回示例任务。它不是任务列表，也不是数据库地址。

   **2.3 给结构体定义方法。** 方法签名必须是：
   ```go
   func (r LocalRepository) List(ctx context.Context, ownerID string) ([]task.Task, error)
   ```
   你需要在签名后加方法体，并填写下面三个分支。`r` 是接收者，表示这次调用使用哪个 Repository 实例；方法体通过 `r.fixture` 读取它的场景。`ctx` 是请求上下文，`ownerID` 是 Service 传入的查询用户 ID。

   **2.4 在方法体中根据 `r.fixture` 返回结果。** 用 `switch r.fixture` 选择分支：
   - `success`：创建一条属于固定用户 `owner-local` 的 `task.Task`，填写 `ID`、`OwnerID`、`Title`、`Status`、`Priority`、`Labels`、`Version`、`CreatedAt` 和 `UpdatedAt`。可使用 `task.StatusBacklog`、`task.PriorityMedium`，版本设为 `1`，时间使用固定的有效时间。只有 `ownerID == "owner-local"` 时返回这条任务；其他用户返回空切片，避免把示例任务返回给任意用户。成功返回值是任务切片和 `nil` error。
   - `empty`：返回 `[]task.Task{}` 和 `nil`。空列表表示查询成功但没有数据。
   - `error`：返回 `nil` 和一个内部 error，例如用 `errors.New("local fixture query failed")` 创建错误。Handler 会把该错误转换为 HTTP 错误响应。
   - 默认分支：返回说明未知 fixture 的 error，避免未知场景被当作查询成功。第 3 步还必须在启动前校验参数。

   **2.5 检查结构体是否满足已有接口。** 在文件中添加编译期检查：
   ```go
   var _ task.Repository = LocalRepository{}
   ```
   Go 根据方法签名自动判断接口实现关系，不需要写 `implements`。如果接收者、参数或返回值不符合接口要求，这行会导致编译失败。

   **2.6 把实例交给 Service。** 在下一步的 `main.go` 中，解析启动参数后按以下关系连接：
   ```go
   repository := LocalRepository{fixture: *fixture}
   service := task.NewService(repository)
   ```
   这里假设 `fixture := flag.String("fixture", "success", "local test scenario")`，并已调用 `flag.Parse()`、校验参数。调用链是 `service.List(ctx, ownerID)` → `repository.List(ctx, ownerID)` → 返回所选场景的任务或错误。

   **这一步的验收：** 第 4 步的测试直接创建三个不同 `fixture` 的 `LocalRepository` 实例并调用 `List`，分别核对示例任务、非 `nil` 空切片和 error；另检查其他用户读不到 `owner-local` 的示例任务。三个场景都不写入数据、不连接数据库、不使用真实用户信息。后续接入 PostgreSQL 时，把传给 Service 的这个数据来源换成数据库实现，Service 仍通过同一个 `List` 方法查询任务。
3. 在 `cmd/api/main.go` 使用 Go 标准库 `flag` 解析 `-fixture=success|empty|error`，用该 Repository 创建 `task.Service`，并使用 `httpapi.NewRouter(service, httpapi.LocalOwnerMiddleware("owner-local"))` 创建 Router。在 `127.0.0.1:8080` 上启动服务。未知 fixture 名称必须在启动前报错并退出。
4. 在 `cmd/api/local_repository_test.go` 为三个 fixture 各写一个测试：`success` 返回一个 `owner-local` Task；`empty` 返回非 `nil` 空切片；`error` 返回 error。运行：
   ```bash
   go test ./cmd/api -run '^TestLocalRepositoryList$' -count=1
   ```
   - 通过：命令退出码为 0。
5. 在 `web/next.config.ts` 定义 rewrite：`/api/:path*` 转发到 `http://127.0.0.1:8080/:path*`。不增加 CORS 中间件。
6. 将 `web/lib/api/client.ts` 的默认 `baseUrl` 修改为 `"/api"`。当并且仅当部署环境已经提供可供浏览器直连的 API 时，才通过 `NEXT_PUBLIC_API_BASE_URL` 覆盖它。
7. 启动前运行 `go run ./cmd/api -fixture=unknown`。
   - 通过：命令以非 0 退出，并说明 fixture 不可用；它不启动监听端口。
8. 启动并验证后端链路：在终端 A 运行下列长进程；它会持续占用终端。等待它开始监听后，在终端 B 运行 `curl`。
   ```bash
   # 终端 A
   go run ./cmd/api -fixture=success

   # 终端 B
   curl -i http://127.0.0.1:8080/v1/tasks
   ```
   - 通过：返回 `200`、`Content-Type: application/json`、非空 `items` 以及非空 `X-Request-ID`。
9. 在终端 C 启动前端，保持终端 A 的 Gin 进程继续运行，再验证 rewrite：
   ```bash
   make run-frontend
   ```
   - 在浏览器打开 `http://localhost:3000`。Network 中的请求 URL 必须是 `http://localhost:3000/api/v1/tasks`，不得直接请求 `:8080`。
   - 通过：页面获得 `200` 和非空 Task 列表；浏览器控制台中没有 CORS 或 `ERR_CONNECTION_REFUSED` 错误。

## 你负责什么

你负责核对 Go Handler 的 `200`、`400`、`401`、`500` 响应与 OpenAPI 定义是否一致，并确认生成类型包含这些响应。Agent 可以修改首页 Task 列表区域的 UI、交互和样式。Agent 必须复用现有 `openapi-fetch` 客户端，不得新建第二套 API 客户端或 Task 响应类型。

## 步骤与验证

1. 对照 `internal/httpapi/task_handler.go` 和 `openapi/openapi.yaml`，补齐 `GET /v1/tasks` 的 `400` 和 `500` 响应；两者都引用 `ErrorResponse`。然后运行 `npm --prefix web run generate:api`。
   - 核对：命令退出码为 0。
   - 通过：生成文件中的 `listTasks` 同时包含 `200`、`400`、`401`、`500` 响应类型。
2. 运行 `git diff -- web/lib/api/generated`。
   - 核对：只允许出现由本次 OpenAPI `400` 和 `500` 声明产生的类型变化；其他 diff 必须逐项解释来源。
   - 通过：没有无法解释的漂移。
3. 在 `web/app/page.tsx` 现有的 Task 列表区域中挂载 `web/app/task-list.tsx` 客户端组件。由该组件通过 `api.GET("/v1/tasks")` 发起请求，并按以下条件渲染状态：
   - loading：请求开始后、收到响应前，显示骨架屏和可读的“正在加载任务”文本；Task 列表容器设置 `aria-busy="true"`。
   - success：接口返回 `200` 且 `items` 非空时，每个列表项以 `title` 为主信息，并显示 `status` 和 `priority`。
   - empty：接口返回 `200` 且 `items` 为空数组时，显示“暂无任务；当前页面仅支持查看任务”；不显示无法使用的创建按钮。
   - HTTP error：服务端返回错误响应时，显示安全的错误说明和重试按钮；仅当响应中的 `requestId` 非空时才显示该 ID。
   - network error：断网、API 未启动或请求未取得 HTTP 响应时，显示网络错误说明和重试按钮，并明确显示“无 request ID”；不生成或伪造 ID。
4. 使用第 0 步的 fixture 逐项复现上述状态。每次更换 fixture 时先停止当前 `go run` 进程，再启动下一个模式。
   - success：运行 `go run ./cmd/api -fixture=success`。
   - empty：运行 `go run ./cmd/api -fixture=empty`，页面显示 `{"items":[]}` 对应的空状态。
   - loading：使用浏览器 Network throttling 延长已发出请求的响应时间。
   - HTTP error：运行 `go run ./cmd/api -fixture=error`，取得真实的 `500` 错误 JSON 和 request ID。
   - network error：停止本地 API 后点击重试。
   - 通过：五种状态都能按上述方法重复出现，且页面中没有为演示而增加的 mock 分支。
5. 在 360px、768px、1024px viewport 检查页面。
   - 核对：页面无水平滚动；Task 主信息和重试按钮可见；键盘可到达重试按钮，且按钮具有可见焦点。
   - 核对：loading 有可读文本，error 使用可被辅助技术及时感知的 alert 语义，状态切换后焦点不会留在已移除的元素上。
   - 通过：三个宽度都没有阻断操作或内容遮挡。
6. 运行：
   ```bash
   npm --prefix web run lint
   npm --prefix web run typecheck
   npm --prefix web run build
   ```
   - 核对：三条命令退出码都为 0。
   - 通过：没有 error；warning 必须记录并判断是否影响本任务。

## 最终通过

- [ ] OpenAPI 与 Go Handler 的 `200`、`400`、`401`、`500` 响应一致，生成类型包含四种响应。
- [ ] 浏览器请求 `http://localhost:3000/api/v1/tasks`，Next.js 成功转发到 `http://127.0.0.1:8080/v1/tasks`，且不依赖 CORS。
- [ ] `success`、`empty` 和 `error` fixture 分别稳定产生非空列表、空列表和 `500 INTERNAL_ERROR`；它们均不写入数据。
- [ ] 首页 Task 列表只通过现有 `openapi-fetch` 客户端发起请求，没有第二套 API 客户端或手写 Task 响应类型。
- [ ] loading、empty、success、HTTP error/retry、network error/retry 都能按文档方法重复出现。
- [ ] empty 准确说明当前无数据且未开放创建能力，不显示无效操作。
- [ ] HTTP error 仅显示服务端实际返回的 request ID；network error 显示“无 request ID”。
- [ ] 360px、768px、1024px 三个 viewport 没有阻断操作、内容遮挡或水平滚动。
- [ ] loading、empty 和 error 对辅助技术可感知，键盘可操作重试按钮。
- [ ] lint、typecheck、build 全部通过。

完成后解释：TypeScript 编译通过为什么不能替代 Go router 测试。
