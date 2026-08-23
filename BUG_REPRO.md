## Bug 是什么

重复装配限流组件会启动多条后台清理循环，后台 goroutine 持续增加并干扰限流状态。

## 如何触发

反复创建并使用限流中间件，观察清理循环启动前后的 goroutine 数量。

## 根因

路由装配、限流实例生命周期和清理任务启动没有共享唯一状态，重复初始化会累积资源。

## 运行指令

```bash
go test ./internal/middleware -run ^TestRateLimiterStartsOnlyOneCleanupLoop$ -race -count=20 -v
```

## 错误信息

清理 goroutine 数量在重复请求后超过预期。

## 错误堆栈

```text
--- FAIL: TestRateLimiterStartsOnlyOneCleanupLoop (0.02s)
    task017_test.go:34: rate limiter started too many cleanup goroutines: before=26 after=34
```
