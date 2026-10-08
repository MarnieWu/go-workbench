# 14 — 接受 Candidate 的回滚和并发保护

**What to build:** 接受 Candidate 时，任一写入失败都不留下半完成状态；两个请求并发接受同一 Candidate 时最多一个成功。

**Blocked by:** 13 — 接受 Candidate 的成功事务。

**Status:** ready-for-agent

**Collaboration:** 你主写。你实现失败注入、rollback 和并发保护；我帮你构造测试夹具、解释失败输出和审查 race 风险。

## 关键步骤

1. 写失败注入测试：建 Task 后失败、连 Evidence 后失败、写 Audit Event 前失败。
2. 写并发接受测试，不能依赖 sleep 假装同步。
3. 用户补事务回滚和并发条件；Agent 帮忙整理测试辅助和错误映射。
4. Web 收到冲突时保留用户编辑内容，并提示刷新。

## 改动范围

- Domain：状态冲突和业务错误。
- Data：事务 rollback、唯一约束、affected rows 检查。
- API：`409 CANDIDATE_STATE_CONFLICT` 或等价稳定 code。
- UI：冲突恢复。
- Tests：失败注入、并发、race、router。

- [ ] 任一失败点后 Task、link、Candidate、Audit Event 没有部分提交。
- [ ] 并发接受最多一个成功。
- [ ] Source Evidence 保持不可变。
- [ ] 冲突响应不泄露 SQL 细节。
