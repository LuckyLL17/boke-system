## Bug 是什么

管理员审批节目评论后，节目评论的聚合数量会超过实际已批准的可见评论数量，并且重复操作会继续放大这个偏差。

## 如何触发

对同一条节目评论执行审批，再重复执行审批操作并查看节目评论数量。

## 根因

文件：internal/handler/stats_handler.go 的 StatsHandler.ApproveComment 在同一请求中调用两次 CommentService.Approve。文件：internal/service/comment_service.go 的 CommentService.Approve 在一次审批中两次调用 IncrementCommentCount，且只根据 UpdateStatus 的 error 判断是否递增。文件：internal/repository/comment_repo.go 的 UpdateStatus 对已批准评论的零行更新仍返回 nil，因此重复审批仍会让反范式计数增加。该状态污染使聚合计数与 comments 表中已批准评论的真实行数失去一致性。

## 运行指令

```bash
go test ./internal/handler -run ^TestCommentApprovalKeepsEpisodeCountConsistent$ -count=1 -v
```

## 错误信息

已批准评论数量为 4，预期为 1。

## 错误堆栈

```text
=== RUN   TestCommentApprovalKeepsEpisodeCountConsistent
--- FAIL: TestCommentApprovalKeepsEpisodeCountConsistent (0.15s)
panic: approved comment count is 4, want 1 [recovered, repanicked]

goroutine 23 [running]:
testing.tRunner.func1.2(...)
\t/opt/homebrew/Cellar/go/1.27.0/libexec/src/testing/testing.go:2123
panic(...)
\t/opt/homebrew/Cellar/go/1.27.0/libexec/src/runtime/panic.go:859
FAIL\tpodcast-platform/internal/handler
```
