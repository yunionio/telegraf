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
    - `memory_total`
    - `memory_used`
    - `memory_free`
    - `memory_gtt_total`
    - `memory_gtt_used`
    - `memory_gtt_free`
    - `memory_vis_vram_total`
    - `memory_vis_vram_used`
    - `memory_vis_vram_free`

> [!NOTE]
> Memory fields are collected from `hy-smi --showmeminfo all`. If the command is
> not supported by the installed `hy-smi` version, the plugin falls back to
> outputting only the original fields above.

## Example Output

```text
hysmi,hcu=0,mode=normal,perf_mode=auto temperature_gpu=42,power_draw=35,power_cap=300,utilization_gpu=0,utilization_memory=1,utilization_encoder=0,utilization_decoder=0,memory_total=65520i,memory_used=2i,memory_free=65518i,memory_gtt_total=515905i,memory_gtt_used=8i,memory_gtt_free=515897i,memory_vis_vram_total=65520i,memory_vis_vram_used=2i,memory_vis_vram_free=65518i 1523991122000000000
```
