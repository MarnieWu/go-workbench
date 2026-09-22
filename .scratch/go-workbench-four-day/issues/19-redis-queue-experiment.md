# 19 — 做一个可删除的 Redis 队列实验

**结果：** 验证 Redis 的重复交付、有限重试、超时和 DLQ，但 PostgreSQL 仍是业务事实来源。

**Blocked by:** 15。

**Status:** ready-for-agent

## 步骤与验证

1. 让 Agent 把 Redis 放在独立 Compose profile/实验目录。
   - 验证：关闭 profile 时核心 Compose config 仍成功。
2. 定义 job ID 与 Capture ID 的关系，写正常 enqueue/consume/ack 测试。
   - 通过：一个 job 对应已有 Capture，不创建第二套业务状态。
3. 模拟消费后、ack 前崩溃。
   - 通过：消息可重复交付，但幂等规则阻止重复 Candidate/Task。
4. 模拟重试上限和 DLQ。
   - 通过：达到上限只进入一次 DLQ，可看到原因摘要。
5. 停止 Redis，再恢复。
   - 通过：核心数据不损坏；故障策略与恢复步骤可重复。
6. 关闭 Redis profile，运行核心 tests 和 Compose。
   - 通过：全部通过，证明实验可拆卸。

## 最终通过

- [ ] Redis 不是 source of truth。
- [ ] 重复消息不产生重复业务结果。
- [ ] retry/timeout 有上限。
- [ ] 关闭实验后核心仍通过。
