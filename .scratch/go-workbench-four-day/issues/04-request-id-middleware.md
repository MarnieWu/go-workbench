# 04 — 给每个请求加入 Request ID

**结果：** 客户端提供合法 request ID 时沿用；没有时生成；响应 header、错误 JSON 和日志使用同一个 ID。

**Blocked by:** 02。

**Status:** ready-for-agent

## 步骤

1. 让 Agent 建立 middleware 测试空壳，你填写三种输入：合法、缺失、超长/非法。
   - 验证：`go test ./... -run 'Test.*RequestID' -count=1` 应因 middleware 缺失而 RED。
   - 通过：失败原因与 request ID 行为缺失有关。
2. 实现 middleware：读取 header、校验长度/字符、必要时生成 ID、写入 context 和响应 header。
   - 验证：目标测试应显示 `ok`。
   - 通过：三种输入均有确定结果，并行测试不会串 ID。
3. 在错误响应和 `slog` JSON 中读取同一 ID。
   - 验证：测试比较响应 header、JSON 字段和捕获日志的 ID。
   - 通过：三处完全一致。
4. 实际启动 API，用 `curl -H 'X-Request-ID: learn-go-001' ...` 请求一次。
   - 验证：响应 header 和日志都含 `learn-go-001`。
   - 通过：能用该 ID 找到唯一请求日志。
5. 运行 `go test ./... -count=1 && go test -race ./... -count=1`。
   - 通过：两条命令均显示 `ok`。

## 最终通过

- [ ] 每个响应都有 `X-Request-ID`。
- [ ] 非法 ID 不直接进入日志。
- [ ] 日志不包含 Authorization、Cookie 或完整正文。
- [ ] 测试、race 和实际 curl 都验证成功。

完成后解释：request ID 用于关联日志，不等同于分布式 trace。
