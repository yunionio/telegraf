# T-Head PPU-SMI Input Plugin

This plugin collects metrics for T-Head (PingTouGe) PPUs including memory and
PPU usage, temperature, power and clocks, using `ppu-smi --query-ppu` CSV output.

> [!IMPORTANT]
> This plugin requires the `ppu-smi` binary and PPU SDK libraries on the system.
> Gather runs
> `env LD_LIBRARY_PATH=<lib_path> <bin_path> --query-ppu=... --format=csv,noheader,nounits`.
> Default `bin_path` is `/usr/local/bin/ppu-smi`. `lib_path` defaults to
> `/usr/local/PPU_SDK/lib64` and is **not** derived from `bin_path`.

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
# Pulls statistics from T-Head PPUs attached to the host
[[inputs.ppusmi]]
  ## Path to ppu-smi binary
  # bin_path = "/usr/local/bin/ppu-smi"
  ## PPU SDK library dir for LD_LIBRARY_PATH. Do not derive from bin_path
  ## (/usr/local/bin/ppu-smi would incorrectly yield /usr/local/lib64).
  # lib_path = "/usr/local/PPU_SDK/lib64"
  ## Timeout for PPU polling
  # timeout = "5s"
```

`lib_path` sets `LD_LIBRARY_PATH` for the `ppu-smi` process. If empty, it
defaults to `/usr/local/PPU_SDK/lib64`.

The `index` tag is the `ppu-smi` enumeration index (same as `/dev/alixpu_ppu{N}`),
not PCI minor number. Fields reported as `N/A` are omitted.

If nvidia-smi-compatible query properties are rejected, the plugin retries with
PPU aliases (`utilization.ppu`, `temperature.ppu`, `clocks.current.cu`).

## Metrics

- measurement: `ppusmi`
  - tags
    - `index` (enumeration index)
    - `name`
    - `uuid`
    - `compute_mode`
    - `pci_bus_id`
  - fields
    - `memory_total` (integer, MiB)
    - `memory_used` (integer, MiB)
    - `memory_free` (integer, MiB)
    - `utilization_gpu` (integer, percentage; PPU utilization)
    - `utilization_memory` (integer, percentage)
    - `temperature_gpu` (integer, degrees C)
    - `temperature_memory` (integer, degrees C)
    - `power_draw` (float, W)
    - `clocks_current_sm` (integer, MHz; CU clocks)
    - `clocks_current_memory` (integer, MHz)

## Example Output

```text
ppusmi,compute_mode=Default,index=0,name=PPU-ZW810E,pci_bus_id=00000000:a8:00.0,uuid=GPU-019e2226-4211-0208-0000-000000ab261d clocks_current_memory=1800i,clocks_current_sm=200i,memory_free=98056i,memory_total=98304i,memory_used=248i,power_draw=51.99,temperature_gpu=42i,temperature_memory=42i,utilization_gpu=38i,utilization_memory=0i 1523991122000000000
```
