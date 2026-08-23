## Bug 是什么

创建含多个章节的节目时，节目记录成功保存，但章节时间轴会遗漏部分条目。

## 如何触发

提交带有多个章节的节目创建请求并查询持久化后的章节。

## 根因

节目创建、章节批量保存和后续查询的链路没有保证全部章节在同一创建结果中被持久化。

## 运行指令

```bash
go test ./internal/repository -run ^TestBatchCreatePersistsEveryChapter$ -count=1 -v
```

## 错误信息

查询到的章节数量少于提交数量。

## 错误堆栈

```text
--- FAIL: TestBatchCreatePersistsEveryChapter (0.00s)
    task012_test.go:30: chapter persistence was incomplete: got 2 want 3
```
