# Vastai vasmi Input Plugin

This plugin collects metrics for Vastai GPUs using the `vasmi` binary.

> [!IMPORTANT]
> This plugin requires the `vasmi` binary to be installed on the system.

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
[[inputs.vasmi]]
  ## Optional: path to vasmi binary, defaults to $PATH via exec.LookPath
  # bin_path = "/usr/bin/vasmi"
  ## Optional: timeout for GPU polling
  # timeout = "5s"
```

## Metrics

- measurement: `vasmi`
  - tags
    - `aic`
    - `dev_id`
    - `die_id`
  - fields
    - `aic_power`
    - `temperature_gpu`
    - `utilization_gpu`
    - `utilization_memory`
    - `utilization_share_memory`
    - `utilization_encoder`
    - `utilization_decoder`
    - `utilization_ai`
    - `clocks_current_gpu`

## Example Output

```text
vasmi,aic=0,dev_id=0,die_id=0 temperature_gpu=41,utilization_gpu=0,utilization_memory=1,aic_power=25 1523991122000000000
```
