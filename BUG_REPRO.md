## Bug 是什么

注册请求遇到邮箱唯一性检查异常时，流程仍可能创建用户并返回成功，调用方无法获得一致的失败响应。

## 如何触发

让注册流程的邮箱检查查询返回存储错误，再观察注册接口是否停止后续创建并返回失败状态。

## 根因

文件：internal/handler/auth_handler.go、internal/service/auth_service.go、internal/repository/user_repo.go。符号：AuthHandler.Register、AuthService.Register、UserRepository.ExistsByEmail、UserRepository.Create。服务层原先把邮箱检查错误当成未发现重复邮箱继续执行，仓储层又缺少稳定的错误上下文，处理层因此可能把后续创建结果当成成功返回。该跨层错误传播失效使数据库检查失败被状态污染为注册成功。

## 运行指令

```bash
go test ./internal/handler -run ^TestRegistrationStopsWhenEmailCheckFails$ -count=1 -v
```

## 错误信息

邮箱检查失败后实际返回 HTTP 200，测试期望注册失败状态。

## 错误堆栈

```text
=== RUN   TestRegistrationStopsWhenEmailCheckFails
task021_test.go:51: expected registration failure status, got 200
--- FAIL: TestRegistrationStopsWhenEmailCheckFails
FAIL podcast-platform/internal/handler
```
