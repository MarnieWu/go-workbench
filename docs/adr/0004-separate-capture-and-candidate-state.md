# Separate Capture and Candidate state / 分离 Capture 与 Candidate 状态

## 中文

Capture 使用 `queued -> processing -> processed | failed` 描述系统处理输入的结果；Candidate 使用 `pending_review -> accepted | rejected` 描述用户审核任务建议的结果。一个 Capture 的全部 Candidate 在同一事务中创建，任一项校验失败则整体失败；人工重放复用原 Capture 和幂等身份。分离两条状态机可以避免把系统处理结果和用户决定混为一谈。

## English

Capture uses `queued -> processing -> processed | failed` to describe system processing, while Candidate uses `pending_review -> accepted | rejected` to describe user review. All Candidates from one Capture are created in one transaction; if any item fails validation, the processing attempt fails as a whole, and manual replay reuses the original Capture and idempotency identity. Separate state machines prevent system processing outcomes from being confused with user decisions.
