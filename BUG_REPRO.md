## Bug 是什么

活跃订阅来源统计把没有来源标识的记录也分成一个空来源，导致来源分布与实际营销渠道不一致。

## 如何触发

创建一条带来源的活跃订阅和一条来源为空的活跃订阅，再查询频道来源分布。

## 根因

internal/handler/channel_handler.go 将订阅请求交给 internal/service/subscriber_service.go，服务层维护退订和重新订阅状态，internal/repository/subscriber_repo.go 的 GroupBySource 直接按 source 分组且只过滤 status=1，没有排除空来源；同时重新订阅会重置时间并把来源改成固定值，导致生命周期数据和来源统计发生跨层数据流失真。

## 运行指令

```bash
go test ./internal/repository -run ^TestSourceDistributionOmitsUnattributedActiveRows$ -count=1 -v
```

## 错误信息

来源分布结果包含不应出现的空来源分组。

## 错误堆栈

```text
--- FAIL: TestSourceDistributionOmitsUnattributedActiveRows (0.00s)
    task019_test.go:34: unattributed active subscriber leaked into source distribution: []map[string]interface {}{map[string]interface {}{"count":1, "source":"campaign"}, map[string]interface {}{"count":1, "source":""}}
```
