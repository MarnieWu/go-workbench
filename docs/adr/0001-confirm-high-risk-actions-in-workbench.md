# Confirm high-risk actions in the workbench / 在工作台中确认高风险操作

## 中文

完成、归档或未来可能增加的批量操作先创建 Pending Action，不立即改变目标。Pending Action 绑定目标 ID、目标版本、操作和参数；目标变化或超过 24 小时后失效。用户必须在工作台中确认后才执行，因为仅信任 Agent 声称“用户已经在对话中确认”无法形成可强制执行的授权边界或可靠审计证据。MVP 不提供批量操作。

## English

Completing, archiving, or applying a future bulk operation creates a Pending Action instead of changing the target immediately. The Pending Action is bound to the target ID, target version, operation, and parameters, and expires after 24 hours or when the target changes. The user must confirm it in the workbench before execution because trusting an Agent to report conversational confirmation does not provide an enforceable authorization boundary or reliable audit evidence. The MVP provides no bulk operations.
