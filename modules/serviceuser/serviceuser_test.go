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

package serviceuser

import "testing"

func TestRegistryUIDs(t *testing.T) {
	expected := map[string]int64{
		"aodh":             AodhUID,
		"barbican":         BarbicanUID,
		"ceilometer":       CeilometerUID,
		"cloudkitty":       CloudkittyUID,
		"cinder":           CinderUID,
		"designate":        DesignateUID,
		"glance":           GlanceUID,
		"heat":             HeatUID,
		"horizon":          HorizonUID,
		"ironic":           IronicUID,
		"ironic-inspector": IronicInspectorUID,
		"keystone":         KeystoneUID,
		"manila":           ManilaUID,
		"memcached":        MemcachedUID,
		"mysql":            MysqlUID,
		"neutron":          NeutronUID,
		"nova":             NovaUID,
		"octavia":          OctaviaUID,
		"placement":        PlacementUID,
		"rabbitmq":         RabbitmqUID,
		"redis":            RedisUID,
		"swift":            SwiftUID,
		"watcher":          WatcherUID,
	}

	for name, uid := range expected {
		entry, ok := Registry[name]
		if !ok {
			t.Errorf("service %q missing from Registry", name)
			continue
		}
		if entry.UID != uid {
			t.Errorf("Registry[%q].UID = %d, want %d", name, entry.UID, uid)
		}
		if entry.Home == "" {
			t.Errorf("Registry[%q].Home is empty", name)
		}
	}

	if len(Registry) != len(expected) {
		t.Errorf("Registry has %d entries, expected %d", len(Registry), len(expected))
	}
}

func TestRegistryNoDuplicateUIDs(t *testing.T) {
	seen := map[int64]string{}
	for name, entry := range Registry {
		if prev, exists := seen[entry.UID]; exists {
			t.Errorf("duplicate UID %d: %q and %q", entry.UID, prev, name)
		}
		seen[entry.UID] = name
	}
}

func TestRegistryGIDMatchesUID(t *testing.T) {
	for name, entry := range Registry {
		if entry.GID != entry.UID {
			t.Errorf("Registry[%q]: GID %d != UID %d", name, entry.GID, entry.UID)
		}
	}
}
