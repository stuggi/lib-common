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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// WritableHomeDirVolume returns an emptyDir Volume for mounting writable
// subdirectories of a service's home directory under ReadOnlyRootFilesystem
// (see WritableHomeDirMounts). sizeLimit is optional — pass nil to leave the
// volume's own SizeLimit unset (usage is then bounded only by the node's
// available ephemeral storage / any pod-level ephemeral-storage limit, not
// by a dedicated per-volume cap).
func WritableHomeDirVolume(volumeName string, sizeLimit *resource.Quantity) corev1.Volume {
	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{SizeLimit: sizeLimit},
		},
	}
}

// WritableHomeDirMounts returns SubPath VolumeMounts onto the emptyDir named
// volumeName (see WritableHomeDirVolume) for the given subdirectories of a
// service's home directory (homeDir). Mounted via SubPath onto specific
// subdirectories rather than shadowing homeDir itself, since the image may
// bake real content there (e.g. shell dotfiles from /etc/skel). Common
// subdirs: "tmp" (e.g. [oslo_concurrency] lock_path, tooz state_path) and
// ".cache" (RHEL's python3-setuptools/pkg_resources downstream patch caches
// iter_entry_points() scans under $HOME/.cache/python-entrypoints/<hash> on
// every process start — a platform-level Python behavior seen across
// multiple services, not specific to any one of them).
func WritableHomeDirMounts(volumeName, homeDir string, subdirs ...string) []corev1.VolumeMount {
	mounts := make([]corev1.VolumeMount, 0, len(subdirs))
	for _, subdir := range subdirs {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: homeDir + "/" + subdir,
			SubPath:   subdir,
		})
	}
	return mounts
}
