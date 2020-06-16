# learning-cadence



### Workflow

1. Workflow 是由 `workflow worker` 來執行的
2. 如果中間 `workflow worker` 掛掉了，整個 workflow 會重跑，但 `activity` 會從 history 找之前的結果
3. 過程中如果有用到 zap 的 logger, 當重複執行他不會列印出來內容

### Activity

1. 每個activity 都需要是 `idempotent`
2. 

### Distributed CRON

1. 只支援到分鐘