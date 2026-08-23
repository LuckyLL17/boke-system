## Bug 是什么

并发播放请求的明细虽然写入成功，节目和频道累计播放数会丢失更新。

## 如何触发

同时提交多个播放启动请求并比较明细数量与两个累计计数。

## 根因

播放明细写入与节目、频道计数更新跨越多个存储操作，竞争条件使累计值无法与并发请求对应。

## 运行指令

```bash
go test ./internal/service -run ^TestConcurrentPlaybackCountersMatchDetails$ -race -count=20 -v
```

## 错误信息

并发执行后累计计数小于播放明细数量。

## 错误堆栈

```text
--- FAIL: TestConcurrentPlaybackCountersMatchDetails (0.01s)
    task013_test.go:18: concurrent counter lost an update: table=episodes value=1
```
