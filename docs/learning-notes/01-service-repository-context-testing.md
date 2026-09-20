# 01 — Service、Repository、Context 与测试盲区

## 对应实现任务

让 `task.Service.List` 调用 `Repository.List`，并原样返回任务列表和错误。

## 问题 1：为什么 Service 依赖 Repository 接口，而不是直接执行 SQL？

### 代码现场

```go
type Repository interface {
	List(ctx context.Context, ownerID string) ([]Task, error)
}

type Service struct {
	repository Repository
}
```

### 你的回答

> 因为我们封装了统一处理数据库的接口类型。

### 面试官点评

方向接近，但“统一处理数据库”不准确。`Repository` 不是通用数据库接口，只定义当前业务需要的存储能力。`Service` 因此不依赖 PostgreSQL、SQL 或具体驱动。测试可以使用假的 Repository，生产环境可以使用真实数据库实现。

### 标准口述答案

> Service 依赖 Repository 接口，是为了让业务逻辑只依赖所需能力，不依赖具体存储技术。这样可以独立测试 Service，并在不改变业务逻辑的情况下替换存储实现。

### 高价值追问

如果目前只有一个 PostgreSQL 实现，为什么仍然值得保留这个接口？

答题要点：重点不是预先支持多个数据库，而是隔离业务逻辑与 I/O，并让 Service 测试不需要真实数据库。

## 问题 2：为什么要把调用者传入的 ctx 继续传给 Repository？

### 目标代码

```go
func (s *Service) List(ctx context.Context, ownerID string) ([]Task, error) {
	return s.repository.List(ctx, ownerID)
}
```

### 你的回答

> 因为 List 函数接收 ctx 参数。

### 面试官点评

这是代码表象，不是设计原因。继续传递同一个 `ctx`，才能让请求取消和截止时间传到数据库调用。否则上层请求已经结束，底层查询仍可能继续占用连接和计算资源。

### 标准口述答案

> `ctx` 表示一次调用的生命周期。Service 必须把它继续传给 Repository，使取消和超时贯穿调用链。Service 不应该在这里改用 `context.Background()`，因为那会切断上层的取消信号。

### 高价值追问

为什么不应该把 `context.Context` 保存到 `Service` 结构体中？

答题要点：`ctx` 属于单次调用，不属于长生命周期对象；保存后容易跨请求复用错误的取消、超时或请求范围数据。

## 问题 3：为什么当前测试通过仍不能证明 ownerID 一定传对了？

### 当前测试代码

```go
func (r stubRepository) List(_ context.Context, _ string) ([]Task, error) {
	return r.tasks, r.err
}
```

### 你的回答

> 因为测试代码中 ownerID 是我们目前写死的。

### 面试官点评

结论方向正确。更直接的原因是假的 Repository 忽略了收到的 `ownerID`。因此，即使 Service 传入空字符串或错误值，测试仍会返回预设任务并通过。

两个参数都使用 `_`，测试没有记录或断言它们。

### 标准口述答案

> 当前 stub 只返回预设结果，没有观察收到的 `ownerID`。所以测试只能证明 Service 返回了 Repository 的结果，不能证明 Service 正确转发了调用参数。需要使用能够记录参数的 fake 或 spy，并对收到的值进行断言。

### 后续测试示例

```go
type recordingRepository struct {
	ownerID string
}

func (r *recordingRepository) List(_ context.Context, ownerID string) ([]Task, error) {
	r.ownerID = ownerID
	return nil, nil
}

func TestServiceListPassesOwnerID(t *testing.T) {
	repository := &recordingRepository{}
	service := NewService(repository)

	_, err := service.List(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if repository.ownerID != "owner-1" {
		t.Fatalf("ownerID = %q, want %q", repository.ownerID, "owner-1")
	}
}
```

该示例用于后续任务。当前任务先完成最小 `Service.List` 实现，不提前扩展测试。

### 高价值追问

除了 `ownerID`，当前测试还没有证明哪个参数被正确传递？

答题要点：`ctx`。可以让 fake Repository 记录收到的 `ctx`，并验证它与调用 Service 时传入的是同一个值。

## 本节结论

- Repository 接口隔离业务逻辑与具体 I/O，不是通用数据库封装。
- `ctx` 必须沿调用链传递，以保留取消和超时语义。
- 只断言返回值的测试可能遗漏参数转发错误。
- Fake 返回结果；Spy 记录调用。一个测试替身可以同时承担两种职责。
