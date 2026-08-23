# 分片上传合并异常

## Bug 是什么
乱序分片到达时，系统可能在所有分片写入前报告合并成功，且合并结果缺少上传内容。

## 如何触发
让同一上传任务的分片以乱序到达，并在最后编号分片先到时观察合并结果。该故障通过受保护验证用例稳定复现。

## 根因
`internal/handler/episode_handler.go` 传入分片总数和上传标识；`internal/service/audio_service.go` 负责检查全部分片并按索引合并；`pkg/utils/file.go` 提供分片路径约定。原实现以最后编号分片到达作为完成条件，且合并遍历使用了与文件命名不一致的索引。

## 运行指令
```text
go test ./internal/handler -run ^TestChunkedUploadMergesEveryChunkInArrivalOrder$ -count=1 -v
```

## 错误信息
## 错误堆栈
```text
=== RUN   TestChunkedUploadMergesEveryChunkInArrivalOrder
--- FAIL: TestChunkedUploadMergesEveryChunkInArrivalOrder
panic: merged audio did not preserve every uploaded chunk
goroutine 4 [running]:
testing.tRunner.func1.2
panic.go:859
FAIL podcast-platform/internal/handler 0.477s
```
