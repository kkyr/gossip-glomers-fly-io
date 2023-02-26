# Broadcast

## Challenge #4: Grow-Only Counter

https://fly.io/dist-sys/4/

Command:

```shell
./maelstrom test -w g-counter --bin ~/go/bin/maelstrom-counter --node-count 3 --rate 100 --time-limit 20 --nemesis partition
```

Results:

```shell
:workload {:valid? true,
        :errors nil,
        :final-reads (1106 1106 1106),
        :acceptable ([1106 1106])},
:valid? true}


Everything looks good! ヽ(‘ー`)ノ
```