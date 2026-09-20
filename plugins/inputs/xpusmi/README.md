# Kunlunxin System Management Interface (xpusmi) Input Plugin

This plugin collects metrics for Kunlunxin XPUs (e.g. P800 OAM) including
memory and XPU usage, L3 usage, temperature, power and ECC errors, using the
`xpu-smi -q -x` XML output.

> [!IMPORTANT]
> This plugin requires the `xpu-smi` binary and the Kunlunxin driver libraries
> on the system. Gather runs `<bin_path> -q -x`, prefixed with
> `env LD_LIBRARY_PATH=<lib_path>` when `lib_path` is set.

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
# Pulls statistics from Kunlunxin XPUs attached to the host
[[inputs.xpusmi]]
  ## Path to xpu-smi binary
  # bin_path = "/usr/local/bin/xpu-smi"
  ## Kunlunxin driver library dir for LD_LIBRARY_PATH. Empty: inherit the
  ## environment of the telegraf process
  # lib_path = "/usr/local/xpu/so"
  ## Timeout for XPU polling
  # timeout = "5s"
```

`lib_path` sets `LD_LIBRARY_PATH` for the `xpu-smi` process. Leave it empty to
inherit the environment of the telegraf process.

The `index` tag is the XML `minor_number` (same as `/dev/xpu{N}`), not the array
order of XPUs. Fields reported as `N/A` or `Invalid Argument` are omitted.

## Metrics

- measurement: `xpusmi`
  - tags
    - `index` (`minor_number`)
    - `name` (`product_name`)
    - `uuid`
    - `pci_bus_id` (xpu `@id`)
    - `oam_id`
  - fields
    - `memory_total` (integer, MiB, from `fb_memory_usage`)
    - `memory_used` (integer, MiB)
    - `memory_free` (integer, MiB)
    - `l3_memory_total` (integer, MiB, from `l3_usage`)
    - `l3_memory_used` (integer, MiB)
    - `l3_memory_free` (integer, MiB)
    - `utilization_gpu` (integer, percentage, from `xpu_util`)
    - `temperature_gpu` (integer, degrees C)
    - `power_draw` (float, W)
    - `power_limit` (float, W, from `enforced_power_limit`)
    - `clocks_current_cluster` (integer, MHz, from `cluster_clock`)
    - `clocks_current_cdnn` (integer, MHz, from `cdnn_clock`)
    - `pcie_link_gen_current` (integer)
    - `pcie_link_width_current` (integer)
    - `ecc_errors_dram_correctable` (integer, volatile)
    - `ecc_errors_dram_uncorrectable` (integer, volatile)
    - `ecc_errors_dram_correctable_aggregate` (integer)
    - `ecc_errors_dram_uncorrectable_aggregate` (integer)
    - `driver_version` (string)
    - `cuda_version` (string)
    - `serial` (string)
    - `product_brand` (string)
    - `product_architecture` (string)
    - `current_ecc` (string)

## Example Output

```text
xpusmi,index=0,name=P800\ OAM,oam_id=3,pci_bus_id=00000000:03:00.0,uuid=GPU-a0ac042a-5a14-5481-bf7e-dd8e523792cd clocks_current_cdnn=1450i,clocks_current_cluster=1450i,cuda_version="10.2",current_ecc="Enabled",driver_version="5.19.0.0",ecc_errors_dram_correctable=0i,ecc_errors_dram_correctable_aggregate=0i,ecc_errors_dram_uncorrectable=0i,ecc_errors_dram_uncorrectable_aggregate=0i,l3_memory_free=96i,l3_memory_total=96i,l3_memory_used=0i,memory_free=98304i,memory_total=98304i,memory_used=0i,pcie_link_gen_current=4i,pcie_link_width_current=16i,power_draw=87,power_limit=400,product_architecture="KL3",product_brand="KUNLUNXIN",serial="02K15K6252V00115",temperature_gpu=36i,utilization_gpu=0i 1523991122000000000
```
