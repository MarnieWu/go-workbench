# 03 — 为 Task 列表建立最小 HTTP 错误边界

**最终结果：** `GET /v1/tasks` 对非法 `status`、缺失 Owner、Repository error 和 Repository panic 返回确定的 HTTP 状态码与安全 JSON。响应不得包含数据库错误、panic 内容、stack、文件路径或其他内部实现信息。

**Blocked by:** 02。任务 02 的成功列表、非空字段映射、空数组行为和 Repository 调用断言通过后再开始。

**Status:** ready-for-user

## 本任务要学会什么

本任务只学习一条错误处理链：

```text
HTTP 输入
    → Handler 校验
    → Service / Repository
    → 将可预期错误映射为安全响应
    → Recovery Middleware 捕获不可预期 panic
```

测试是错误边界的护栏，不是本任务的主体。先写一个最小失败测试，再实现对应行为；不要求一开始使用 table-driven tests。少量重复比过早抽象更适合本阶段。

## 必须实现的四个场景

| 场景 | HTTP 状态码 | 稳定 code | Repository 调用次数 | 说明 |
| --- | ---: | --- | ---: | --- |
| `status` 不在允许值中 | `400` | `INVALID_STATUS` | `0` | 只校验参数；合法 `status` 暂不用于过滤 |
| 缺失或为空的 Owner | `401` | `UNAUTHORIZED` | `0` | 当前只检查 Middleware 是否提供 Owner，不实现真实认证 |
| Repository 返回 error | `500` | `INTERNAL_ERROR` | `1` | 不向客户端返回 `err.Error()` |
| Repository 发生 panic | `500` | `INTERNAL_ERROR` | `1` | 由生产 Router 注册的 Recovery Middleware 捕获 |

任务 03 的阶段性错误响应为：

```json
{
  "code": "INTERNAL_ERROR",
  "message": "internal server error"
}
```

`code` 是客户端判断错误类型的稳定字段。`message` 使用固定、安全的文本，不拼接内部 error 或 panic。

OpenAPI 当前要求 `ErrorResponse` 包含 `requestId`，但 Request ID 属于任务 04。任务 03 因此只完成 `code`、`message` 和状态码，不能声称错误响应已经完全符合最终 OpenAPI 契约。任务 04 必须补齐 `requestId`，并验证它与响应 Header 和日志一致。

## 明确不做什么

- 不实现 Request ID；属于任务 04。
- 不实现合法 `status` 的数据库过滤；属于任务 08。
- 不实现 OIDC、Token 或真实身份验证；属于任务 17。
- 不写 SQL，不新增 Repository 实现。
- 不新增错误处理框架、接口或第三方依赖。
- 不为了展示 `errors.Is` 人为增加当前业务不需要的 sentinel error。
- 不记录 Authorization、Cookie、完整请求正文或 panic stack。结构化日志在任务 04 结合 Request ID 处理。

## 推荐实施顺序

### 第 0 步：补齐任务 02 的测试缺口

任务 02 的非空列表测试至少确认：

- 状态码为 `200`。
- Content-Type 是 JSON。
- 响应中的 Task 必填字段和值符合 OpenAPI。
- Repository 只调用一次，并收到正确的 `ownerID`。

空列表测试继续确认 `items` 是 `[]`，不是 `null`。

验证：

```bash
go test ./... -run '^TestRouterListTasks' -count=1
```

### 第 1 步：实现 Repository error

先写一个普通测试，让测试 Repository 返回带有明显内部信息的 error，例如 `database connection detail`。

测试只证明：

- 状态码为 `500`。
- `code` 为 `INTERNAL_ERROR`。
- 响应正文不包含内部 error 文本。
- Repository 调用一次。

然后在 HTTP 包中定义一个最小错误响应类型和一个集中写响应的函数。Handler 遇到 Repository error 时调用该函数并立即 `return`。不要在多个分支复制 JSON 结构。

### 第 2 步：实现非法 status

先写测试请求：

