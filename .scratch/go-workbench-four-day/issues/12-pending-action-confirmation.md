# 12 — 先提议，再确认完成或归档

**结果：** 完成/归档请求只创建 Pending Action；用户确认后才改 Task，过期或旧版本 Action 不执行。

**Blocked by:** 10、11。

**Status:** ready-for-agent

## 步骤与验证

1. 写测试证明 propose 后 Task 没变化，只多一条 pending Action。
2. 写有效确认、旧 version、24 小时过期、取消、重复确认、执行失败测试。
   - 验证：`go test ./... -run 'Test.*PendingAction' -count=1`。
   - RED 通过：失败原因是确认逻辑未实现。
3. 你实现状态机和事务；时间通过可替换 clock 注入，测试不要真实等待 24 小时。
4. 重跑测试。
   - 通过：有效确认变 executed；stale/expired/cancelled 不改 Task；失败不误标 executed。
5. 通过页面完成 propose、cancel、confirm 各一次。
   - 通过：mutation 时按钮禁用，失败时上下文保留。
6. 运行完整 tests 和 race。
   - 通过：全部 `ok`。

## 最终通过

- [ ] propose 不直接改 Task。
- [ ] confirm 再检查 owner、version、参数和过期时间。
- [ ] 重复确认不会重复执行。
- [ ] 成功动作和 Audit Event 原子提交。
