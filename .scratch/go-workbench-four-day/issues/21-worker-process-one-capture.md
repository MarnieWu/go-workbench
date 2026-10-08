# 21 — Worker 处理一条 Capture

**What to build:** Worker 领取一条 `queued` Capture，生成 0..N 个 pending Candidate，并把 Capture 标记为 processed；任一候选项非法时本次不提交任何 Candidate。

**Blocked by:** 20 — 第二阶段一致性总验收。

**Status:** ready-for-agent

**Collaboration:** 你主写。你实现 Capture claim、Candidate 批次事务和状态机；我提供 worker 入口壳、fixture、RED 测试和审查。

## 关键步骤

1. 写状态机说明：Capture 与 Candidate 是两条状态机，不能混用。
2. 写 0、1、N Candidate 和非法中间项的 RED 测试。
3. 用户实现原子 claim、候选校验和单事务写入；Agent 只接 worker 启动壳和 fixture。
4. 验证 API、worker 复用同一领域规则和 repository。

## 改动范围

- Domain：Capture processing 规则和 Candidate proposal 模型。
- Data：atomic claim、Candidate batch insert、Capture status update。
- Runtime：worker 入口。
- Tests：worker repository、服务、race 基础。

- [ ] 只处理 `queued` Capture。
- [ ] 0/1/N Candidate 场景结果正确。
- [ ] 任一 Candidate 非法时本批次全不提交。
- [ ] 错误摘要有长度限制，且不含敏感正文。
