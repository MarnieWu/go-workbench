# 19 — Project 列表、创建和归档确认

**What to build:** 用户能列出和明确创建 Project；归档 Project 走 Pending Action 确认，归档后不级联修改 Task，且拒绝新 Task 分配给 archived Project。

**Blocked by:** 18 — 确认、取消和过期 Pending Action。

**Status:** ready-for-agent

**Collaboration:** 共同完成。你主写 Project 归档规则和 archived Project 拒绝逻辑；我主写 Project 页面、OpenAPI、常规 CRUD 胶水和回归测试。

## 关键步骤

1. 写 Project create/list 的 owner scope 和校验测试。
2. 写归档 propose + confirm 测试，确认不级联更新现有 Task。
3. 在 Task 创建或更新路径拒绝 archived Project。
4. Web 增加 Project 页的列表、创建、归档 pending 状态。

## 改动范围

- Domain：Project create、archive 和 active/archived 规则。
- Data：projects 查询、写入、version 和 Pending Action 集成。
- API：Project endpoints。
- UI：Project 页面和状态。
- Tests：repository、router、Web typecheck。

- [ ] Agent 或 worker 不会自动创建 Project。
- [ ] Project 归档不改变既有 Task 的业务状态。
- [ ] archived Project 不能接收新 Task。
- [ ] 页面明确展示 active 与 archived。
