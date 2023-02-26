# Broadcast

## Challenge #5: Single-Node Kafka-Style Log

https://fly.io/dist-sys/5a/

Command:

```shell
./maelstrom test -w kafka --bin ~/go/bin/maelstrom-kafka --node-count 1 --concurrency 2n --time-limit 20 --rate 1000
```

Results:

```shell
:workload {:valid? true,
        :worst-realtime-lag {:time 33.057959416,
                             :process 4,
                             :key "9",
                             :lag 32.89983675},
        :bad-error-types (),
        :error-types (),
        :info-txn-causes ()},
 :valid? true}
 
Everything looks good! ヽ(‘ー`)ノ
```