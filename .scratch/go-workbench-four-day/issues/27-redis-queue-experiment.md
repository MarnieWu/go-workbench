# 27 — 可删除的 Redis 队列实验

**What to build:** 在独立实验中验证 Redis 重复交付、有限重试、超时和 DLQ；PostgreSQL 仍是业务事实来源，关闭实验后核心系统继续通过。

**Blocked by:** 22 — Worker 并发、有限重试和人工重放。

**Status:** ready-for-agent

**Collaboration:** 我主写。你审查 Redis 是否没有侵入核心业务事实来源；我实现隔离实验、故障复现、DLQ 记录和可拆卸验证。

## 关键步骤

1. 把 Redis 放在独立 Compose profile 或实验目录。
2. 定义 job ID 与 Capture ID 的关系，不能创建第二套业务状态。
3. 模拟消费后 ack 前崩溃、重试上限、DLQ 和 Redis 不可用。
4. 关闭 Redis profile 后运行核心测试和 Compose config。

## 改动范围

- Experiments：Redis queue profile 和实验 worker。
- Docs：故障策略、DLQ 观察、恢复步骤。
- Tests：重复交付、上限、不可用、核心关闭实验仍通过。

- [ ] Redis 不是 source of truth。
- [ ] 重复消息不产生重复 Candidate 或 Task。
- [ ] retry、timeout、DLQ 有上限和证据。
- [ ] 关闭实验后核心构建、测试和 Compose 不依赖 Redis。
