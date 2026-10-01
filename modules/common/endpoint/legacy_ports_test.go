/*
Copyright 2026 Red Hat

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package endpoint

import "testing"

func TestLegacyPublicTLSPorts(t *testing.T) {
	expected := map[string]int32{
		"keystone":         13000,
		"placement":        13778,
		"nova":             13774,
		"glance":           13292,
		"cinder":           13776,
		"neutron":          13696,
		"heat-api":         13004,
		"heat-cfnapi":      13005,
		"swift":            13808,
		"barbican":         13311,
		"octavia":          13876,
		"manila":           13786,
		"ironic":           13385,
		"ironic-inspector": 13050,
		"designate":        13001,
		"aodh":             13042,
		"nova-novncproxy":  13080,
		"horizon":          443,
	}

	for name, port := range expected {
		if LegacyPublicTLSPorts[name] != port {
			t.Errorf("LegacyPublicTLSPorts[%q] = %d, want %d", name, LegacyPublicTLSPorts[name], port)
		}
	}

	if len(LegacyPublicTLSPorts) != len(expected) {
		t.Errorf("LegacyPublicTLSPorts has %d entries, expected %d", len(LegacyPublicTLSPorts), len(expected))
	}
}

func TestGetLegacyPublicTLSPort(t *testing.T) {
	port, ok := GetLegacyPublicTLSPort("placement")
	if !ok || port != 13778 {
		t.Errorf("GetLegacyPublicTLSPort(%q) = (%d, %v), want (13778, true)", "placement", port, ok)
	}

	if _, ok := GetLegacyPublicTLSPort("does-not-exist"); ok {
		t.Errorf("GetLegacyPublicTLSPort(%q) ok = true, want false", "does-not-exist")
	}
}
