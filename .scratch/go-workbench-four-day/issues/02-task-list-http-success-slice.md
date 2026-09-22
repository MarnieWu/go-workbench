# 02 — 让 HTTP 请求经过 Gin Handler 返回 Task 列表

**最终结果：** 测试向真实 Gin Router 发送 `GET /v1/tasks` 请求。请求经过 Handler 和 Service，到达测试 Repository，最终返回 HTTP `200`，且响应正文符合 OpenAPI 定义的 `{"items":[...]}`。`X-Request-ID` Header 留到任务 04，因此任务 02 尚未完成全部响应契约。

**Blocked by:** 01。任务 01 的全部测试通过后才能开始。

**Status:** blocked

## 这次要学会什么

完成本任务后，你应该能够解释下面这条调用链，而不只是让测试变绿：

```text
测试发送 HTTP 请求
        ↓
Gin Router 根据请求方法和路径找到 Handler
        ↓
Handler 读取 HTTP 输入并调用 Service
        ↓
Service 调用测试 Repository
        ↓
Handler 把结果转换为 JSON 响应
```

本任务只实现成功响应和空列表。错误映射属于任务 03，Request ID 属于任务 04，真实登录和 Owner 身份验证属于任务 17。

## 开始前先理解这些词

### Router

Router 是“路由表”。它根据 HTTP 方法和路径决定由哪个函数处理请求：

```go
router.GET("/v1/tasks", listTasks(service))
```

这行代码表示：收到 `GET /v1/tasks` 时，调用 `listTasks` 创建的 Handler。

### Handler

Handler 是 HTTP 层的处理函数。它只负责：

1. 读取请求需要的信息。
2. 调用 Service。
3. 把结果写成 HTTP 响应。

Handler 不写 SQL，也不实现业务规则。

```go
func listTasks(service *task.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 读取 HTTP 请求需要的信息。
		// 调用 service.List(...)。
		// 使用 c.JSON(...) 返回响应。
	}
}
```

### Gin Context 和 Go Context

这两个名字相似，但职责不同：

```go
c *gin.Context       // 读取请求、写响应、读取 Gin middleware 放入的值
c.Request.Context()  // Go 标准库 context，向 Service 传递取消和超时信号
```

Handler 会同时使用两者。

### JSON

JSON 是 HTTP API 常用的数据格式。本接口的最外层对象必须包含 `items`：

```json
{
  "items": []
}
```

### `httptest`

`net/http/httptest` 是 Go 标准库提供的 HTTP 测试工具。它可以在测试中发送请求并记录响应，不需要真的监听网络端口。

### Service 和 Repository

Service 负责业务调用，Repository 负责数据访问能力。它们已在任务 01 中解释。本任务的 Handler 只调用 Service，不直接调用 Repository。

### Middleware

Middleware 是在 Handler 前执行的一层公共处理。任务 02 的测试 Middleware 只负责放入固定测试 Owner；任务 17 才实现真实身份验证。

```text
请求 → Middleware 放入测试 ownerID → Handler
```

### Owner

Owner 是工作台数据的拥有者。`ownerID` 用于确保只查询当前 Owner 的 Task。

### Slice

Slice 是 Go 中表示一组同类型数据的结构。`[]task.Task` 表示一组 Task。

### Go module

Go module 是一组一起构建和测试的 Go package。本项目的 module 名是 `go-workbench`，记录在根目录的 `go.mod` 中。

## 命令从哪里运行

本文所有命令都在仓库根目录运行：

```text
/Users/wuwenqi/Documents/ChatGPT/go-workbench
```

可以先运行下面的命令确认当前位置：

```bash
pwd
```

`pwd` 表示 print working directory，会输出当前目录。

## 本任务会涉及哪些文件

任务开始时，Agent 只创建低价值脚手架，不实现 Handler 核心逻辑：

```text
cmd/api/main.go                    API 启动空壳
internal/httpapi/router.go         创建 Gin Router 和注册路由
internal/httpapi/task_handler.go   由你完成 Handler 核心逻辑
internal/httpapi/router_test.go    真实 Router 的失败测试
```

