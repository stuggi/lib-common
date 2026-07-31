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

package pod

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestWritableHomeDirVolumeNoSizeLimit(t *testing.T) {
	v := WritableHomeDirVolume("home", nil)

	if v.Name != "home" {
		t.Errorf("expected Name %q, got %q", "home", v.Name)
	}
	if v.EmptyDir == nil {
		t.Fatal("expected EmptyDir to be set")
	}
	if v.EmptyDir.SizeLimit != nil {
		t.Errorf("expected no SizeLimit, got %v", v.EmptyDir.SizeLimit)
	}
}

func TestWritableHomeDirVolumeWithSizeLimit(t *testing.T) {
	limit := resource.MustParse("64Mi")
	v := WritableHomeDirVolume("home", &limit)

	if v.EmptyDir == nil || v.EmptyDir.SizeLimit == nil {
		t.Fatal("expected SizeLimit to be set")
	}
	if v.EmptyDir.SizeLimit.String() != "64Mi" {
		t.Errorf("expected SizeLimit 64Mi, got %v", v.EmptyDir.SizeLimit)
	}
}

func TestWritableHomeDirMounts(t *testing.T) {
	mounts := WritableHomeDirMounts("home", "/var/lib/service", "tmp", ".cache")

	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(mounts))
	}
	if mounts[0].Name != "home" || mounts[0].MountPath != "/var/lib/service/tmp" || mounts[0].SubPath != "tmp" {
		t.Errorf("unexpected first mount: %+v", mounts[0])
	}
	if mounts[1].Name != "home" || mounts[1].MountPath != "/var/lib/service/.cache" || mounts[1].SubPath != ".cache" {
		t.Errorf("unexpected second mount: %+v", mounts[1])
	}
}

func TestWritableHomeDirMountsNoSubdirs(t *testing.T) {
	mounts := WritableHomeDirMounts("home", "/var/lib/service")

	if len(mounts) != 0 {
		t.Errorf("expected no mounts, got %v", mounts)
	}
}
