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
4. 实际启动 API，用 `curl -H 'X-Request-ID: learn-go-001' ...` 请求一次。
   - 验证：响应 header 返回非空 ID 且不等于 `learn-go-001`；日志包含响应中的 ID，不包含调用方预设 ID。
   - 通过：能用响应 header 中的 ID 找到该请求的日志。
5. 运行 `go test ./... -count=1 && go test -race ./... -count=1`。
   - 通过：两条命令均显示 `ok`。

## 最终通过

- [ ] 每个响应都有 `X-Request-ID`。
- [ ] 不沿用调用方传入的 `X-Request-ID`。
- [ ] 不同请求使用不同的 request ID。
- [ ] 日志不包含 Authorization、Cookie 或完整正文。
- [ ] 测试、race 和实际 curl 都验证成功。

完成后解释：request ID 用于关联日志，不等同于分布式 trace。
