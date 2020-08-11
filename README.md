# learning-Temporal

Web: http://localhost:8088/

### Workflow

1. Workflow 是由 `workflow worker` 來執行的
2. 如果中間 `workflow worker` 掛掉了，整個 workflow 會重跑，但 `activity` 會從 history 找之前的結果
3. 過程中如果有用到 zap 的 logger, 當重複執行他不會列印出來內容

### Activity

1. 每個activity 都需要是 `idempotent`
2. 如果再執行 activity 的過程中發生　`panic` 這些錯誤都將在主要的 workflow 裡面當成一般錯誤被攔截起來, 並不會直接 panic 掉然後造成workflow 無法成功執行下去
3. 

### Distributed CRON

1. 只支援到分鐘