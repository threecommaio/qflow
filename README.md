# QFlow
Replicates traffic to various http endpoints backed by a durable queue.

# Usage
```
Replicates traffic to various http endpoints backed by a durable queue

This tool helps replicate to various http endpoints backed by a
durable disk queue in the event of failures or slowdowns.

Usage:
  qflow [flags]

Flags:
  -a, --addr string       listen addr (default ":8080")
  -c, --config string     config file for the clusters in yaml format
  -d, --data-dir string   data directory for storage
      --debug             enable debug logging
  -h, --help              help for qflow

$ qflow --help
$ qflow -c config.yml -d ./data
INFO[2018-08-15T15:07:11Z] registered (example1) with endpoints: [http://localhost:9090]
INFO[2018-08-15T15:07:11Z] config options: (http timeout: 10s, maxMsgSize: 10485760)
INFO[2018-08-15T15:07:11Z] listening on :8080
```