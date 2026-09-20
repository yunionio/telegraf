# Radeontop Input Plugin

This plugin collects metrics for AMD GPUs using the `radeontop` binary.

> [!IMPORTANT]
> This plugin requires the `radeontop` binary to be installed on the system.

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
[[inputs.radeontop]]
  ## Optional: path to radeontop binary, defaults to $PATH via exec.LookPath
  # bin_path = "/usr/bin/radeontop"
  # device_paths = ["/dev/dri/renderD128", "/dev/dri/renderD129"]
  ## Optional: timeout for GPU polling
  # timeout = "5s"
```

## Metrics

- measurement: `radeontop`
  - tags
    - `device_path`
    - `bus`
  - fields
    - `memory_total` (integer, MiB)
    - `memory_used` (integer, MiB)
    - `memory_free` (integer, MiB)
    - `utilization_gpu` (float, percentage)
    - `utilization_memory` (float, percentage)
    - `clocks_current_memory` (float, MHz)
    - `clocks_current_shader` (float, MHz)

## Example Output

```text
radeontop,bus=0000:03:00.0,device_path=/dev/dri/renderD128 memory_total=8192,memory_used=1024,utilization_gpu=12.5 1523991122000000000
```
