## Bug 是什么

个人资料持久化失败时，更新接口可能直接崩溃或返回不稳定错误，调用方无法按统一方式处理失败。

## 如何触发

让昵称或头像更新的数据库写入返回错误，再观察个人资料接口是否返回稳定的失败响应而不是发生 panic。

## 根因

文件：internal/handler/auth_handler.go、internal/service/auth_service.go、internal/repository/user_repo.go。符号：AuthHandler.UpdateProfile、AuthService.UpdateProfile、UserRepository.Update。仓储层写入错误需要沿服务层传递，处理层还需要区分业务错误与原始数据库错误；原实现对错误类型转换结果处理不完整，失败路径可能解引用空错误对象。该跨层错误传播失效会把持久化异常升级为运行时崩溃或不一致响应。

## 运行指令

```bash
go test ./internal/handler -run ^TestProfileUpdateReturnsPersistenceFailure$ -count=1 -v
```

## 错误信息

资料持久化失败路径触发 nil 指针崩溃。

## 错误堆栈

```text
=== RUN   TestProfileUpdateReturnsPersistenceFailure
task030_test.go:47: profile update panicked on persistence failure
--- FAIL: TestProfileUpdateReturnsPersistenceFailure
FAIL podcast-platform/internal/handler
```
