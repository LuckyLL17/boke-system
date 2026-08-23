# 统计任务取消后的缓存状态不一致

## Bug 是什么

夜间统计任务收到服务重启或超时取消后，仍可能继续把部分频道的统计结果写入缓存；历史汇总写入后日期也可能偏移到当前日期，导致看板刷新前后出现半批新数据、半批旧数据或历史数据落在错误日期。

## 如何触发

让统计聚合任务处理历史播放数据，并在服务重启或任务超时期间观察缓存写入；也可以直接执行下面的验证指令检查历史日期是否被错误写成当前日期。

## 运行指令

```text
go test ./internal/service -run ^TestDailyCacheUsesRequestedHistoricalDate$ -count=1 -v
```

## 错误信息

验证失败，统计缓存实际保存为当前日期，而调用方要求保存为前一天的历史日期。

## 错误堆栈

```text
=== RUN   TestDailyCacheUsesRequestedHistoricalDate

2026/08/23 20:44:24 /private/var/folders/jv/437xkwbd18g8bsws_8_qqdy00000gn/T/go-an-stability-r7ua0qiy/env/internal/repository/playback_repo.go:140 record not found
[0.059ms] [rows:0] SELECT * FROM `stats_caches` WHERE channel_id = 8 AND DATE(date) = DATE("2026-08-22 20:44:24.956") ORDER BY `stats_caches`.`id` LIMIT 1
    task020_test.go:33: daily cache used the wrong date: got=2026-08-23 20:44:24.957157 +0800 +0800 want=2026-08-22 20:44:24.956847 +0800 CST
--- FAIL: TestDailyCacheUsesRequestedHistoricalDate (0.00s)
FAIL
FAIL podcast-platform/internal/service 0.869s
```

## 根因

1、文件：`cmd/server/main.go`、`internal/worker/scheduler.go`、`internal/worker/jobs.go`、`internal/service/stats_service.go`、`internal/repository/playback_repo.go`。2、符号：服务退出处理、`Worker.Stop`、`safeRun`、`AggregateHandler.Run`、`StatsService.UpdateDailyCache`、`PlaybackRepository.SaveStatsCache` 和 `PlaybackRepository.GetStatsCache`。3、运行时失效机制：服务退出时传给 `Worker.Stop` 的上下文没有取消 `safeRun` 内部独立创建的任务上下文，调度器只停止接收新任务；即使任务上下文已经取消，聚合任务只在外层频道循环检查，内层统计行仍会调用统计服务。统计服务和仓储接口不接收上下文，且仓储关闭默认事务后逐条保存，因此取消期间已经写入的部分缓存会独立提交。与此同时，`UpdateDailyCache` 用 `utils.Now()` 填充缓存日期，却用历史日期查询已有记录，导致历史统计按当前日期保存、历史记录无法正确复用并产生状态污染。

## 调用链

服务退出处理先调用 `Worker.Stop`，`Worker.Stop` 只关闭调度器，不取消 `safeRun` 中由 `context.WithTimeout` 创建的任务上下文。任务进入 `AggregateHandler.Run` 后按频道查询历史汇总，在内层结果循环中调用 `StatsService.UpdateDailyCache`。该方法忽略传入的历史日期并写入当前时间，随后通过 `PlaybackRepository.GetStatsCache` 按历史日期查找，再由 `PlaybackRepository.SaveStatsCache` 使用非事务会话保存。这样取消时可能留下部分已提交缓存，日期错位则让后续查询和更新继续偏离目标历史日期。
