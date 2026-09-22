# 10 — 用一个事务接受 Candidate

**结果：** 接受 Candidate 时，要么同时完成“建 Task、连 Evidence、改 Candidate、写 Audit Event”，要么四项都不发生。

**Blocked by:** 08、09。

**Status:** ready-for-agent

## 步骤与验证

1. 画出事务内 4 个写步骤，标出每步失败后预期数据库状态。
2. 先写成功测试和至少两个失败注入测试。
   - 验证：`go test ./... -run 'Test.*AcceptCandidate' -count=1`。
   - RED 通过：事务未实现导致失败；数据库连接错误不算。
3. 你使用 `pgx.Tx` 实现事务；任一步失败都 rollback。
   - 核对：commit 只在最后发生；rollback 错误不会覆盖原始业务错误。
4. 重跑测试并直接查询测试库记录数量。
   - 通过：成功时四项齐全；失败时四项都没有部分变化。
5. 写两个并发接受同一 Candidate 的测试。
   - 通过：最多一个成功，数据库恰好一个 Task。
6. 运行完整 tests 和 race。
   - 通过：全部 `ok`，无 DATA RACE。

## 最终通过

- [ ] 成功事务四项一致。
- [ ] 两个失败点都证明完整回滚。
- [ ] 并发接受最多一次成功。
- [ ] Source Evidence 未被修改。
