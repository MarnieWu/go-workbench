# 16 — 让 Worker 安全停止

**结果：** 收到 SIGTERM 后停止领取新任务，在超时内完成或安全释放当前任务，恢复后继续 queued 工作。

**Blocked by:** 15。

**Status:** ready-for-agent

## 步骤与验证

1. 写三个测试：停止后不再 claim、短任务完成、长任务超时释放。
   - 验证：目标测试因 shutdown 协调未实现而 RED。
2. 你使用 `signal.NotifyContext`、context cancellation 和明确的 goroutine 等待实现退出。
   - 核对：每个 goroutine 都有停止条件；channel 只由发送方 owner 关闭。
3. 重跑测试。
   - 通过：没有新 claim；短任务完成；长任务不会永久卡在 processing。
4. 启动真实 Worker，发送一次 SIGTERM。
   - 核对：日志顺序是停止领取、处理/释放当前工作、关闭数据库、退出。
   - 通过：进程在配置超时内退出，重启后 queued Capture 继续处理。
5. 运行 race 和 goroutine leak 检查。
   - 通过：无 DATA RACE、无 leak 报告。

## 最终通过

- [ ] shutdown 后不领取新工作。
- [ ] 未完成 Capture 可恢复。
- [ ] 数据库池最后关闭。
- [ ] 重复 shutdown 不 panic。