这些文件当前尚不存在。在任务 01 完成后，先让 Agent 创建它们，再按本文继续。

## 第 0 步：解决模型与契约不一致

当前 [`internal/task/service.go`](../../../internal/task/service.go) 中的 `Task` 只有：

```go
type Task struct {
	ID      string
	OwnerID string
	Title   string
	Status  Status
}
```

但是 [`openapi/openapi.yaml`](../../../openapi/openapi.yaml) 规定成功响应还必须包含 `priority`、`labels`、`version`、`createdAt` 和 `updatedAt`。

因此，开始写 Handler 前必须先解决模型与契约不一致的问题。当前计划已经确认这些字段属于 Task，所以默认修复方向是：由你补全 Go `Task` 模型，Agent 提供字段说明和测试建议。不能在 Handler 中伪造时间、版本或其他默认值来绕过契约。

需要补齐的必填数据及建议 Go 类型是：

```go
Priority  Priority
Labels    []string
Version   int64
CreatedAt time.Time
UpdatedAt time.Time
```

`Priority` 还需要定义允许值：`none`、`low`、`medium`、`high`。`time.Time` 来自 Go 标准库 `time` 包。具体修改仍由你完成。

修改后运行：

```bash
go test ./internal/task -count=1
```

**通过标准：** Go `Task` 能表达 OpenAPI 成功响应的全部必填字段，并且 Task 包测试退出码为 0。未完成时停止，不进入 Handler 实现。

## 第 1 步：让 Agent 创建脚手架和失败测试

你对 Agent 说：

> 为任务 02 创建 Gin 依赖、`cmd/api` 启动空壳、`internal/httpapi` Router 空壳和一个真实 Router 的失败测试。不要实现 Handler 核心逻辑。测试使用假的 Repository，不连接数据库，不读取真实环境文件。

Agent 完成后，你先确认：

- `go.mod` 增加 Gin 依赖。
- 上面列出的 4 个文件已经存在。
- 测试通过 Router 发送请求，不是直接调用 Handler。
- Handler 中仍保留明确的待实现位置。

## 第 2 步：阅读 OpenAPI 契约

运行：

```bash
sed -n '1,120p' openapi/openapi.yaml
```

### 这条命令是什么意思

- `sed`：读取并处理文本的命令。这里仅用它查看文件，不会修改文件。
- `-n`：默认不输出全部内容。
- `1,120p`：只打印第 1 至 120 行。
- `openapi/openapi.yaml`：要查看的文件。

如果暂时不想学习 `sed`，也可以直接在编辑器中打开 [`openapi/openapi.yaml`](../../../openapi/openapi.yaml)。两种方式的目的相同。

### 为什么要先看契约

Handler 返回的状态码、字段名和字段类型不能由我们临时决定。它们必须符合前后端共同使用的 OpenAPI 契约。

这一步需要找到：

```text
GET /v1/tasks
可选查询参数：status
成功状态码：200
响应外形：{"items":[Task, ...]}
```

当前任务只实现“不带 `status`”的成功列表。任务 03 校验非法 `status`，任务 08 才把合法 `status` 传入 PostgreSQL 查询并真正过滤。任务 02 完成时，过滤能力仍应明确标为未实现。

`X-Request-ID` 也在契约中，但它由任务 04 实现，本任务不提前实现。因此本任务只验证响应正文，不声称完整响应契约已经实现。

**通过标准：** 你能用自己的话说明这个接口接收什么请求、成功时返回什么。

## 第 3 步：运行目标失败测试

运行 Agent 创建的确切测试名。如果测试名为 `TestRouterListTasks`，命令是：

```bash
go test ./... -run '^TestRouterListTasks$' -count=1
```

### 这条命令是什么意思

