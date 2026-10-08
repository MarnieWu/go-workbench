# 23 — Worker 安全停止

**What to build:** Worker 收到终止信号后停止领取新 Capture，在超时内完成或释放当前工作；重启后 queued 工作继续处理。

**Blocked by:** 22 — Worker 并发、有限重试和人工重放。

**Status:** ready-for-agent

**Collaboration:** 共同完成。你主写 shutdown 协调、context cancellation 和资源关闭顺序；我主写进程级测试、日志核对和 leak 检查脚手架。

## 关键步骤

1. 写停止后不再 claim、短任务完成、长任务超时释放三类测试。
2. 用户实现 signal context、goroutine 等待和数据库池关闭顺序。
3. 用真实 worker 进程发送一次终止信号，记录日志顺序。
4. 检查 goroutine leak 和 race。

## 改动范围

- Runtime：worker main、shutdown timeout、signal handling。
- Data：processing 工作的安全释放。
- Logs：停止领取、完成/释放、关闭资源。
- Tests：shutdown、race、leak 检查。

- [ ] shutdown 后不领取新工作。
- [ ] 未完成 Capture 可恢复。
- [ ] 数据库连接最后关闭。
- [ ] 重复 shutdown 不 panic。
