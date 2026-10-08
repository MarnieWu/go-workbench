# 22 — Worker 并发、有限重试和人工重放

**What to build:** 多个 Worker 可以并行运行，但同一 Capture 同时只能被一个 Worker 领取；失败按有限次数重试，人工重放复用原 Capture 和幂等身份。

**Blocked by:** 21 — Worker 处理一条 Capture。

**Status:** ready-for-agent

**Collaboration:** 你主写。你实现并发 claim、retry 上限、failed 终态和 replay；我帮你构造无 sleep 的并发测试和 race 验证。

## 关键步骤

1. 写两个 Worker 抢同一 Capture 的并发测试，不能依赖 sleep。
2. 写 transient、permanent、达到上限和 replay 的测试。
3. 用户实现 claim 条件、attempt_count、终态 failed 和 replay 规则。
4. 目标并发测试至少重复运行 10 次，再跑 race。

## 改动范围

- Domain：retry、failed、replay 规则。
- Data：atomic claim、attempt_count、状态转移。
- Runtime：worker loop 的并发控制。
- Tests：并发、重复运行、race。

- [ ] 同一 Capture 不被重复处理。
- [ ] retry/backoff 有上限。
- [ ] 达到上限进入 `failed`，不会无限循环。
- [ ] replay 不创建第二条 Capture。
