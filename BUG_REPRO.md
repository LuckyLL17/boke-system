## Bug 是什么

音频分片处理结束后，包含中间分片的临时目录仍残留在存储中。

## 如何触发

创建临时分片目录和文件后执行文件清理流程，再检查目录是否存在。

## 根因

上传处理、音频合并和文件工具对目录级资源的回收语义不一致，失败或完成后都可能遗留目录。

## 运行指令

```bash
go test ./pkg/utils -run ^TestRemoveFileRemovesTemporaryDirectory$ -count=1 -v
```

## 错误信息

清理函数执行后临时目录没有被删除。

## 错误堆栈

```text
--- FAIL: TestRemoveFileRemovesTemporaryDirectory (0.00s)
    task018_test.go:26: temporary directory was not removed: <nil>
```