- `go test`：运行 Go 测试。
- `./...`：检查当前 Go module 下的全部包。
- `-run`：只运行测试名与后面规则匹配的测试。
- `^TestRouterListTasks$`：只匹配名字完全等于 `TestRouterListTasks` 的测试。
- `-count=1`：不使用之前的测试缓存，本次重新执行。

`^` 表示字符串开头，`$` 表示字符串结尾。这种匹配规则叫正则表达式。当前只需要认识这两个符号，不需要先系统学习全部正则语法。

### 什么是正确的 RED

测试应该成功编译，然后因为路由或 Handler 尚未完成而失败，例如：

```text
status = 404, want 200
```

如果错误是缺少依赖、语法错误或找不到文件，这不是预期 RED，应先修复测试环境。

**通过标准：** 失败原因明确指向尚未实现的 HTTP 行为。

## 第 4 步：读懂失败测试

打开：

```text
internal/httpapi/router_test.go
```

测试中会看到类似结构：

```go
recorder := httptest.NewRecorder()
request := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)

router.ServeHTTP(recorder, request)
```

- `NewRecorder()` 创建一个响应记录器，用来保存状态码、Header 和响应正文。
- `NewRequest(...)` 创建一个假的 `GET /v1/tasks` 请求。
- `ServeHTTP(...)` 把请求交给真实 Gin Router 处理。

测试不能直接调用 `listTasks(...)`，否则无法证明路由注册正确。

**通过标准：** 你能指出测试中“创建请求”“交给 Router”“读取响应”分别是哪一行。

## 第 5 步：实现成功路径 Handler

打开：

```text
internal/httpapi/task_handler.go
```

本步骤会修改三个文件，代码不要混放：

```text
internal/httpapi/task_handler.go   ownerIDKey、响应类型、listTasks
internal/httpapi/router.go         NewRouter 和路由注册
internal/httpapi/router_test.go    testOwner 和测试 Repository
```

按下面顺序完成，每完成一项再写下一项：

1. 从 `*gin.Context` 读取测试 middleware 放入的 `ownerID`。
2. 取出 `c.Request.Context()`。
3. 调用 `service.List(ctx, ownerID)`。
4. 立即检查 `err`；出错时返回并停止，不能继续写 `200`。
5. 把 `[]task.Task` 映射为 OpenAPI 规定的 JSON 字段。
6. 使用 `c.JSON(http.StatusOK, response)` 返回结果。

核心控制流程如下，字段映射由你完成：

```go
func listTasks(service *task.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ownerID := c.GetString(ownerIDKey)
		tasks, err := service.List(c.Request.Context(), ownerID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		var items []taskResponse
		for _, currentTask := range tasks {
			// TODO: 把 currentTask 的契约字段追加到 items。
		}

		c.JSON(http.StatusOK, listTasksResponse{Items: items})
	}
}
```

### 为什么任务 02 也必须检查 `err`

`service.List` 同时返回任务列表和错误：

```go
tasks, err := service.List(...)
```

如果 `err != nil`，`tasks` 不能当作成功结果继续返回。任务 02 只做最小的安全处理：返回 HTTP `500` 并立即结束 Handler。任务 03 再补稳定错误 code、JSON 格式、日志和不泄露内部错误的测试。

不要写成：

```go
tasks, _ := service.List(...)
```

这会直接丢弃错误。

### Owner 从哪里来

浏览器不能直接声明自己是谁。任务 02 的测试 middleware 会放入固定测试 Owner，使我们只学习成功链路。任务 17 才会用验证后的登录令牌产生真实 Owner。

先在 `internal/httpapi/task_handler.go` 定义生产代码和测试共同使用的私有 key：

```go
const ownerIDKey = "ownerID"
```

再在 `internal/httpapi/router_test.go` 定义只供测试使用的 middleware：

```go
func testOwner(ownerID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ownerIDKey, ownerID)
		c.Next()
	}
}
```

不要把 `testOwner` 放进 `task_handler.go`，否则测试辅助代码会被编译进生产程序。

