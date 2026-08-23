## Bug 是什么

节目评论提交成功后，按节目查询评论却找不到该记录，频道审核入口也可能看不到同一条评论。

## 如何触发

为节目创建一条评论，再使用节目评论查询入口读取该节目的评论列表。

## 根因

internal/handler/episode_handler.go 的评论请求负责传递节目上下文，internal/service/comment_service.go 的 Create 负责建立评论对象，internal/repository/comment_repo.go 的 ListByEpisode 负责按节目归属查询；其中归属字段在跨层传播后被错误地按频道标识参与筛选，导致已保存评论与节目查询条件不一致，属于跨层数据流问题。

## 运行指令

```bash
go test ./internal/repository -run ^TestEpisodeCommentsUseEpisodeAndNotChannelID$ -count=1 -v
```

## 错误信息

节目评论查询返回空结果，已提交评论不可见。

## 错误堆栈

```text
--- FAIL: TestEpisodeCommentsUseEpisodeAndNotChannelID (0.00s)
    task016_test.go:31: episode comment was hidden by wrong ownership filter: total=0 rows=0
```
