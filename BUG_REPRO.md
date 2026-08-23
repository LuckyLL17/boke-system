## Bug 是什么

凌晨补算前一天播放数据时，历史统计缓存被写成当天日期，导致缓存中的日期语义与实际聚合数据不一致。

## 如何触发

准备一条前一天的播放汇总并执行历史缓存更新，再读取该缓存记录的日期。

## 根因

internal/worker/jobs.go 的 AggregateHandler.Run 在得到前一天的汇总行后把非零日期路径的 cacheDate 设为 utils.Now；internal/service/stats_service.go 的 UpdateDailyCache 又用 utils.Now 填充 StatsCache.Date，忽略传入日期；internal/repository/playback_repo.go 的 GetStatsCache 通过日期查询已有缓存但没有唯一日期约束。结果是跨层日期传播被覆盖，历史数据落到当天并可能形成重复缓存行，属于状态一致性问题。

## 运行指令

```bash
go test ./internal/service -run ^TestHistoricalCacheUsesRequestedDate$ -count=1 -v
```

## 错误信息

历史缓存记录的日期与任务请求的前一天不一致。

## 错误堆栈

```text
--- FAIL: TestHistoricalCacheUsesRequestedDate (0.00s)
    task014_test.go:33: historical cache used the wrong date: got=2026-08-23 21:13:44 +0800 +0800 want=2026-08-22 21:13:44 +0800 CST
```
