## Bug 是什么

只修改频道标题时，已有简介或封面可能被空值覆盖，导致部分更新意外清除未修改资料。

## 如何触发

创建带简介和封面的频道，只提交新的标题，再读取频道资料并检查原有字段是否保留。

## 根因

文件：api/dto/channel_dto.go、internal/handler/channel_handler.go、internal/service/channel_service.go、internal/repository/channel_repo.go。符号：UpdateChannelRequest、ChannelHandler.Update、ChannelService.UpdateChannel、ChannelRepository.Update。请求层没有可靠区分缺省字段与显式空值，服务层合并部分更新时又把空字段直接写入持久化对象，存储层最终保存了被清空的字段。该跨层数据流和状态一致性失效使调用方只修改一个字段却污染其他资料。

## 运行指令

```bash
go test ./internal/handler -run ^TestPartialChannelUpdatePreservesExistingFields$ -count=1 -v
```

## 错误信息

部分更新后未提交修改的字段被清空。

## 错误堆栈

```text
=== RUN   TestPartialChannelUpdatePreservesExistingFields
task025_test.go:59: partial update cleared an untouched field
--- FAIL: TestPartialChannelUpdatePreservesExistingFields
FAIL podcast-platform/internal/handler
```
