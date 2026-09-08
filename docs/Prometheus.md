# Prometheus

## Requirements

Neither of these is on by default.

```yaml
# The remote-write endpoint only exists with this flag.
--web.enable-remote-write-receiver

# Backdating with --at needs a window. Measured backwards from the newest
# sample, so a week of history costs a week-wide out-of-order window.
storage:
  tsdb:
    out_of_order_time_window: 168h
```

Without the flag the endpoint returns 404.
Without the window, Prometheus rejects any `--at` older than the newest sample as too-old.
