Workers stop on shutdown without hanging on one that will not stop. Each starts through `goroutine.Run` and returns when its context is done, and shutdown cancels that context and waits with `waitgroup.SafeWaitContext` under a one-second deadline. The exporter ignores its context, so the wait returns `context.DeadlineExceeded` at the deadline and the program reports which workers stopped.

```sh
examples/stop-workers-at-a-deadline/run.sh
```

```text
shutdown: gave up at the deadline: context deadline exceeded
stopped: indexer, poller
still running: exporter
wvet: no findings
```
