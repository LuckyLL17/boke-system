## Bug 是什么

编辑单集并提交空章节列表时，已有章节可能继续保留，章节替换结果与请求内容不一致。

## 如何触发

先为单集创建一个章节，再通过单集更新接口提交空章节列表，检查旧章节是否被清除。

## 根因

文件：internal/handler/episode_handler.go、internal/service/episode_service.go、internal/repository/chapter_repo.go、internal/repository/episode_repo.go。符号：EpisodeHandler.Update、EpisodeService.Update、ChapterRepository.ReplaceByEpisode、EpisodeRepository.Update。处理层、服务层和仓储层对空列表的含义没有保持一致，空集合可能被当成“未提交章节”而跳过替换；章节删除、插入和单集更新又依赖跨层事务协同。该状态污染和持久化一致性失效会使旧章节残留或在异常时留下不完整更新。

## 运行指令

```bash
go test ./internal/handler -run ^TestEpisodeUpdateClearsSubmittedChapters$ -count=1 -v
```

## 错误信息

提交空章节列表后旧章节仍存在。

## 错误堆栈

```text
=== RUN   TestEpisodeUpdateClearsSubmittedChapters
task024_test.go:66: empty chapter submission left old chapters
--- FAIL: TestEpisodeUpdateClearsSubmittedChapters
FAIL podcast-platform/internal/handler
```
