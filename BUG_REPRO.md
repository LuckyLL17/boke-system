## Bug 是什么

单集完成率接口使用了错误的播放数量，并把已经聚合的平均进度再次缩小，造成统计结果失真。

## 如何触发

准备多条播放记录，其中部分记录完成、部分未完成，再请求单集完成率统计并比较平均进度、完成率和总播放数。

## 根因

文件：internal/handler/stats_handler.go、internal/service/stats_service.go、internal/repository/playback_repo.go。符号：StatsHandler.EpisodeCompletion、StatsService.GetEpisodeCompletion、PlaybackRepository.CompletionRate。仓储层计算了真实总播放数却只返回完播数，服务层把平均进度除以完播数，处理层又把同一个错误值同时作为平均进度和完成率输出。该跨层数据流失效使总播放数与真实记录不一致，完播数越多时完成率越被错误缩小。

## 运行指令

```bash
go test ./internal/handler -run ^TestEpisodeCompletionReportsPlaybackStatistics$ -count=1 -v
```

## 错误信息

完成率和播放数量与播放记录不一致。

## 错误堆栈

```text
=== RUN   TestEpisodeCompletionReportsPlaybackStatistics
task027_test.go:64: completion statistics were inconsistent
--- FAIL: TestEpisodeCompletionReportsPlaybackStatistics
FAIL podcast-platform/internal/handler
```
