// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

//go:build windows

package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/yusufpapurcu/wmi"

	"github.com/jarvisfriends/devices-common/devices"
)

// win32PhysicalMemory is the WMI class shape for one installed DIMM.
type win32PhysicalMemory struct {
	BankLabel            string
	DeviceLocator        string
	Capacity             uint64
	Speed                uint32
	ConfiguredClockSpeed uint32
	Manufacturer         string
	PartNumber           string
	SerialNumber         string
	SMBIOSMemoryType     uint16
}

// smbiosMemoryTypes maps SMBIOS type 17 memory-type codes to names. Only the
// kinds seen in machines this decade; unknown codes print numerically.
var smbiosMemoryTypes = map[uint16]string{
	20: "DDR", 21: "DDR2", 24: "DDR3", 26: "DDR4", 30: "LPDDR4", 34: "DDR5", 35: "LPDDR5",
}

// collectModules inventories DIMMs via WMI. Serial numbers are Full-level:
// they identify the machine, so a Detailed report shouldn't leak them.
func collectModules(_ context.Context, level devices.Level) ([]devices.Device, []devices.Gap) {
	var dimms []win32PhysicalMemory
	if err := wmi.Query("SELECT BankLabel, DeviceLocator, Capacity, Speed, ConfiguredClockSpeed, Manufacturer, PartNumber, SerialNumber, SMBIOSMemoryType FROM Win32_PhysicalMemory", &dimms); err != nil {
		return nil, []devices.Gap{{Field: "modules", Reason: "Win32_PhysicalMemory query failed: " + err.Error()}}
	}

	out := make([]devices.Device, 0, len(dimms))
	for i, m := range dimms {
		id := strings.TrimSpace(m.DeviceLocator)
		if id == "" {
			id = fmt.Sprintf("DIMM%d", i)
		}
		d := devices.Device{ID: id}
		d.Fields = append(d.Fields,
			devices.Bytes("size", "Size", m.Capacity, devices.Static, devices.Detailed))
		if t, ok := smbiosMemoryTypes[m.SMBIOSMemoryType]; ok {
			d.Fields = append(d.Fields, devices.Str("type", "Type", t, devices.Static, devices.Detailed))
		} else if m.SMBIOSMemoryType != 0 {
			d.Fields = append(d.Fields,
				devices.Int("type", "Type", int64(m.SMBIOSMemoryType), "", devices.Static, devices.Detailed))
		}
		if m.Speed > 0 {
			d.Fields = append(d.Fields,
				devices.Int("speed", "Rated speed", int64(m.Speed), "MT/s", devices.Static, devices.Detailed))
		}
		if m.ConfiguredClockSpeed > 0 && m.ConfiguredClockSpeed != m.Speed {
			d.Fields = append(d.Fields,
				devices.Int("speed_configured", "Configured speed", int64(m.ConfiguredClockSpeed), "MT/s", devices.Static, devices.Detailed))
		}
		if v := strings.TrimSpace(m.Manufacturer); v != "" {
			d.Fields = append(d.Fields, devices.Str("vendor", "Vendor", v, devices.Static, devices.Detailed))
		}
		if v := strings.TrimSpace(m.PartNumber); v != "" {
			d.Fields = append(d.Fields, devices.Str("part_number", "Part number", v, devices.Static, devices.Detailed))
		}
		if v := strings.TrimSpace(m.BankLabel); v != "" {
			d.Fields = append(d.Fields, devices.Str("bank", "Bank", v, devices.Static, devices.Detailed))
		}
		if level >= devices.Full {
			if v := strings.TrimSpace(m.SerialNumber); v != "" {
				d.Fields = append(d.Fields, devices.Str("serial", "Serial", v, devices.Static, devices.Full))
			}
		}
		out = append(out, d)
	}
	return out, nil
}
