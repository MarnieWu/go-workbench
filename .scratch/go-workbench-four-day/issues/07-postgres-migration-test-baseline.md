# 07 — 建立测试 PostgreSQL 和第一版数据库结构

**结果：** 只连接隔离测试数据库，并能从空库创建核心表和约束。

**Blocked by:** 06。

**Status:** ready-for-agent

## 步骤与验证

1. 让 Agent 创建 PostgreSQL Compose、脱敏示例配置和 migration runner，不写表结构。
   - 验证：`docker compose -f deploy/compose.yaml config`。
   - 通过：退出码 0，输出中没有真实密码或未解析变量。
2. 启动仅数据库服务：`docker compose -f deploy/compose.yaml up -d db`。
   - 验证：运行 `docker compose -f deploy/compose.yaml ps`。
   - 通过：db 状态为 healthy；否则先看 logs，不继续 migration。
3. 你写 migration：Owner、Project、Task、Capture、Candidate、Source Evidence、关联表、Pending Action、Audit Event。
   - 核对：状态有 CHECK；关系有 foreign key；幂等身份有 unique；owner 字段/关系清楚。
   - 通过：能够逐个解释约束防止什么错误。
4. 对空测试库运行 migration，再运行 schema 集成测试。
   - 验证：按 runner 命令执行后运行 `go test ./... -run 'Test.*Migration' -count=1`。
   - 通过：输出 `ok`，重复执行不会意外破坏 schema。
5. 写非法状态和重复键测试。
   - 验证：测试必须看到数据库拒绝写入，并确认事务后无非法行。
   - 通过：目标测试 `ok`，不是手工观察。

## 最终通过

- [ ] 有防误连测试数据库保护。
- [ ] 空库 migration 和重复执行验证通过。
- [ ] CHECK、unique、foreign key 都有失败测试。
- [ ] 回滚限制写清楚，未执行生产 migration。
