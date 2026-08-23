# 订阅人数不一致

## Bug 是什么
重复订阅或重新激活同一邮箱时，活跃订阅记录与频道冗余订阅人数可能分叉。

## 如何触发
对同一频道使用同一邮箱重复订阅，或先取消再重新订阅，然后比较活跃记录统计和频道人数。该故障通过受保护验证用例稳定复现。

## 根因
`internal/service/subscriber_service.go` 决定已有记录是否产生计数副作用；`internal/repository/subscriber_repo.go` 查找逻辑同一订阅；`internal/repository/channel_repo.go` 执行人数增减。当前计数副作用与已有活跃记录的判定传播不一致，导致重复请求仍可能增加冗余人数，退订再订阅后偏差继续累积。

## 运行指令
```text
go test ./internal/service -run ^TestActiveSubscriptionDoesNotIncreaseChannelCount$ -count=1 -v
```

## 错误信息
## 错误堆栈
```text
panic: active subscription was treated as a new subscriber
goroutine 19 [running]:
testing.tRunner.func1.2
testing.go:2123
panic.go:859
FAIL podcast-platform/internal/service 1.192s
```
