// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

//go:build !windows

package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/jaypipes/ghw"

	devices "github.com/jarvisfriends/devices-common"
)

// collectModules asks ghw for the DIMM inventory. On Linux the DMI tables
// behind it need root, so an unprivileged run usually gets nothing — that is
// reported as a gap with the actual fix (run as root) rather than silence.
func collectModules(_ context.Context, level devices.Level) ([]devices.Device, []devices.Gap) {
	m, err := ghw.Memory(ghw.WithDisableWarnings())
	if err != nil {
		return nil, []devices.Gap{{Field: "modules", Reason: err.Error()}}
	}
	if len(m.Modules) == 0 {
		return nil, []devices.Gap{{
			Field:  "modules",
			Reason: "DIMM inventory needs root to read DMI tables (try sudo, or dmidecode -t memory)",
		}}
	}
	out := make([]devices.Device, 0, len(m.Modules))
	for i, mod := range m.Modules {
		id := strings.TrimSpace(mod.Location)
		if id == "" {
			id = fmt.Sprintf("DIMM%d", i)
		}
		d := devices.Device{ID: id, Name: strings.TrimSpace(mod.Label)}
		if mod.SizeBytes > 0 {
			d.Fields = append(d.Fields,
				devices.Bytes("size", "Size", uint64(mod.SizeBytes), devices.Static, devices.Detailed))
		}
		if v := strings.TrimSpace(mod.Vendor); v != "" {
			d.Fields = append(d.Fields, devices.Str("vendor", "Vendor", v, devices.Static, devices.Detailed))
		}
		if level >= devices.Full {
			if v := strings.TrimSpace(mod.SerialNumber); v != "" {
				d.Fields = append(d.Fields, devices.Str("serial", "Serial", v, devices.Static, devices.Full))
			}
		}
		if len(d.Fields) > 0 {
			out = append(out, d)
		}
	}
	return out, nil
}
