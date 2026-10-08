# 20 — 第二阶段一致性总验收

**What to build:** 用隔离 PostgreSQL 和真实 HTTP router 证明 Capture、Candidate、Task、Check Item、Project、Pending Action 和 Audit Event 的一致性规则成立。

**Blocked by:** 10、12、14、16、18、19。

**Status:** ready-for-agent

**Collaboration:** 共同完成。我整理门禁命令、验收记录和失败归因；你运行关键命令、读完整输出，并判断是否允许进入下一阶段。

## 关键步骤

1. 按顺序运行 Go 单元、集成、race、vet 和 Compose config。
2. 在隔离测试库中验证 owner scope、幂等、事务回滚、version 冲突、Pending Action 过期和 archived Project 拒绝。
3. 检查 OpenAPI 生成客户端无漂移。
4. 记录未验证项，不把 blocked 或 unrun 写成 passed。

## 改动范围

- Tests：聚合验证脚本或 Make target。
- Docs：更新验收清单中的证据列。
- Fixes：只修复 gate 发现的最小问题。

- [ ] 所有相关 Go 测试和 race 检查通过。
- [ ] 数据库验证只连接隔离测试数据库。
- [ ] OpenAPI generated client 无漂移。
- [ ] 验收记录区分事实、推测和未验证假设。
