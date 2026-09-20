# Iluvatar System Management Interface (ixsmi) Input Plugin

This plugin collects metrics for Iluvatar GPUs including memory and GPU usage,
temperature and power, using the `ixsmi -q -x` XML output.

> [!IMPORTANT]
> This plugin requires the `ixsmi` binary and CoreX libraries on the system.
> Gather runs `env LD_LIBRARY_PATH=<lib_path> <bin_path> -q -x`.

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
# Pulls statistics from Iluvatar GPUs attached to the host
[[inputs.ixsmi]]
  ## Path to ixsmi binary
  # bin_path = "/usr/local/corex-4.4.0/bin/ixsmi"
  ## CoreX lib64 for LD_LIBRARY_PATH. Empty: derived from bin_path (.../bin/ixsmi -> .../lib64)
  # lib_path = "/usr/local/corex-4.4.0/lib64"
  ## Timeout for GPU polling
  # timeout = "5s"
```

`lib_path` sets `LD_LIBRARY_PATH` for the `ixsmi` process. If empty, it is
derived from `bin_path` (`.../bin/ixsmi` → `.../lib64`).

The `index` tag is the XML `minor_number` (same as `/dev/iluvatar{N}`), not the
array order of GPUs. Fields reported as `N/A` are omitted.

## Metrics

- measurement: `ixsmi`
  - tags
    - `index` (`minor_number`)
    - `name` (`product_name`)
    - `uuid`
    - `pstate` (`performance_state`)
    - `compute_mode`
    - `pci_bus_id` (gpu `@id`)
  - fields
    - `memory_total` (integer, MiB)
    - `memory_used` (integer, MiB)
    - `memory_free` (integer, MiB)
    - `utilization_gpu` (integer, percentage)
    - `utilization_memory` (integer, percentage)
    - `utilization_encoder` (integer, percentage)
    - `utilization_decoder` (integer, percentage)
    - `temperature_gpu` (integer, degrees C)
    - `temperature_memory` (integer, degrees C)
    - `power_draw` (float, W; from `gpu_power_draw`)
    - `power_limit` (float, W; from `current_gpu_power_limit`)
    - `board_power_draw` (float, W)
    - `board_power_limit` (float, W)
    - `clocks_current_sm` (integer, MHz)
    - `clocks_current_memory` (integer, MHz)
    - `pcie_link_gen_current` (integer)
    - `pcie_link_width_current` (integer)
    - `ecc_errors_single_bit` (integer)
    - `ecc_errors_double_bit` (integer)
    - `fan_speed` (integer, omitted if `N/A`)
    - `driver_version` (string)
    - `cuda_version` (string)
    - `serial` (string)
    - `vbios_version` (string)
    - `current_ecc` (string)

## Example Output

```text
ixsmi,compute_mode=Default,index=0,name=Iluvatar\ BI-V150S,pci_bus_id=00000000:26:00.0,pstate=P0,uuid=GPU-5fe225c8-c806-50f5-9132-90c020e40005 clocks_current_memory=1600i,clocks_current_sm=500i,cuda_version="10.2",current_ecc="Enabled",driver_version="4.4.0",memory_free=32700i,memory_total=32768i,memory_used=68i,pcie_link_gen_current=4i,pcie_link_width_current=16i,power_draw=13,power_limit=205,temperature_gpu=35i,utilization_gpu=0i,utilization_memory=1i 1523991122000000000
```
