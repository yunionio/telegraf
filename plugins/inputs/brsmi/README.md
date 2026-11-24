# BIREN brsmi Input Plugin

This plugin collects metrics for BIREN GPUs using the `brsmi` binary.

> [!IMPORTANT]
> This plugin requires the `brsmi` binary to be installed on the system.

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
## Path to brsmi executable.
bin_path = "/usr/bin/brsmi"

## Command timeout.
# timeout = "5s"
```

## Metrics

- measurement: `brsmi`
  - tags
    - `index`
  - fields
    - `power`
    - `temperature_gpu`
    - `temperature_memory`
    - `utilization_memory`
    - `utilization_spc`
    - `utilization_encoder`
    - `utilization_decoder`
    - `clocks_memory`
    - `clocks_gpu`

## Example Output

```text
brsmi,aic=0,dev_id=0,die_id=0 temperature_gpu=41,utilization_gpu=0,utilization_memory=1,aic_power=25 1523991122000000000
```
