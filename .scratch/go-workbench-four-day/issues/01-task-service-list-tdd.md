# 01 — 让 `task.Service.List` 测试通过

**你要完成的结果：** `Service.List` 接收 `ctx` 和 `ownerID` ，调用已有 Repository，并返回 Repository 的结果。

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

## 先理解 3 个词

* `Service`：放业务调用逻辑。当前任务中，它只负责调用 Repository。
* `Repository`：读取数据的接口。当前测试使用假的 Repository，还没有连接数据库。
* `ctx`：一次调用的上下文。它用于取消、超时和传递请求范围信息，必须继续传给 Repository。

## 第 1 步：确认测试现在会失败

在仓库根目录运行：

```bash
go test ./internal/task -run '^TestServiceListReturnsOwnerTasks$' -count=1
```

如果出现 Go 缓存权限错误，改用：

```bash
env GOCACHE=/private/tmp/go-workbench-go-build GOTMPDIR=/private/tmp \
  go test ./internal/task -run '^TestServiceListReturnsOwnerTasks$' -count=1
```

**怎么核对：** 输出中应包含 `List() error = not implemented` 和 `FAIL` 。

**怎么算通过：** 这一步的“通过”是测试按预期失败。若错误是编译失败、找不到文件或权限错误，则不算通过，先处理环境问题。

## 第 2 步：实现最小代码

打开 `internal/task/service.go` ，找到 `Service.List` 。

你只需要做三件事：

1. 给两个参数命名为 `ctx` 和 `ownerID`。
2. 调用 `s.repository.List(ctx, ownerID)`。
3. 把 Repository 返回的 Task 列表和 error 原样返回。

不要添加校验、日志、错误包装、Gin 或数据库代码。目标调用应是：

```go
s.repository.List(ctx, ownerID)
```

保存文件后运行格式化：

```bash
gofmt -w internal/task/service.go
```

**怎么核对：** 命令不应输出错误。再次打开文件，确认缩进由 `gofmt` 统一。

**怎么算通过：** `Service.List` 不再返回 `ErrNotImplemented` ，且没有修改其他函数。

## 第 3 步：让当前测试转绿

重新运行：

```bash
go test ./internal/task -run '^TestServiceListReturnsOwnerTasks$' -count=1
```

遇到缓存权限错误时使用第 1 步的 `env GOCACHE=...` 版本。

**怎么核对：** 输出最后应类似：

```text
ok   go-workbench/internal/task
```

**怎么算通过：** 命令退出码为 0，输出中没有 `FAIL` 、panic 或编译错误。

## 第 4 步：检查整个 Task 包

运行：

```bash
go test ./internal/task -count=1
go test -race ./internal/task -count=1
go vet ./internal/task
```

**怎么核对：** 两条 test 命令都显示 `ok` ； `go vet` 正常情况下没有输出并以 0 退出。

**怎么算通过：** 三条命令全部成功。任何一条失败都不能把本任务标为完成。

## 第 5 步：检查你只改了需要修改的地方

运行：

```bash
git diff -- internal/task/service.go internal/task/service_test.go
```

**怎么核对：** 应只看到 `Service.List` 的最小实现；当前阶段不应出现 Gin、PostgreSQL、日志或额外依赖。

**怎么算通过：** diff 与本任务直接相关，没有顺手重构其他代码。

## 最终通过清单

* [ ] 首次测试因 `not implemented` 失败。
* [ ] `Service.List` 把 `ctx` 和 `ownerID` 传给 Repository。
* [ ] 目标测试显示 `ok`。
* [ ] Task 包测试、race test 和 vet 全部通过。
* [ ] diff 只有必要修改。

完成后，把 `git diff` 和三条验证命令的完整输出发给 Agent。Agent 会审查你的实现，然后再由你补“ownerID 是否正确传递”和“Repository 错误是否原样返回”的测试。不要在本任务中提前进入 Gin handler。

## 你应该能回答

1. 为什么 `Service` 依赖 `Repository` 接口，而不是直接写 SQL？
2. 为什么要把调用者传入的 `ctx` 继续传下去？
3. 为什么当前测试通过仍不能证明 ownerID 一定传对了？
