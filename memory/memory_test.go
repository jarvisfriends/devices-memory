// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package memory

import (
	"context"
	"testing"

	"github.com/jarvisfriends/devices-common/devices"
)

func TestCategory(t *testing.T) {
	if got := New().Category(); got != "memory" {
		t.Fatalf("Category() = %q", got)
	}
}

func TestStaticHasTotals(t *testing.T) {
	snap, err := New().Static(context.Background(), devices.Basic)
	if err != nil {
		t.Fatal(err)
	}
	total, ok := snap.Field("total_physical")
	if !ok {
		total, ok = snap.Field("total_usable")
	}
	if !ok {
		t.Fatalf("no memory total reported; gaps: %+v", snap.Gaps)
	}
	if v, _ := total.Float(); v <= 0 {
		t.Errorf("memory total = %v", total.Value)
	}
}

func TestDynamicUsageSane(t *testing.T) {
	snap, err := New().Dynamic(context.Background(), devices.Basic)
	if err != nil {
		t.Fatal(err)
	}
	pct, ok := snap.Field("used_percent")
	if !ok {
		t.Fatal("no used_percent field")
	}
	if v, _ := pct.Float(); v <= 0 || v > 100 {
		t.Errorf("used_percent = %v, want (0,100]", v)
	}
}
