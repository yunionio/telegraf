# Hygon hy-smi Input Plugin

This plugin collects metrics for Hygon DCUs using the `hy-smi` binary.

> [!IMPORTANT]
> This plugin requires the `hy-smi` binary to be installed on the system.

⭐ Telegraf v1.36.3
🏷️ system, hardware
💻 linux

## Global configuration options <!-- @/docs/includes/plugin_config.md -->

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration settings. These settings are used to
modify metrics, tags, and field or create aliases and configure ordering, etc.
See the [CONFIGURATION.md][CONFIGURATION.md] for more details.

[CONFIGURATION.md]: ../../../docs/CONFIGURATION.md#plugins

## Configuration

```toml @sample.conf
[[inputs.hysmi]]
  ## Optional: path to hy-smi binary, defaults to /opt/hyhal/bin/hy-smi
  # bin_path = "/opt/hyhal/bin/hy-smi"
  ## Optional: timeout for GPU polling
  # timeout = "5s"
```

## Metrics

- measurement: `hysmi`
  - tags
    - `hcu`
    - `perf_mode`
    - `mode`
  - fields
    - `temperature_gpu`
    - `power_draw`
    - `power_cap`
    - `utilization_gpu`
    - `utilization_memory`
    - `utilization_encoder`
    - `utilization_decoder`

## Example Output

```text
hysmi,hcu=0,mode=normal,perf_mode=auto temperature_gpu=42,power_draw=35,power_cap=300,utilization_gpu=0,utilization_memory=1,utilization_encoder=0,utilization_decoder=0 1523991122000000000
```
