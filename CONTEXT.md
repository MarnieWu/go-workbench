# Personal Workbench / 个人工作台

个人工作台是用户可能选择执行之工作的单一事实来源。它将输入证据、任务建议和用户已经明确接受的任务分开。

The personal workbench is the single source of truth for work the user may choose to perform. It separates incoming evidence and suggestions from tasks the user has explicitly accepted.

## Language / 领域语言

**Capture（捕获记录）**:
一次来自对话、手工录入或已授权来源的提交记录。一条 Capture 可以产生零个或多个 Candidate；它是输入证据，不是 Task。

A record of one submission received from a conversation, manual entry, or approved source. One Capture can produce zero or more Candidates; it is evidence of input, not a Task.

_Avoid / 避免_: Raw task, inbox item / 原始任务、收件箱项

**Candidate（候选任务）**:
从一个 Capture 中提出、等待用户审核的任务建议。Candidate 只能被接受一次；被拒绝后不再出现在默认 Inbox，但仍保留为审核记录。

A proposed task derived from one Capture and awaiting user review. A Candidate can be accepted only once; after rejection, it leaves the default Inbox but remains as a review record.

_Avoid / 避免_: Draft task, captured task / 草稿任务、已捕获任务

**Inbox（收件箱）**:
展示待审核 Candidate 的视图。Inbox 不是独立存储的领域实体。

The view that presents Candidates awaiting review. Inbox is not a stored domain entity.

_Avoid / 避免_: Capture queue / 捕获队列

**Task（正式任务）**:
用户明确接受，或明确要求工作台创建的工作。直接创建的 Task 可以没有 Source Evidence；由 Candidate 创建的 Task 可以关联多个来源。完成表示工作进度，归档表示从日常视图隐藏，两者相互独立且都可通过明确指令恢复。

Work the user has explicitly accepted or explicitly instructed the workbench to create. A directly created Task can have no Source Evidence, while a Task created from a Candidate can reference multiple sources. Completion represents work progress and archival represents removal from routine views; they are independent and can each be reversed by explicit instruction.

_Avoid / 避免_: Candidate, TODO suggestion / 候选任务、TODO 建议

**Check Item（检查项）**:
完成一个 Task 所需、但不值得独立安排、设置优先级或指定 Project 的步骤。

A step needed to complete a Task that does not justify independent scheduling, priority, or Project assignment.

_Avoid / 避免_: Subtask, child task / 子任务

**Project（项目）**:
用户明确创建、用于组织 Task 的工作范围。Agent 可以建议现有 Project，但不能自动创建 Project。MVP 中 Project 只区分 active 和 archived；归档不改变其中 Task 的业务状态，也不允许新 Task 再分配给该 Project。

A user-created scope used to organize Tasks. An Agent can suggest an existing Project but cannot create one automatically. In the MVP a Project is either active or archived; archival does not change the business state of its Tasks and prevents new Tasks from being assigned to it.

_Avoid / 避免_: Label, topic / 标签、主题

**Pending Action（待确认操作）**:
尚未在工作台中确认的完成、归档或批量操作建议。它绑定目标及其确认时的版本、操作和参数；目标变化或超过确认时限后失效。确认时必须再次验证目标版本，失效操作不能自动应用到新版本。

A proposed completion, archive, or bulk operation that has not been confirmed in the workbench. It is bound to the target and its version, operation, and parameters at proposal time; confirmation rechecks the target version, and an invalidated action cannot be applied automatically to a newer version.

_Avoid / 避免_: Approval, queued task / 审批、排队任务

**Label（标签）**:
为 Candidate 建议的轻量分类，仅在 Candidate 被接受时复制到 Task。Label 不是独立领域实体。

A lightweight classification suggested for a Candidate and copied to a Task only when the Candidate is accepted. A Label is not an independent domain entity.

_Avoid / 避免_: Tag

**Source Subscription（来源订阅）**:
用户明确授权工作台在限定范围内检查新输入的外部位置。暂停会停止抓取但允许恢复；撤销会停止抓取并移除访问凭据，但不会隐式删除已有来源证据。Source Subscription 既不是 Capture，也不是 Task。

A user-approved external location that the workbench may check for new input within an explicit scope. Pausing stops ingestion but allows resumption; revocation stops ingestion and removes access credentials without implicitly deleting existing provenance. A Source Subscription is neither a Capture nor a Task.

_Avoid / 避免_: Scraper, polling task, source / 抓取器、轮询任务、来源

**Source Evidence（来源证据）**:
标识 Capture 来源所需的最小且不可变引用。用户可以编辑 Candidate 建议，但不能改写原始 Capture 或 Source Evidence。多个 Source Evidence 可以支持同一个 Task，且不会被自动合并或丢弃。

The minimal immutable reference that identifies where a Capture came from. A user can edit a Candidate proposal but cannot rewrite the original Capture or Source Evidence. Multiple pieces of Source Evidence can support one Task without being automatically merged or discarded.

_Avoid / 避免_: Full source content, conversation archive / 完整来源内容、完整会话归档

**Audit Event（审计事件）**:
对领域操作形成的不可变、仅追加记录。它通过实体标识和允许记录的摘要描述发生了什么，不保存敏感正文。

An immutable, append-only record of a domain action. It identifies what happened through entity references and allowlisted summaries without storing sensitive content.

_Avoid / 避免_: Mutable activity log, application log / 可修改活动记录、应用日志

**Owner（所有者）**:
通过托管身份服务认证、拥有私有工作台及其中全部数据的人。MVP 只有一个 Owner，不包含成员、角色或共享 Project。

The person authenticated by the managed identity provider who owns the private workbench and all data within it. The MVP has one Owner and no members, roles, or shared Projects.

_Avoid / 避免_: Member, tenant, account / 成员、租户、账户

