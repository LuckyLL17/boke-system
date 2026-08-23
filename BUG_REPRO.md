## Bug 是什么

频道绑定自定义域名后，订阅链接生成流程仍可能使用平台地址，更新配置与读取链接的结果不一致。

## 如何触发

为频道设置自定义域名，再通过频道更新和订阅链接入口读取生成地址，检查所有链接是否使用同一域名。

## 根因

文件：internal/handler/channel_handler.go、internal/service/channel_service.go、internal/service/rss_service.go、cmd/server/main.go、web/channels.html、web/rss-view.html。符号：ChannelHandler.Update、ChannelService.UpdateChannel、RSSService.GetSubscribeLinks、频道页面链接生成逻辑。自定义域名从请求层进入后没有在所有服务和页面入口保持同一份配置，RSS 链接生成仍可能回退到平台基础地址。该跨层数据流失效使配置写入结果与对外订阅地址不一致。

## 运行指令

```bash
go test ./internal/handler -run ^TestCustomDomainReachesSubscribeLinks$ -count=1 -v
```

## 错误信息

订阅链接忽略了已配置的自定义域名。

## 错误堆栈

```text
=== RUN   TestCustomDomainReachesSubscribeLinks
task028_test.go:58: subscribe link ignored the configured custom domain
--- FAIL: TestCustomDomainReachesSubscribeLinks
FAIL podcast-platform/internal/handler
```
