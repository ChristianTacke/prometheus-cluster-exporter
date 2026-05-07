# Metrics Reference

All metrics are prefixed with `cluster_`.

## Exporter Metrics

These metrics are always exported.

| Metric | Type | Labels | Description |
|---|---|---|---|
| `cluster_exporter_scrape_ok` | Gauge | — | `1` if scrape succeeded, `0` if skipped or failed |
| `cluster_exporter_stage_execution_seconds` | Gauge | `name` | Wall-clock duration of a specific exporter stage |

### Stage Names

The `name` label on `cluster_exporter_stage_execution_seconds` takes the following values:

| Value | Description |
|---|---|
| `retrieve_running_jobs` | Time spent running `squeue` |
| `retrieve_user_name_info` | Time spent running `getent passwd` |
| `retrieve_group_name_info` | Time spent running `getent group` |
| `build_metadata_metrics` | Time spent querying and processing metadata operations |
| `build_read_throughput_metrics` | Time spent querying and processing read throughput |
| `build_write_throughput_metrics` | Time spent querying and processing write throughput |

## SLURM Job Metrics

Emitted for Lustre job IDs that are plain integers (i.e. SLURM job IDs).

| Metric | Type | Labels | Description |
|---|---|---|---|
| `cluster_job_metadata_operations` | Gauge | `account`, `user`, `target` | Metadata operations per SLURM account and user on a MDT |
| `cluster_job_read_throughput_bytes` | Gauge | `account`, `user` | Read throughput per SLURM account and user (bytes/s) |
| `cluster_job_write_throughput_bytes` | Gauge | `account`, `user` | Write throughput per SLURM account and user (bytes/s) |

## Process Name Metrics

Emitted for Lustre job IDs in the `procname.uid` format (non-SLURM processes).
The UID is resolved via `getent passwd` / `getent group` to obtain user and group names.

| Metric | Type | Labels | Description |
|---|---|---|---|
| `cluster_proc_metadata_operations` | Gauge | `proc_name`, `group_name`, `user_name`, `target` | Metadata operations per process, group, and user on a MDT |
| `cluster_proc_read_throughput_bytes` | Gauge | `proc_name`, `group_name`, `user_name` | Read throughput per process, group, and user (bytes/s) |
| `cluster_proc_write_throughput_bytes` | Gauge | `proc_name`, `group_name`, `user_name` | Write throughput per process, group, and user (bytes/s) |

> **Note:** Metadata metrics (`*_metadata_operations`) include a `target` label and are limited to MDT targets
> matching the pattern `^.*-MDT[[:xdigit:]]{4}$` (e.g. `lustre-MDT0000`).
> Throughput metrics do not carry a `target` label.
