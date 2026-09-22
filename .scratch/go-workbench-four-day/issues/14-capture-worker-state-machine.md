# 14 — 让 Worker 处理一个 Capture

**结果：** Worker 领取 queued Capture，创建 0..N 个 Candidate，然后把 Capture 改为 processed；失败时不留下部分 Candidate。

**Blocked by:** 13。

**Status:** ready-for-agent

## 步骤与验证

1. 画状态：`queued -> processing -> processed | failed`，另画 Candidate 的 `pending_review -> accepted | rejected`。
   - 通过：两条状态机没有混在一起。
2. 让 Agent 创建 Worker 启动空壳和测试 fixture，你写 0、1、N Candidate 及非法中间项测试。
   - 验证：`go test ./... -run 'Test.*Worker.*Capture' -count=1` 应 RED。
3. 你实现领取、校验和单事务写入。
   - 核对：任一 Candidate 非法时，本次一个 Candidate 都不提交。
4. 重跑目标测试并查询数据库。
   - 通过：0/1/N 场景数量正确；失败场景无部分记录；Capture 状态正确。
5. 运行 `go test ./... -count=1 && go test -race ./... -count=1`。
   - 通过：全部 `ok`，无 DATA RACE。

## 最终通过

- [ ] 领取前只处理 queued Capture。
- [ ] Candidate 批次全有或全无。
- [ ] 错误摘要有长度限制且无敏感正文。
- [ ] API 和 Worker 复用业务 Service/Repository。
