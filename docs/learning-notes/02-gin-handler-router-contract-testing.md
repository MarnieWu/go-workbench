# 02：Gin Handler、HTTP 契约与 Router 测试

## 问题 1：请描述 `GET /v1/tasks` 的完整执行过程

### 回答

> 测试先调用 `router.ServeHTTP`，把请求交给 Gin。Router 根据请求方法和路径匹配 `GET /v1/tasks`，然后执行注册在 Router 上的 Middleware。测试 Middleware 把 `ownerID` 放进 Gin Context，再通过 `c.Next()` 进入 Handler。
>
> Handler 从 Gin Context 取出 `ownerID`，从 HTTP Request 取出 Go Context，然后调用 `Service.List`。Service 再调用 `Repository.List`。Repository 返回任务列表后，Handler 把领域模型转换成响应模型，最后由 `c.JSON` 写入状态码和 JSON 响应体。

```text
HTTP Request
    → Gin Router
    → Middleware
    → Handler
    → Service
    → Repository
    → Service
    → Handler 映射响应
    → JSON Response
```

### 子问题 1：Router 和 Handler 分别负责什么？

#### 回答

> Router 负责匹配请求方法和路径，并决定调用哪个 Handler。Handler 负责读取 HTTP 输入、调用 Service，再把结果写成 HTTP 响应。Router 解决“请求交给谁”，Handler 解决“这次请求怎么处理”。

### 子问题 2：Handler 为什么不直接调用 Repository？

#### 回答

> Handler 属于 HTTP 层，它不应该知道数据从 PostgreSQL、内存还是其他存储中取得。它只调用 Service。Service 负责业务流程，Repository 提供存储能力。这样更换存储实现时，HTTP 层不需要跟着修改。

### 子问题 3：为什么 Service 不接收 `*gin.Context`？

#### 回答

> `*gin.Context` 属于 Gin，包含读取请求和写响应的方法。Service 接收它以后会依赖 Gin，业务代码就不能脱离 HTTP 层使用和测试。Service 只需要 Go 标准库的 `context.Context`，用于传递取消和超时信号。

### 子问题 4：如果 Handler 改用 `context.Background()` 会怎样？

#### 回答

> `context.Background()` 会创建一条新的 Context 链。客户端断开连接或请求超时后，原请求的取消信号无法继续传给 Service 和 Repository，数据库查询可能继续占用连接和计算资源。

## 问题 2：ownerID 为什么由 Middleware 提供？

### 回答

> `ownerID` 决定当前用户能够读取哪些任务，所以服务端必须从已经验证的身份中取得它。任务 02 还没有实现认证，测试 Middleware 先把固定的 `owner-1` 写入 Gin Context，用来验证后面的调用链。生产环境需要由认证 Middleware 验证令牌或会话，再把可信的 ownerID 放入 Context。

```go
func testOwner(ownerID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ownerIDKey, ownerID)
		c.Next()
	}
}
```

### 子问题 1：为什么不能让客户端通过查询参数决定 ownerID？

#### 回答

> 客户端可以修改查询参数。如果接口相信客户端提交的 ownerID，用户可能把它改成其他人的 ID，读取不属于自己的任务。这属于越权访问。服务端应使用认证结果中的身份来限定查询范围。

### 子问题 2：`c.GetString(ownerIDKey)` 找不到值时会怎样？

#### 回答

> Gin 会返回空字符串。当前 Handler 会把空字符串继续传给 Service。任务 02 只验证成功路径；接入真实认证后，认证 Middleware 应在身份缺失时终止请求并返回 `401`，不让请求进入 Handler。

### 子问题 3：怎样证明 ownerID 到达了 Repository？

#### 回答

> 测试 Repository 在 `List` 中记录收到的 ownerID。请求结束后，测试断言 `gotOwnerID == "owner-1"`。只检查 ownerID 非空还不够，因为错误的非空值也能通过。

```go
func (r *stubTaskRepository) List(_ context.Context, ownerID string) ([]task.Task, error) {
	r.callCount++
	r.gotOwnerID = ownerID
	return r.tasks, r.err
}
```

### 子问题 4：测试 Middleware 能否证明真实认证已经完成？

#### 回答

> 不能。它没有读取凭证，也没有验证用户身份。它只为测试提供一个确定的 ownerID，让测试能够检查 Handler、Service 和 Repository 之间的参数传递。

## 问题 3：为什么要把 `task.Task` 转换成 `taskResponse`？

### 回答

> `task.Task` 是领域模型，`taskResponse` 是 HTTP 响应模型。两者分开以后，Handler 可以只公开 OpenAPI 允许的字段，并通过 JSON tag 控制字段名。`OwnerID` 留在领域模型中，不会出现在响应里。内部模型以后增加字段时，API 也不会跟着暴露新数据。

```go
type taskResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	Labels    []string  `json:"labels"`
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
```

### 子问题 1：为什么不能直接执行 `c.JSON(http.StatusOK, tasks)`？

#### 回答

