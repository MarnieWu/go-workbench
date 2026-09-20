# Separate Task progress from archival / 分离任务进度与归档

## 中文

Task 的业务进度只使用 `backlog | in_progress | blocked | done`；归档通过独立的生命周期信息表示，不作为进度状态。这样可以区分“工作是否完成”和“是否出现在日常视图”，并避免 `next` 同时承担优先级和状态语义。MVP 不提供批量完成或批量归档，但 Pending Action 边界可以在未来扩展到批量操作。

## English

Task progress uses only `backlog | in_progress | blocked | done`; archival is represented by separate lifecycle data rather than a progress state. This distinguishes whether work is complete from whether it appears in routine views and avoids making `next` serve as both priority and workflow state. The MVP provides no bulk completion or archival, while the Pending Action boundary can support bulk operations later.
