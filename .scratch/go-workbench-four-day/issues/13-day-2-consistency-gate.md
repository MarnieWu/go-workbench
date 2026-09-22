# 13 — 第二天总验收

**结果：** 用真实隔离 PostgreSQL 证明 owner scope、幂等、事务、version 和 Pending Action 正确。

**Blocked by:** 09、10、11、12。

**Status:** ready-for-agent

## 按顺序验证

1. `docker compose -f deploy/compose.yaml config`：退出码 0。
2. 启动 db 并检查 healthy；不是 healthy 就停止。
3. 对空测试库运行 migration：成功且无真实数据库连接。
4. `go test ./... -count=1`：全部 `ok`。
5. `go test -race ./... -count=1`：全部 `ok`，无 DATA RACE。
6. 运行 integration tests：必须实际访问隔离 PostgreSQL，不使用 mock 冒充。
7. 停止再启动 db，重跑关键集成测试：结果仍为 `ok`。

## 必须核对的数据库结果

- [ ] Owner A 查不到 Owner B 数据。
- [ ] 并发幂等提交只产生一个 Capture。
- [ ] Candidate 失败注入没有半完成记录。
- [ ] 旧 version 没有改变 Task。
- [ ] stale/expired Pending Action 没有执行。

## 通过标准

所有命令退出码为 0，上述五项均有测试名和数据库记录证据。任何一项失败，不进入第三天；不得靠重复运行把偶发失败当通过。