```text
GET /v1/tasks?status=unknown
```

Handler 只接受 OpenAPI 已定义的值：

```text
backlog
in_progress
blocked
done
```

非法值返回 `400` 和 `INVALID_STATUS`，并且不得调用 Repository。没有 `status` 时继续执行原有成功列表流程。合法值在本任务中只通过校验，不传给 Repository，也不执行过滤。

### 第 3 步：实现缺失 Owner

测试 Router 时不要注册 `testOwner` Middleware，以真实表示 Owner 缺失。不要使用客户端 query 或 header 提供 Owner。

Handler 在调用 Service 前检查 Owner。缺失或空值返回 `401` 和 `UNAUTHORIZED`，并且不得调用 Repository。

### 第 4 步：实现 panic 边界

给现有测试 Repository 增加一个最小 `panicOnList` 开关。panic 测试应通过真实 Router 调用接口，不直接调用 Handler。

在 Recovery Middleware 尚未实现时，测试会直接因 panic 失败，后续 response 断言不会执行。这是正确的 RED。不要在测试内部使用 `recover()` 或临时注册测试专用 Recovery，否则无法证明生产 Router 具有错误边界。

实现一个最小 Recovery Middleware，并在 `NewRouter` 中注册。Middleware 捕获 panic 后返回 `500` 和 `INTERNAL_ERROR`，不得把 panic 值写入响应。当前任务不输出 panic stack 或请求敏感信息。

### 第 5 步：验证没有破坏成功路径

依次运行：

```bash
go test ./... -run '^TestRouterListTasks' -count=1
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
```

通过标准：三次测试均显示 `ok`，`go vet` 退出码为 `0` 且没有错误。

## 最小测试要求

可以先写四个独立测试。只有当重复代码已经影响阅读时，才把行为相同的场景整理为 table-driven tests。panic 路径与普通 error 路径不同，可以继续保留独立测试。

每个测试只验证与该场景有关的行为：

- 非法 `status`：`400`、`INVALID_STATUS`、Repository 未调用。
- 缺失 Owner：`401`、`UNAUTHORIZED`、Repository 未调用。
- Repository error：`500`、`INTERNAL_ERROR`、内部 error 未泄露。
- Repository panic：Recovery 后返回 `500`、`INTERNAL_ERROR`、panic 内容未泄露。

不要为了提高覆盖率测试 Gin 或 Go 标准库本身，也不要在本任务锁定与业务无关的实现细节。

## 最终通过清单

- [ ] 任务 02 的非空响应和空数组测试满足其验收标准。
- [ ] 四个错误场景各有一个能够先 RED、后 GREEN 的行为测试。
- [ ] `401`、`400` 和 `500` 没有混用。
- [ ] 非法输入和缺失 Owner 不会调用 Repository。
- [ ] Repository error 和 panic 都返回稳定、安全的错误 JSON。
- [ ] Handler 写入错误响应后立即结束，不再写入 `200`。
- [ ] panic 由生产 Router 的 Recovery Middleware 捕获。
- [ ] 响应不包含数据库错误、panic、stack、文件路径或内部错误文本。
- [ ] 没有提前实现 Request ID、真实认证、SQL 或 status 数据库过滤。
- [ ] 全部测试、race test 和 vet 通过。

## 完成后你应该能回答

1. Go 的 `error` 与 `panic` 有什么区别，为什么处理位置不同？
2. 为什么非法参数和缺失 Owner 必须在调用 Repository 前返回？
3. 为什么客户端应依赖稳定 `code`，而不是匹配 `message` 或 `err.Error()`？
4. 为什么 Recovery 应注册在生产 Router，而不是只写在测试中？
5. 为什么任务 03 还不能声称完全满足最终 OpenAPI 错误响应契约？

完成后，把四个目标测试、全部测试、race test、vet、`git diff` 和 `git status --short` 的完整输出发给 Agent。Agent 先审查行为与边界，再整理本任务的学习问答。
