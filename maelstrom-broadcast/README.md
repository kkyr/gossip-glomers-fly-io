# Broadcast

## Challenge #3d: Efficient Broadcast, Part I

https://fly.io/dist-sys/3d/

Command:

```shell
./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100 --topology tree4
```

Results:

```shell
:servers {:send-count 40560,
         :recv-count 40560,
         :msg-count 40560,
         :msgs-per-op 23.485813}
 
:stable-latencies {0 0,
                   0.5 366,
                   0.95 491,
                   0.99 504,
                   1 512}
```

Verified correct results with `--nemesis partition`.