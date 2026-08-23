# 独立听众统计异常

## Bug 是什么
重复访问者存在时，按天的独立听众曲线被播放次数替代，页面和导出数据因此失去指标含义。

## 如何触发
在同一天写入同一访问者的多条播放记录，再读取统计看板或导出结果，比较每日独立听众和播放次数。该故障通过受保护验证用例稳定复现。

## 根因
`internal/repository/playback_repo.go` 提供总播放和独立听众聚合；`internal/service/stats_service.go` 组装看板与导出数据；`internal/handler/stats_handler.go` 返回每日曲线。服务层构造 `ListenersByDate` 时读取了播放次数聚合，处理器还会用播放曲线填补零值，跨层字段选择把独立听众指标替换成了播放指标。

## 运行指令
```text
go test ./internal/service -run ^TestDailyListenersRemainDistinctFromPlayCount$ -count=1 -v
```

## 错误信息
## 错误堆栈
```text
panic: daily listeners were replaced by play count
goroutine 22 [running]:
testing.tRunner.func1.2
testing.go:2123
panic.go:859
FAIL podcast-platform/internal/service 0.857s
```