然后在 `internal/httpapi/router.go` 修改 `NewRouter`。为了让测试 middleware 在 Handler 前执行，Router 需要先注册 middleware，再注册路由：

```go
func NewRouter(service *task.Service, middlewares ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(middlewares...)
	router.GET("/v1/tasks", listTasks(service))
	return router
}
```

`middlewares ...gin.HandlerFunc` 表示可以传入零个或多个 middleware。测试传入 `testOwner("owner-1")`；任务 17 再替换为真实鉴权 middleware。

`c.Set` 把测试 Owner 放入本次 Gin 请求，`c.GetString` 在 Handler 中读取它。当前代码中的硬编码 `ownerID := "123"` 必须移除，否则无法证明请求身份被正确传递。

### 为什么 Handler 不写 SQL

Handler 只处理 HTTP。数据读取已经由下面这条链路负责：

```text
Handler → Service.List → Repository.List
```

如果 Handler 直接写 SQL，Service 和 Repository 的边界就失去作用。

### 为什么需要响应类型

Go 领域模型和 HTTP JSON 不是同一个概念。建议定义明确的响应结构并添加 JSON tag，例如：

```go
type listTasksResponse struct {
	Items []taskResponse `json:"items"`
}
```

`json:"items"` 告诉 Go：输出 JSON 时字段名必须是小写 `items`。

不要直接使用 `Items []task.Task`。否则会暴露 HTTP 契约没有声明的 `OwnerID`，字段名也可能不符合 OpenAPI。`taskResponse` 只包含契约允许返回的字段。

**本步不要实现：** 错误 JSON、panic recovery、Request ID、真实鉴权和 SQL。

## 第 6 步：让目标测试转绿

重新运行第 3 步的命令：

```bash
go test ./... -run '^TestRouterListTasks$' -count=1
```

测试至少需要断言：

- HTTP 状态码是 `200`。
- `Content-Type` 是 JSON。
- 响应存在 `items`。
- Task 字段名和数据符合 OpenAPI 契约。
- Service 最终调用了测试 Repository。测试 Repository 应记录调用次数和收到的 `ownerID`，测试再断言这些值，不能只根据响应内容推测调用发生过。

**通过标准：** 命令退出码为 0，输出中没有 `FAIL`、panic 或编译错误。

## 第 7 步：增加空列表测试

打开 `internal/httpapi/router_test.go`，先把测试 Repository 改为可以配置返回值：

```go
type stubRepository struct {
	tasks      []task.Task
	err        error
	callCount  int
	gotOwnerID string
}

func (r *stubRepository) List(_ context.Context, ownerID string) ([]task.Task, error) {
	r.callCount++
	r.gotOwnerID = ownerID
	return r.tasks, r.err
}
```

这里使用指针接收者 `*stubRepository`，因为测试需要在调用后读取更新后的 `callCount` 和 `gotOwnerID`。第 6 步和第 7 步都使用这一个测试替身，不要再定义第二个同名类型。

然后新增完整测试。这里故意让 Repository 返回 `nil, nil`，因为 Go 会把 `nil` slice 默认编码为 `null`：

```go
func TestRouterListTasksReturnsEmptyItems(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repository := &stubRepository{tasks: nil}
	service := task.NewService(repository)
	router := NewRouter(service, testOwner("owner-1"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Items == nil {
		t.Fatal("items = null or missing, want []")
	}
	if len(response.Items) != 0 {
		t.Fatalf("len(items) = %d, want 0", len(response.Items))
	}
	if repository.callCount != 1 {
		t.Fatalf("repository calls = %d, want 1", repository.callCount)
	}
	if repository.gotOwnerID != "owner-1" {
		t.Fatalf("ownerID = %q, want %q", repository.gotOwnerID, "owner-1")
	}
}
```

需要在测试文件中增加标准库导入：

```go
import "encoding/json"
```

### 这个测试怎样判断“空”

