# 06 — 第一天总验收

**结果：** 第一条只读链路可运行，你能解释请求从浏览器到 Repository 的每一层。

**Blocked by:** 03、04、05。

**Status:** ready-for-agent

## 按顺序验证

1. 运行 `gofmt -w .`，再运行 `git diff --check`。
   - 通过：`git diff --check` 无输出。
2. 运行 `go vet ./...`。
   - 通过：退出码 0，无 vet 错误。
3. 运行 `go test ./... -count=1`。
   - 通过：所有包为 `ok`，无 FAIL。
4. 运行 `go test -race ./... -count=1`。
   - 通过：所有包为 `ok`，无 `DATA RACE`。
5. 运行前端 generate、lint、typecheck、build。
   - 通过：全部退出码为 0，生成客户端无无法解释的 diff。
6. 启动 API 和 Web，实际复现成功、空列表、非法参数和内部错误。
   - 通过：状态码、JSON、页面状态与自动测试一致。
7. 从浏览器 Network 复制 request ID，在 API JSON 日志中查找。
   - 通过：能定位唯一对应请求，且日志无敏感字段。

## 你必须交付的证据

- [ ] 上述命令的原始输出。
- [ ] 一条成功和一条失败请求。
- [ ] 一条匹配 request ID 的日志。
- [ ] 一张你自己画的 `browser -> client -> Gin -> handler -> service -> repository` 图。
- [ ] 2 分钟口述每一层职责。

任意检查失败都不能进入第二天。真实 OIDC 和 PostgreSQL 尚未实现，必须继续标为未验证。
