# Prometheus Cluster Exporter

A [Prometheus](https://prometheus.io/) exporter for Lustre metadata operations and IO throughput metrics 
associated to SLURM accounts and process names with user and group information on a cluster.

[Grafana dashboard](https://grafana.com/grafana/dashboards/14668) is also available.

## Building

`go build -o prometheus-cluster-exporter *.go`

## Requirements

### Lustre Exporter

[Lustre exporter](https://github.com/GSI-HPC/lustre_exporter) that exposes enabled Lustre Jobstats on the filesystem.

### Squeue Command

The squeue command from SLURM must be accessable locally to the exporter to retrieve the running jobs.  

For instance running the exporter on the SLURM controller is advisable, since the target host should be most stable for a productional environment.

### Getent

The getent command is required for the uid to user and group mapping used for the process names throughput metrics.

## Execution

### Parameter

| Name       | Default           | Description                                                                                                                        |
| ---------- | ----------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| version    | false             | Print version                                                                                                                      | 
| promserver | \-                | [REQUIRED] Prometheus Server to be used e.g. http://prometheus-server:9090                                                         |
| log        | ERROR             | Sets log level - ERROR, WARNING, INFO, DEBUG or TRACE                                                                              | 
| port       | 9846              | The port to listen on for HTTP requests                                                                                            |
| timeout    | 15                | HTTP request timeout in seconds for exporting Lustre Jobstats on Prometheus HTTP API                                               |
| timerange  | 1m                | Time range used for rate function on the retrieving Lustre metrics from Prometheus - A number (1–3 digits) with unit s, m, h or d  |

### Running in a Productive Environment

For a productive environment it is advisable to run the exporter on the SLURM controller,  
since the target host should be most stable.

### Prometheus Scrape Settings

Depending on the required resolution and runtime of the exporter,  
* the `scrape interval` should be set as appropriate e.g. at least 1 minute or higher.  
* the `scrape timeout` should be set close to the specified scrape interval.

## Metrics

See [docs/architecture.md](docs/architecture.md) for an internal overview and dataflow explanation.

Cluster exporter metrics are prefixed with `cluster_`.
See [docs/metrics.md](docs/metrics.md) for the full metrics reference including labels and descriptions.

In short, the exporter provides:

- **Exporter health** — `cluster_exporter_scrape_ok` and per-stage execution times.
- **SLURM job metrics** — metadata operations and read/write throughput per account and user.
- **Process name metrics** — the same, but for non-SLURM processes, resolved via UID to user/group names.

## Multiple Scrape Prevention

Since the forked processes do not have a timeout handling, they might block for a uncertain amount of time.  
It is very unlikely that reexecuting the processes will solve the problem of being blocked.
Therefore multiple scrapes at a time will be prevented by the exporter.  

The following warning will be displayed on afterward scrape executions, were a scrape is still active:  
    *"Collect is still active... - Skipping now"*

Besides that, the cluster\_exporter\_scrape\_ok metric will be set to 0 for skipped scrape attempts.  

