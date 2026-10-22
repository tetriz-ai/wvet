# Examples

| You want | Example |
|---|---|
| To run wvet over a module, as the Quick start does | [vet-a-module](vet-a-module/) |
| To silence one finding on purpose, and see where a directive counts | [suppress-a-finding](suppress-a-finding/) |
| To check files behind a build tag, the way CI builds them | [check-tagged-files](check-tagged-files/) |
| To start goroutines and wait for them the way wvet asks | [fix-the-findings](fix-the-findings/) |
| To stop those goroutines on shutdown without hanging on one that will not stop | [stop-workers-at-a-deadline](stop-workers-at-a-deadline/) |
| To write your own architecture rule as a Go test | [write-an-architecture-rule](write-an-architecture-rule/) |

Run them all from the repository root with `scripts/run-examples.sh`. It checks each script
with shellcheck, runs it, and compares its standard output with the last block in that
example's README.

Each script builds wvet from the clone and works offline, on a sample module copied into a
temporary directory or on this repository. The output shown writes that temporary directory as
`/home/you/project` and each test duration as `<time>`.