- `response.Items == nil`：JSON 是 `null`，或者缺少 `items`，测试失败。
- `len(response.Items) != 0`：数组中仍有元素，测试失败。
- `response.Items != nil` 且长度为 `0`：JSON 是空数组 `[]`，测试通过。

Handler 中的下面一行会创建“非 nil、长度为 0”的结果；即使 `tasks` 是 `nil`，最终也会编码为 `[]`：

```go
items := make([]taskResponse, 0, len(tasks))
```

先运行新增测试，确认它因为 `items = null or missing, want []` 而 RED；然后把第 5 步中的 `var items []taskResponse` 替换为上面的 `make` 写法，再重跑测试。这样可以证明测试确实捕获了 `null` 问题。

接口必须返回：

```json
{"items":[]}
```

不能返回：

```json
{"items":null}
```

这是因为前端可以直接遍历数组；`null` 会迫使前端增加额外判断。

运行：

```bash
go test ./... -run '^TestRouterListTasksReturnsEmptyItems$' -count=1
```

**通过标准：** 响应正文中的 `items` 是空数组 `[]`。

## 第 8 步：格式化并检查全部 Go 代码

先格式化本任务涉及的 Go 文件：

```bash
gofmt -w internal/task/*.go internal/httpapi/*.go cmd/api/*.go
```

`gofmt` 是 Go 官方格式化工具；`-w` 表示把格式化结果写回文件。三个路径限定本任务涉及的目录，`*.go` 表示匹配目录中的全部 `.go` 文件。

再运行：

```bash
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
```

- 第一条运行全部测试。
- 第二条额外检查并发数据竞争。
- 第三条执行 Go 静态检查；成功时通常没有输出。

**通过标准：** 三条命令退出码都为 0；输出中没有 `FAIL`、panic、编译错误或 vet 错误。

## 第 9 步：检查修改范围

运行：

```bash
git diff -- go.mod go.sum cmd/api internal/httpapi internal/task
```

`git diff` 查看修改内容；独立的 `--` 表示后面的内容都是文件或目录路径，不再解析为命令选项。

检查：

- Handler 没有 SQL。
- 没有实现任务 03、04 或 17 的内容。
- 没有读取真实 `.env`、Token 或凭据。
- 没有为了当前单一路由创建多余抽象。

仓库当前还没有首个提交，新的未跟踪文件可能不会出现在普通 `git diff` 中，因此还要运行：

```bash
git status --short
```

`--short` 使用两列简短状态码显示文件状态，例如 `??` 表示未跟踪文件，`M` 表示已修改文件。

## 最终通过清单

- [ ] 任务 01 已完成。
- [ ] Go `Task` 与 OpenAPI 必填字段已经对齐。
- [ ] 你能解释 Router、Handler、Service 和 Repository 的职责。
- [ ] 目标测试先因功能缺失而 RED，之后变为 GREEN。
- [ ] 测试通过真实 Router 发送 HTTP 请求。
- [ ] 不带 `status` 的 `GET /v1/tasks` 返回契约规定的 `200` JSON 正文。
- [ ] `status` 过滤和 `X-Request-ID` 明确标为后续任务，未被误报为已完成。
- [ ] 空列表返回 `{"items":[]}`，不是 `null`。
- [ ] Handler 没有 SQL、真实鉴权或 Request ID 逻辑。
- [ ] 全部 Go 测试、race test 和 vet 通过。
- [ ] diff 只包含本任务需要的修改。

## 完成后你应该能回答

1. Router 和 Handler 分别负责什么？
2. 为什么 Router 测试比直接调用 Handler 更完整？
3. `*gin.Context` 和 `context.Context` 有什么区别？
4. 为什么 Handler 应调用 Service，而不是直接写 SQL？
5. 为什么空列表应该编码为 `[]` 而不是 `null`？
6. 为什么 Go 领域模型不能不经检查就直接作为 API JSON 返回？

完成后，把目标测试、完整测试、race test、vet、`git diff` 和 `git status --short` 的完整输出发给 Agent。Agent 会先审查证据，再进入任务 03。
