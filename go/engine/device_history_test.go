// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package engine

import (
	"testing"

	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
)

func TestDeviceHistoryBasic(t *testing.T) {
	tc := SetupEngineTest(t, "devhist")
	defer tc.Cleanup()

	CreateAndSignupFakeUser(tc, "dhst")

	ctx := &Context{}
	eng := NewDeviceHistorySelf(tc.G)
	if err := RunEngine(eng, ctx); err != nil {
		t.Fatal(err)
	}
	devs := eng.Devices()
	if len(devs) != 2 {
		t.Errorf("num devices: %d, expected 2", len(devs))
	}

	var desktop keybase1.DeviceDetail
	var paper keybase1.DeviceDetail

	for _, d := range devs {
		switch d.Device.Type {
		case libkb.DeviceTypePaper:
			paper = d
		case libkb.DeviceTypeDesktop:
			desktop = d
		default:
			t.Fatalf("unexpected device type %s", d.Device.Type)
		}
	}

	// paper's provisioner should be desktop
	if paper.Provisioner == nil {
		t.Fatal("paper device has no provisioner")
	}
	if paper.Provisioner.DeviceID != desktop.Device.DeviceID {
		t.Errorf("paper provisioned id: %s, expected %s", paper.Provisioner.DeviceID, desktop.Device.DeviceID)
		t.Logf("desktop: %+v", desktop)
		t.Logf("paper:   %+v", paper)
	}
}

func TestDeviceHistoryRevoked(t *testing.T) {
	tc := SetupEngineTest(t, "devhist")
	defer tc.Cleanup()

	u := CreateAndSignupFakeUser(tc, "dhst")

	ctx := &Context{}
	eng := NewDeviceHistorySelf(tc.G)
	if err := RunEngine(eng, ctx); err != nil {
		t.Fatal(err)
	}

	var desktop keybase1.DeviceDetail
	var paper keybase1.DeviceDetail

	for _, d := range eng.Devices() {
		switch d.Device.Type {
		case libkb.DeviceTypePaper:
			paper = d
		case libkb.DeviceTypeDesktop:
			desktop = d
		default:
			t.Fatalf("unexpected device type %s", d.Device.Type)
		}
	}

	// paper's provisioner should be desktop
	if paper.Provisioner == nil {
		t.Fatal("paper device has no provisioner")
	}
	if paper.Provisioner.DeviceID != desktop.Device.DeviceID {
		t.Errorf("paper provisioned id: %s, expected %s", paper.Provisioner.DeviceID, desktop.Device.DeviceID)
		t.Logf("desktop: %+v", desktop)
		t.Logf("paper:   %+v", paper)
	}

	// revoke the paper device
	ctx.SecretUI = u.NewSecretUI()
	ctx.LogUI = tc.G.UI.GetLogUI()
	reng := NewRevokeDeviceEngine(RevokeDeviceEngineArgs{ID: paper.Device.DeviceID}, tc.G)
	if err := RunEngine(reng, ctx); err != nil {
		t.Fatal(err)
	}

	// get history after revoke
	eng = NewDeviceHistorySelf(tc.G)
	if err := RunEngine(eng, ctx); err != nil {
		t.Fatal(err)
	}

	var desktop2 keybase1.DeviceDetail
	var paper2 keybase1.DeviceDetail

	for _, d := range eng.Devices() {
		switch d.Device.Type {
		case libkb.DeviceTypePaper:
			paper2 = d
		case libkb.DeviceTypeDesktop:
			desktop2 = d
		default:
			t.Fatalf("unexpected device type %s", d.Device.Type)
		}
	}

	// paper's provisioner should (still) be desktop
	if paper2.Provisioner == nil {
		t.Fatal("paper device has no provisioner")
	}
	if paper2.Provisioner.DeviceID != desktop2.Device.DeviceID {
		t.Errorf("paper provisioned id: %s, expected %s", paper2.Provisioner.DeviceID, desktop2.Device.DeviceID)
		t.Logf("desktop: %+v", desktop2)
		t.Logf("paper:   %+v", paper2)
	}

	if paper2.RevokedAt == nil {
		t.Fatal("paper device RevokedAt is nil")
	}
}
