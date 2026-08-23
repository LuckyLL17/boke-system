## Bug 是什么

音频文件已经落盘但探测失败时，上传接口可能返回失败；分片合并还可能在文件不完整时报告成功并清理临时资源。

## 如何触发

让音频探测命令返回错误，或让分片合并过程中缺少一个分片，再观察最终文件、临时目录和接口结果是否一致。

## 根因

文件：internal/handler/episode_handler.go、internal/service/audio_service.go、pkg/ffmpeg/processor.go。符号：EpisodeHandler.UploadChunk、AudioService.UploadAudio、AudioService.MergeChunks、Processor.GetAudioInfo。最终文件写入后，上传服务没有在探测失败路径清理文件；探测组件又吞掉命令错误，分片合并则对缺失分片继续处理并提前删除临时目录。该资源生命周期和运行时错误传播失效使最终文件、临时分片和接口状态互相矛盾。

## 运行指令

```bash
go test ./pkg/ffmpeg -run ^TestAudioProbePreservesCommandFailure$ -count=1 -v
```

## 错误信息

音频探测命令失败时底层错误被隐藏。

## 错误堆栈

```text
=== RUN   TestAudioProbePreservesCommandFailure
task026_test.go:12: audio probe failure was hidden
--- FAIL: TestAudioProbePreservesCommandFailure
FAIL podcast-platform/pkg/ffmpeg
```