> 直接序列化领域模型会把 HTTP 契约绑在内部结构上。当前 `Task` 还包含 `OwnerID`，直接返回会暴露这个字段。Go 默认生成的 JSON 字段名也不一定符合 OpenAPI 规定的 `createdAt`、`updatedAt` 等名称。

### 子问题 2：为什么响应字段可以继续使用 `time.Time`？

#### 回答

> `time.Time` 实现了 JSON 编码接口。`c.JSON` 最终使用标准库编码它，并输出符合 RFC 3339 的时间字符串。`time.Time.String()` 更适合日志和调试，它生成的文本不应作为 API 时间契约。

### 子问题 3：为什么需要处理 `nil` labels？

#### 回答

> Go 会把 `nil` slice 编码成 `null`，把非 nil 的空 slice 编码成 `[]`。OpenAPI 把 `labels` 定义为数组，所以没有标签时也应返回 `[]`。

```go
Labels: append(make([]string, 0, len(t.Labels)), t.Labels...),
```

> 这行代码创建一个长度为 0 的非 nil slice，再复制原标签。`t.Labels` 为 nil 时，结果仍然会编码成 `[]`。

### 子问题 4：显式转换有什么代价？

#### 回答

> 开发者需要维护字段映射。OpenAPI 新增必填字段后，如果 Handler 忘记补充映射，代码仍可能编译通过。HTTP 响应测试需要使用非零测试数据，并逐项检查契约字段。

## 问题 4：`make([]taskResponse, 0, len(tasks))` 表示什么？

### 回答

> 这行代码创建一个长度为 0、容量为 `len(tasks)` 的 slice。长度表示当前可以访问的元素数量，容量表示底层数组在重新分配前还能容纳多少元素。循环每调用一次 `append`，长度就增加 1。提前设置容量可以减少扩容次数。

```go
items := make([]taskResponse, 0, len(tasks))
for _, t := range tasks {
	items = append(items, taskResponse{
		ID: t.ID,
	})
}
```

### 子问题 1：容量是 slice 能保存的最大元素数吗？

#### 回答

> 不是。`append` 可以让长度超过原容量。Go 会申请更大的底层数组，把原有元素复制过去，再继续追加。容量只表示当前底层数组在重新分配前能够容纳多少元素。

### 子问题 2：如果改成 `make([]taskResponse, len(tasks))` 后继续使用 `append` 会怎样？

#### 回答

> 创建时 slice 已经包含 `len(tasks)` 个零值元素。循环又追加 `len(tasks)` 个转换结果，最终长度会变成两倍，前半部分都是空对象。

### 子问题 3：怎样使用指定长度的 slice？

#### 回答

> 创建指定长度的 slice 后，应通过索引写入，而不是继续追加。

```go
items := make([]taskResponse, len(tasks))
for i, t := range tasks {
	items[i] = taskResponse{
		ID: t.ID,
	}
}
```

### 子问题 4：预分配容量会改变接口行为吗？

#### 回答

> 正确使用时不会。它减少了底层数组扩容，返回的 JSON 应保持一致。代码如果混淆长度和容量，就会产生额外的零值元素，接口行为也会改变。

## 问题 5：怎样证明 `GET /v1/tasks` 已经正确实现？

### 回答

> 测试需要通过真实 Gin Router 发送请求，确认路由注册、Middleware 和 Handler 能连在一起。Repository 可以使用 stub，因为任务 02 验证的是 HTTP 调用链，不是数据库。
>
> 非空列表测试要检查状态码、JSON 元素数量和所有契约字段，还要检查 Repository 的调用次数与 ownerID。空列表测试要确认 `items` 是 `[]`，不能是 `null`。这些断言一起证明请求走过了正确的路由，参数传递正确，Handler 也完成了响应映射。

### 子问题 1：只断言状态码 `200` 有什么问题？

#### 回答

> Handler 即使永远返回 `{"items":[]}`，状态码测试也会通过。这个测试只能证明请求得到了 `200`，无法证明 Repository 返回的任务进入了响应体。

### 子问题 2：为什么测试通过 Router，而不直接调用 Handler？

#### 回答

> 直接调用 Handler 只能检查处理函数。通过 Router 发送请求还能检查 HTTP 方法、路径、Middleware 注册和 Handler 绑定。路由误写成 `/tasks` 时，请求 `/v1/tasks` 会得到 `404`，测试可以发现这个错误。

### 子问题 3：为什么要检查 Repository 的调用次数？

#### 回答

> 响应内容可以被 Handler 写死。断言 `callCount == 1` 可以证明 Handler 通过 Service 调用了 Repository，而且没有重复查询。

### 子问题 4：Service 返回错误后，Handler 为什么需要立即 `return`？

#### 回答

> Handler 写入错误状态后如果继续执行，还会尝试写入成功 JSON。一次请求会出现两套互相冲突的响应逻辑。`return` 让错误路径在写入错误响应后结束。任务 03 再定义不同错误对应的状态码和错误响应体。
