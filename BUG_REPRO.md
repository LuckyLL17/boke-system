## Bug 是什么

单集发布状态切换后，状态、计划时间和实际发布时间可能互相矛盾，定时任务与订阅内容会采用不同判断。

## 如何触发

创建带计划时间的单集，将其转为已发布状态，再检查计划时间是否被错误覆盖或保留，并对比定时任务和订阅查询的判断条件。

## 根因

文件：internal/handler/episode_handler.go、internal/service/episode_service.go、internal/repository/episode_repo.go、internal/worker/publish.go。符号：EpisodeHandler.Update、EpisodeService.Update、EpisodeRepository.UpdateStatus、EpisodeRepository.Publish、PublishHandler.Run。status、scheduled_at 和 published_at 在多个入口分散更新且没有事务或跨字段约束；定时任务同时判断状态和计划时间，订阅查询却主要按已发布状态筛选。该状态污染和跨层数据流失效会让定时发布判断与订阅展示判断背离。

## 运行指令

```bash
go test ./internal/repository -run ^TestPublishStatusPreservesScheduledTime$ -count=1 -v
```

## 错误信息

发布状态切换覆盖了原有计划时间。

## 错误堆栈

```text
=== RUN   TestPublishStatusPreservesScheduledTime
task029_test.go:40: publishing changed the existing schedule time
--- FAIL: TestPublishStatusPreservesScheduledTime
FAIL podcast-platform/internal/repository
```
