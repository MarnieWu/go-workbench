# 04 — 给每个请求加入 Request ID

**结果：** Go API 为每个请求生成 request ID，不沿用调用方传入的 `X-Request-ID`；响应 header、错误 JSON 和日志使用同一个 ID。

**Blocked by:** 02。

**Status:** ready-for-agent

## 步骤

1. 让 Agent 建立 middleware 测试空壳，你填写三种场景：请求未带 ID、请求带有预设 ID、两个独立请求。
   - 验证：`go test ./... -run 'Test.*RequestID' -count=1` 应因 middleware 缺失而 RED。
   - 通过：测试分别要求响应 ID 非空、预设 ID 未被沿用、两个请求的 ID 不同；失败原因与 request ID 行为缺失有关。
2. 实现 middleware：为每个请求生成 ID，写入 context 和响应 header；不读取或沿用调用方传入的 `X-Request-ID`。
   - 验证：目标测试应显示 `ok`。
   - 通过：每个请求都有由 Go API 生成的 ID，并行请求不会复用或串用 ID。
3. 在错误响应和 `slog` JSON 中读取同一 ID。
   - 验证：测试比较响应 header、JSON 字段和捕获日志的 ID。
   - 通过：三处完全一致。
4. 先运行普通测试，再运行 race detector：
   ```bash
   go test ./... -count=1
   go test -race ./... -count=1
   ```
   - 原因：普通测试验证各 package 的编译结果和功能断言；race detector 在执行同一批测试时监测并发读写共享内存的数据竞争。request ID middleware 会同时处理多个请求，因此既要验证功能，也要检查是否错误复用了可变的全局状态。
   - 参数：`./...` 表示当前 module 下的全部 package；`-count=1` 禁用测试结果缓存，确保本次实际执行；`-race` 启用 Go 数据竞争检测器。
   - 限制：`-race` 只能发现本次测试实际执行路径中的数据竞争，测试通过不代表所有并发路径都已覆盖。
   - 通过：两条命令退出码均为 0，相关 package 显示 `ok`。普通测试失败时先修复，不继续把 race 结果解释为有效验收。

## 延后验证（不阻塞本 issue）

当前 `cmd/api/main.go` 没有 HTTP server bootstrap，`go run ./cmd/api` 会正常退出，不能进行真实 `curl` 验证。不要只为本 issue 增加临时 server 或假 repository。

API bootstrap 完成后，再启动 API 并执行 `curl -H 'X-Request-ID: learn-go-001' ...`：

- 响应 header 返回非空 ID 且不等于 `learn-go-001`。
- 错误 JSON 和日志包含响应 header 中的同一个 ID。
- 日志不包含调用方预设 ID。
- 能用响应 header 中的 ID 找到该请求的日志。

该验证必须在 21 的端到端交付门禁中再次执行。在实际完成前，只能记录为 deferred，不能声称通过。

## 最终通过

- [ ] 每个响应都有 `X-Request-ID`。
- [ ] 不沿用调用方传入的 `X-Request-ID`。
- [ ] 不同请求使用不同的 request ID。
- [ ] 日志不包含 Authorization、Cookie 或完整正文。
- [ ] 目标测试和 race 测试均验证成功。

完成后解释：request ID 用于关联日志，不等同于分布式 trace。
