# 17 — 完成或归档只创建 Pending Action

**What to build:** 用户请求完成或归档 Task 时，系统只创建 Pending Action，不立即修改 Task；页面显示待确认操作和过期时间。

**Blocked by:** 15 — 用 version 更新 Task。

**Status:** ready-for-agent

**Collaboration:** 你主写。你实现确认边界、version 绑定和 propose 规则；我主写页面提示、HTTP 接线和验收记录模板。

## 关键步骤

1. 写 propose RED 测试：complete、archive、跨 owner、旧 version、重复 pending。
2. 用户实现 Pending Action 创建规则，绑定 target ID、target version、action、parameters 和 expires_at。
3. 接通 HTTP 和 Web；Task 当场保持原状态。
4. 页面展示 pending、cancel 入口和确认入口，但本票不执行确认。

## 改动范围

- Domain：Pending Action propose 规则。
- Data：pending_actions insert 和唯一性策略。
- API：propose endpoint。
- UI：Pending Action 列表和 Task 上的待确认提示。
- Tests：repository、router、Web state。

- [ ] propose 后 Task 状态和 archived_at 不变。
- [ ] Pending Action 绑定目标 version 和 24 小时过期时间。
- [ ] 跨 owner 和旧 version 返回稳定错误。
- [ ] 页面能区分“已提议”与“已执行”。
