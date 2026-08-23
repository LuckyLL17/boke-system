## Bug 是什么

频道审核的存储更新失败时，接口可能仍返回成功，频道状态也不会按调用方预期改变。

## 如何触发

让待审核频道的状态更新返回数据库错误，再观察审核接口是否返回失败以及错误是否沿调用链传递。

## 根因

文件：internal/handler/channel_handler.go、internal/service/channel_service.go、internal/repository/channel_repo.go、pkg/errors/errors.go。符号：ChannelHandler.Approve、ChannelService.Approve、ChannelRepository.UpdateStatusFrom、appErr.As。仓储层需要区分数据库错误与零行状态转换，服务层需要保留错误并表达状态不匹配，处理层需要兼容业务错误和原始错误。原实现吞掉服务层错误，第一次 Claude 处理还未覆盖原始错误转换，最终由第二次 Claude 补齐错误分支。该跨层错误传播失效会把审核失败转成成功或触发失败路径崩溃。

## 运行指令

```bash
go test ./internal/handler -run ^TestChannelApprovalPropagatesStorageFailure$ -count=1 -v
```

## 错误信息

数据库更新失败时，原始版本报告成功。

## 错误堆栈

```text
=== RUN   TestChannelApprovalPropagatesStorageFailure
no such table: channels
task023_test.go:31: approval storage failure was reported as success
--- FAIL: TestChannelApprovalPropagatesStorageFailure
FAIL podcast-platform/internal/handler
```
