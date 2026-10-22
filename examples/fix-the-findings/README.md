The fix wvet points at, as one program: each job starts through `goroutine.Run`, which recovers a panic, and the wait is bounded by `waitgroup.SafeWait`. One job panics and the program still finishes. Since `goroutine.Run` calls its panic handler after the job's own deferred calls, the job calls `Done` at its end and the handler calls it after a panic.

```sh
examples/fix-the-findings/run.sh
```

```text
alpha: done
beta: recovered "beta failed"
gamma: done
wvet: no findings
```
