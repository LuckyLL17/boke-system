## Bug 是什么

刷新频道订阅源后，生成内容仍使用旧的站点地址，条目链接指向错误位置。

## 如何触发

在站点地址更新后刷新频道订阅源并检查生成的节目链接。

## 根因

刷新入口、订阅服务和 RSS 生成器之间复用了旧地址，未将当前站点配置传递到新内容。

## 运行指令

```bash
go test ./pkg/rss -run ^TestRSSRefreshUsesCurrentSiteURL$ -count=1 -v
```

## 错误信息

订阅源条目保留了旧站点地址。

## 错误堆栈

```text
--- FAIL: TestRSSRefreshUsesCurrentSiteURL (0.00s)
    task015_test.go:17: RSS feed used a stale site URL
```
