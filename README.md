# learning-Temporal

Dashboard Web: http://localhost:8088/

## Question

1. 要如何處裡每一個 workflow 的版本差異?
2. 如何控制 workflow 的 history 的保存時間? (RETENTION PERIOD) , 目前看起來是依據namespace 下去設定
3. 每一個 workflow history 的最大可以保存的資料量是多少?

### Workflow

1. Workflow 是由 `workflow worker` 來執行的
2. 如果中間 `workflow worker` 掛掉了，整個 workflow 會重跑，但 `activity` 會從 history 找之前的結果
3. 過程中如果有用到 logger, 當重複執行他不會列印出來內容
4. Go SDK, goroutine 和 sleep 等一些 func 需要改用 workflow SDK 裡面的對應func
   https://docs.temporal.io/docs/go-create-workflows/#special-temporal-sdk-functions-and-types
   舉例: 假設一個場景，我們想讓某個 activity 暫時休息 3 秒, 如果使用 `time.Sleep`, 這樣其實 server 是不知道要 sleep, 所以導致 activity 還是會觸發actitity 的timeout 機制, 預設 10 秒, 正確的動作應該要用 `workflow.Sleep`
5. 如果 workflow 一直用 `for` 和 ‵sleep` 一直跑的話，這會造成 workflow history 變很大, 當達到 server 設定的 workflow history 上限後就會出問題喔 

### Activity

1. 每個activity 都需要是 `idempotent`
2. 如果再執行 activity 的過程中發生　`panic` 這些錯誤都將在主要的 workflow 裡面當成一般錯誤被攔截起來, 並不會直接 panic 掉然後造成workflow 無法成功執行下去
3. 

### Distributed CRON

1. 只支援到分鐘
2. cron 語法可以參考: https://crontab.guru/
3. 每個執行都是獨立一個 workflow instance, instance, 獨立的 runID
4. 透過 dashboard 把 cron workflow terminated 之後的排程都將不在執行