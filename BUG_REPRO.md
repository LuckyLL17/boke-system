## Bug 是什么

频道节目缓存失效后，较早的刷新结果仍可能覆盖较新的频道内容，订阅方会继续收到刷新前的节目列表。

## 如何触发

在频道节目发布和 RSS 缓存失效交错执行时，读取频道内容并使缓存失效。

## 根因

缓存写入没有把频道内容版本与失效操作统一协调，旧读取结果在失效后仍可写回缓存。

## 运行指令

```bash
go test ./internal/service -run ^TestConcurrentCacheInvalidationAdvancesChannelVersion$ -race -count=20 -v
```

## 错误信息

缓存失效后频道版本没有前进。

## 错误堆栈

```text
=== RUN   TestConcurrentCacheInvalidationAdvancesChannelVersion
--- FAIL: TestConcurrentCacheInvalidationAdvancesChannelVersion (0.00s)
panic: cache invalidation did not advance channel version for 6006 [recovered, repanicked]

goroutine 35 [running]:
testing.tRunner.func1.2(...)
\t/opt/homebrew/Cellar/go/1.27.0/libexec/src/testing/testing.go:2123
panic(...)
\t/opt/homebrew/Cellar/go/1.27.0/libexec/src/runtime/panic.go:859
FAIL\tpodcast-platform/internal/service
```
