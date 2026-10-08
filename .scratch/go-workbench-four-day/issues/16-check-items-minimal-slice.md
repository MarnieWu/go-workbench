# 16 — 管理 Task 的 Check Item

**What to build:** 用户能为一个 Task 添加、勾选、重新排序和查看 Check Item；Check Item 不拥有独立 Project、优先级或 Task 状态。

**Blocked by:** 15 — 用 version 更新 Task。

**Status:** ready-for-agent

**Collaboration:** 共同完成。你主写 owner scope、position 约束和核心 repository 测试；我主写 API 接线、UI 交互、可访问性状态和样板测试。

## 关键步骤

1. 写 repository RED 测试：按 position 排序、跨 owner 拒绝、重复 position 冲突。
2. 接通 create、patch 和 list 的服务与 HTTP；每次写入都校验 Task owner。
3. Web 在 Task 详情中展示检查项，mutation 期间禁用重复操作。
4. 补键盘操作和可访问名称。

## 改动范围

- Domain：Check Item 输入、position 和完成状态。
- Data：Check Item 查询和写入，Task owner scope。
- API：Task 下的 Check Item endpoints。
- UI：Task detail 的检查项区域。
- Tests：repository、router、Web typecheck、可访问状态。

- [ ] Check Item 只属于 Task，不独立关联 Project。
- [ ] Owner A 不能读写 Owner B Task 的 Check Item。
- [ ] 列表按 position 稳定排序。
- [ ] 失败 mutation 保留页面状态。
