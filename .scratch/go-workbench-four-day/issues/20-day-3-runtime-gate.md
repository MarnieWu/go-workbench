# 20 — 第三天总验收

**结果：** Worker、认证和 MCP 在正常、失败、并发、终止场景下可验证，Redis 可完全关闭。

**Blocked by:** 16、17、18、19。

**Status:** ready-for-agent

## 按顺序验证

1. `gofmt -w . && git diff --check`：diff check 无输出。
2. `go vet ./...`：退出码 0。
3. `go test ./... -count=1`：全部 `ok`。
4. `go test -race ./... -count=1`：无 FAIL、无 DATA RACE。
5. `govulncheck ./...`：通过或每个发现有书面处置；工具不可用则 blocked。
6. 启动 API、Worker、MCP 三个入口：各自可独立启动和停止。
7. 发送 SIGTERM：Worker 在超时内退出，恢复后继续 queued Capture。
8. 关闭 Redis profile：核心 tests 和 Compose 仍通过。

## 必须核对

- [ ] 并发 Worker 只处理一次。
- [ ] retry 最终停止。
- [ ] 认证负向用例全部拒绝。
- [ ] MCP complete/archive 不直接改 Task。
- [ ] 日志不含 token、Cookie、完整正文或 secret。

任何重复处理、权限绕过、无限重试、goroutine 泄漏或 race 都必须先修复，不能进入第四天。
