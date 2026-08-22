# devices-memory

RAM information for the [devices-*](https://github.com/jarvisfriends/devices-common)
family: capacities and DIMM inventory as static values, usage and swap as
dynamic ones. Windows + Linux, no cgo required.

| | basic | detailed | full |
|---|---|---|---|
| **static** | physical/usable totals, swap total | per-DIMM size, type (DDR4/DDR5), speeds, vendor, part number (Windows: WMI; Linux: needs root) | DIMM serial numbers |
| **dynamic** | used, available, used %, swap | cached/buffers/dirty/committed where the OS reports them | — |

Sources: [gopsutil](https://github.com/shirou/gopsutil),
[ghw](https://github.com/jaypipes/ghw), and `Win32_PhysicalMemory` on
Windows. On Linux the DIMM inventory lives in root-only DMI tables; an
unprivileged run reports that as a gap rather than an empty machine.

```console
$ devices-memory
$ devices-memory --level detailed
$ devices-memory --every 0.5 --json
$ devices-memory tui
$ devices-memory web :8080
```

MIT — see [LICENSE](LICENSE).
