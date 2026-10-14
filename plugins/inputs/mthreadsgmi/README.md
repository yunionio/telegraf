# Moore Threads System Management Interface (mthreads-gmi) Input Plugin

This plugin collects metrics for Moore Threads GPUs from `mthreads-gmi -q --json`.

Only GPUs whose `DeviceType` is `Physical` are reported. MPC virtual instances
are skipped. Fields whose value is `N/A` are omitted.

> [!IMPORTANT]
> This plugin requires the `mthreads-gmi` binary on the system.

## Configuration

```toml @sample.conf
# Pulls statistics from Moore Threads GPUs attached to the host
[[inputs.mthreadsgmi]]
  ## Path to mthreads-gmi binary
  # bin_path = "/usr/bin/mthreads-gmi"
  ## Timeout for GPU polling
  # timeout = "5s"
```

Tags: `index`, `name`, `uuid`, `pci_bus_id`, `pstate`.

Fields: `utilization_gpu`, `utilization_memory`, `utilization_encoder`,
`utilization_decoder`, `temperature_gpu`, `memory_total`, `memory_used`,
`memory_free`, `power_draw`, `power_limit`, `clocks_current_graphics`,
`clocks_current_memory`, `pcie_link_gen_current`, `pcie_link_width_current`.
Memory sizes are MiB. Clocks are MHz. Power is watts. Link width is the lane
count (`16x` is stored as `16`).
