# 09 — 幂等创建 Capture

**结果：** 同一 Owner 使用相同幂等键重复提交相同内容，只得到一个 Capture；内容不同则返回 `409 IDEMPOTENCY_CONFLICT`。

**Blocked by:** 07。

**Status:** ready-for-agent

## 步骤与验证

1. 先写 4 个集成测试：首次提交、相同重试、不同内容冲突、两个并发相同请求。
   - 验证：`go test ./... -run 'Test.*Capture.*Idempoten' -count=1`。
   - RED 通过：失败原因是逻辑未实现。
2. 你实现输入 hash、Repository unique 冲突处理和 Service 规则。
   - 核对：Capture 初始状态为 queued，不创建 Candidate/Task。
3. 重跑目标测试。
   - 通过：相同请求返回同一 ID；冲突返回稳定业务错误；并发后数据库只有一行。
4. 写 router 测试检查 `409`、错误 code 和 request ID。
   - 通过：测试 `ok`，响应不泄露另一请求正文。
5. 运行 `go test ./... -count=1 && go test -race ./... -count=1`。
   - 通过：无 FAIL、无 DATA RACE。

## 最终通过

- [ ] 顺序重试和并发重试都只创建一个 Capture。
- [ ] 不同 Owner 可复用相同幂等键。
- [ ] 不同内容返回 409，不自动合并语义相似请求。
- [ ] Capture 只保存最小来源证据。
