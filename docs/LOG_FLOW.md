# Log Flow

本文说明运行日志从后端到前端的当前链路。

## 1. 后端写入日志

日志入口在 `core/runlog`：

- `core/runlog/logger.go`
  - `Logger.Println` / `Printf` / `PrintlnWithTransaction` 最终都会进入 `Logger.write`。
  - 每条日志会被格式化为 `[HH:mm:ss] message`。
  - `recordLogs=true` 时，日志会写入内存数组 `lines`，并调用 `OnLine` 回调。
  - `recordLogs=false` 时，仍会写到 stdout，但不会缓存，也不会触发前端事件。
- `core/runlog/runner.go`
  - `Runner.Run` 执行外部命令。
  - 启动命令前记录 `Running command` 和 `Working directory`。
  - stdout/stderr 每读到一行都会写入 `Logger`。
  - 命令退出非 0 时返回 `command failed: ...`。

打包相关日志在 `core/packaging/service.go` 中带 transaction 写入：

- `Package` 创建 `runlog.Transaction{ID, Type: "package", Title}`。
- 打包开始、失败、完成，以及构建命令输出，都会通过 `PrintlnWithTransaction` 携带 transaction 信息。

## 2. 后端发送 Wails 事件

事件桥接在 `desktop/run.go`：

- 启动时调用 `application.RegisterEvent[contracts.LogLineEvent]("manager:log-line")` 注册事件类型。
- 创建全局 `logSink := runlog.New()`。
- 设置 `logSink.OnLine` 回调，将 `runlog.Line` 转成 `contracts.LogLineEvent`。
- 通过 `app.Event.Emit("manager:log-line", payload)` 推送给前端。

事件 payload 定义在 `core/contracts/types.go`：

```go
type LogLineEvent struct {
    Line             string `json:"line"`
    TransactionID    string `json:"transactionId"`
    TransactionType  string `json:"transactionType"`
    TransactionTitle string `json:"transactionTitle"`
}
```

## 3. 前端接收事件

前端入口在 `frontend/src/App.vue`：

- `onMounted` 时调用 `Events.On("manager:log-line", handler)`。
- handler 读取 `event.data`，调用 `store.appendLogEntry(event.data)`。
- `onBeforeUnmount` 时调用返回的取消订阅函数。

## 4. 前端保存和分组

状态管理在 `frontend/src/store/index.js`：

- `logEntries` 保存最近的全局日志。
- `packageTransactions` 按 `transactionId` 保存打包过程日志。
- `appendLogEntry` 会把事件写入 `logEntries`。
- 当 `transactionType === "package"` 时，同步写入 `packageTransactions[transactionId].entries`。
- `trimLogs` 保留最近 500 条，超出的日志会从全局列表和对应 transaction 中移除。
- `clearLogLines` 清空全局日志和 transaction 分组。

## 5. 前端渲染

日志展示有两个入口：

- `frontend/src/components/Settings.vue`
  - 展示全局 `store.logEntries`。
  - 提供“跟随最新”开关，开启时每来一条日志自动滚动到底部。
  - 调用 `SettingsService.ClearLogs()` 后同步清空前端 store。
- `frontend/src/components/PackageRunCard.vue`
  - 展示当前 `transactionId` 对应的打包日志。
  - 打包弹窗打开或新增日志时自动滚动到底部。

日志颜色由 `frontend/src/utils/logs.js` 统一判断：

- error：包含 `error`、`failed`、`fatal`、`compile aborted`、`exit status` 等。
- warning：包含 `warn`、`missing`、`not found`、`skipped`、`disabled` 等。
- success：包含 `completed`、`generated`、`created`、`up to date` 等。
- command：包含 `Running command`、`Working directory`、`go build`、`wails3`、`ISCC` 等。
- package/info：兜底分类。

## 6. 设置开关

日志记录开关在 `core/settings/service.go`：

- 前端保存设置时调用 `SettingsService.SaveSettings`。
- 后端保存配置后调用 `Logger.SetRecordLogs(saved.RecordLogs)`。
- 关闭后，新日志不会进入内存缓存，也不会通过 `manager:log-line` 推给前端。
