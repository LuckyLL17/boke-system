## Bug 是什么

删除播客单集时，关联内容与主记录可能只完成部分清理，删除流程失败也会被接口伪装成成功。

## 如何触发

让单集查询或删除操作返回存储错误，再观察删除接口的响应状态和关联数据清理结果。

## 根因

文件：internal/handler/episode_handler.go、internal/service/episode_service.go、internal/repository/episode_repo.go。符号：EpisodeHandler.Delete、EpisodeService.Delete、EpisodeRepository.Delete、ChapterRepository.DeleteByEpisode。处理层在删除服务返回错误时仍返回成功响应；服务层分别执行章节和单集删除，没有统一事务，且主记录与关联资源的删除语义不同。该错误传播和状态一致性失效会造成删除失败被隐藏、章节与单集部分成功，以及音频、评论或播放记录残留。

## 运行指令

```bash
go test ./internal/handler -run ^TestEpisodeDeleteReportsServiceFailure$ -count=1 -v
```

## 错误信息

删除查询遇到存储错误时，接口仍报告删除成功。

## 错误堆栈

```text
=== RUN   TestEpisodeDeleteReportsServiceFailure
no such table: episodes
task022_test.go:35: delete failure was reported as success
--- FAIL: TestEpisodeDeleteReportsServiceFailure
FAIL podcast-platform/internal/handler
```
