// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package memory reports RAM information for the jarvisfriends devices-*
// family: capacities and DIMM inventory as static values, usage and swap as
// dynamic ones. Works without cgo on Windows and Linux.
//
// Sources: gopsutil (usage, swap), ghw (physical/usable totals), and on
// Windows the Win32_PhysicalMemory WMI class for the per-DIMM inventory.
// Linux exposes DIMM details only through DMI tables that need root, so at
// Detailed an unprivileged Linux run reports that as a gap rather than
// pretending the machine has no modules.
package memory

import (
	"context"
	"time"

	"github.com/jaypipes/ghw"
	gmem "github.com/shirou/gopsutil/v4/mem"

	"github.com/jarvisfriends/devices-common/devices"
)

// Collector implements devices.Collector for the memory category.
type Collector struct{}

// New builds the memory collector.
func New() *Collector { return &Collector{} }

// Category satisfies devices.Collector.
func (*Collector) Category() string { return "memory" }

// Static satisfies devices.Collector.
func (*Collector) Static(ctx context.Context, level devices.Level) (devices.Snapshot, error) {
	snap := devices.Snapshot{Taken: time.Now()}

	if m, err := ghw.Memory(ghw.WithDisableWarnings()); err != nil {
		snap.Gaps = append(snap.Gaps, devices.Gap{Field: "memory.total", Reason: err.Error()})
	} else {
		if m.TotalPhysicalBytes > 0 {
			snap.Fields = append(snap.Fields,
				devices.Bytes("total_physical", "Physical", uint64(m.TotalPhysicalBytes), devices.Static, devices.Basic))
		}
		if m.TotalUsableBytes > 0 {
			snap.Fields = append(snap.Fields,
				devices.Bytes("total_usable", "Usable", uint64(m.TotalUsableBytes), devices.Static, devices.Basic))
		}
	}

	if swap, err := gmem.SwapMemoryWithContext(ctx); err == nil && swap.Total > 0 {
		snap.Fields = append(snap.Fields,
			devices.Bytes("swap_total", "Swap total", swap.Total, devices.Static, devices.Basic))
	}

	if level >= devices.Detailed {
		mods, gaps := collectModules(ctx, level)
		snap.Devices = append(snap.Devices, mods...)
		snap.Gaps = append(snap.Gaps, gaps...)
	}
	return snap, nil
}

// Dynamic satisfies devices.Collector.
func (*Collector) Dynamic(ctx context.Context, level devices.Level) (devices.Snapshot, error) {
	snap := devices.Snapshot{Taken: time.Now()}

	vm, err := gmem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return snap, err
	}
	snap.Fields = append(snap.Fields,
		devices.Bytes("used", "Used", vm.Used, devices.Dynamic, devices.Basic),
		devices.Bytes("available", "Available", vm.Available, devices.Dynamic, devices.Basic),
		devices.Pct("used_percent", "Used %", float64(int(vm.UsedPercent*10))/10, devices.Dynamic, devices.Basic),
	)

	if swap, err := gmem.SwapMemoryWithContext(ctx); err == nil && swap.Total > 0 {
		snap.Fields = append(snap.Fields,
			devices.Bytes("swap_used", "Swap used", swap.Used, devices.Dynamic, devices.Basic),
			devices.Pct("swap_used_percent", "Swap used %", float64(int(swap.UsedPercent*10))/10, devices.Dynamic, devices.Basic),
		)
	}

	if level >= devices.Detailed {
		// Linux exposes the cache picture; the fields are zero elsewhere and
		// zero fields would only be noise, so include what's real.
		for _, f := range []struct {
			key, name string
			v         uint64
		}{
			{"cached", "Cached", vm.Cached},
			{"buffers", "Buffers", vm.Buffers},
			{"dirty", "Dirty", vm.Dirty},
			{"committed", "Committed", vm.CommittedAS},
		} {
			if f.v > 0 {
				snap.Fields = append(snap.Fields,
					devices.Bytes(f.key, f.name, f.v, devices.Dynamic, devices.Detailed))
			}
		}
	}
	return snap, nil
}
