/*
Copyright 2023 Red Hat

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

package tls

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
)

func TestAPIEnabled(t *testing.T) {
	tests := []struct {
		name string
		api  *APIService
		want bool
	}{
		{
			name: "empty API",
			api:  &APIService{},
			want: false,
		},
		{
			name: "defined API Endpoint map",
			api: &APIService{
				Disabled: nil,
				Endpoint: map[service.Endpoint]GenericService{},
			},
			want: true,
		},
		{
			name: "empty API Endpoint map",
			api: &APIService{
				Disabled: ptr.To(true),
				Endpoint: map[service.Endpoint]GenericService{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(tt.api.Enabled()).To(BeEquivalentTo(tt.want))
		})
	}
}

func TestAPIEndpointToService(t *testing.T) {
	tests := []struct {
		name string
		api  *APIService
		want map[service.Endpoint]Service
	}{
		{
			name: "empty API",
			api:  &APIService{},
			want: map[service.Endpoint]Service{},
		},
		{
			name: "empty API.Endpoint",
			api: &APIService{
				Endpoint: map[service.Endpoint]GenericService{},
			},
			want: map[service.Endpoint]Service{},
		},
		{
			name: "empty API.Endpoint entry",
			api: &APIService{
				Endpoint: map[service.Endpoint]GenericService{
					service.EndpointInternal: {},
				},
			},
			want: map[service.Endpoint]Service{},
		},
		{
			name: "empty API.Endpoint entry",
			api: &APIService{
				Endpoint: map[service.Endpoint]GenericService{
					service.EndpointInternal: {
						SecretName: ptr.To("foo"),
					},
				},
			},
			want: map[service.Endpoint]Service{
				service.EndpointInternal: {
					SecretName: "foo",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			s, err := tt.api.EndpointToServiceMap()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(s).NotTo(BeNil())
		})
	}
}

func TestGenericServiceToService(t *testing.T) {
	tests := []struct {
		name    string
		service *GenericService
		want    Service
	}{
		{
			name:    "empty APIService",
			service: &GenericService{},
			want:    Service{},
		},
		{
			name: "APIService SecretName specified",
			service: &GenericService{
				SecretName: ptr.To("foo"),
			},
			want: Service{
				SecretName: "foo",
			},
		},
		{
			name: "APIService SecretName nil",
			service: &GenericService{
				SecretName: nil,
			},
			want: Service{
				SecretName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			s, err := tt.service.ToService()
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(s).NotTo(BeNil())
		})
	}
}

func TestServiceCreateVolumeMounts(t *testing.T) {
	tests := []struct {
		name    string
		service *Service
		id      string
		want    []corev1.VolumeMount
	}{
		{
			name:    "No TLS Secret",
			service: &Service{},
			id:      "foo",
			want:    []corev1.VolumeMount{},
		},
		{
			name:    "Only TLS Secret",
			service: &Service{SecretName: "cert-secret"},
			id:      "foo",
			want: []corev1.VolumeMount{
				{
					MountPath: "/etc/pki/tls/certs/foo.crt",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.crt",
				},
				{
					MountPath: "/etc/pki/tls/private/foo.key",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.key",
				},
			},
		},
		{
			name:    "Only TLS Secret no serviceID",
			service: &Service{SecretName: "cert-secret"},
			want: []corev1.VolumeMount{
				{
					MountPath: "/etc/pki/tls/certs/default.crt",
					Name:      "default-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.crt",
				},
				{
					MountPath: "/etc/pki/tls/private/default.key",
					Name:      "default-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.key",
				},
			},
		},
		{
			name: "TLS and CA Secrets",
			service: &Service{
				SecretName: "cert-secret",
				CaMount:    ptr.To("/mount/my/ca.crt"),
			},
			id: "foo",
			want: []corev1.VolumeMount{
				{
					MountPath: "/etc/pki/tls/certs/foo.crt",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.crt",
				},
				{
					MountPath: "/etc/pki/tls/private/foo.key",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "tls.key",
				},
				{
					MountPath: "/mount/my/ca.crt",
					Name:      "foo-tls-certs",
					ReadOnly:  true,
					SubPath:   "ca.crt",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			mounts := tt.service.CreateVolumeMounts(tt.id)
			g.Expect(mounts).To(HaveLen(len(tt.want)))
			g.Expect(mounts).To(Equal(tt.want))
		})
	}
}

func TestServiceCreateVolume(t *testing.T) {
	tests := []struct {
		name    string
		service *Service
		id      string
		want    corev1.Volume
	}{
		{
			name:    "No Secrets",
			service: &Service{},
			want:    corev1.Volume{},
		},
		{
			name:    "Only TLS Secret",
			service: &Service{SecretName: "cert-secret"},
			id:      "foo",
			want: corev1.Volume{
				Name: "foo-tls-certs",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "cert-secret",
						DefaultMode: ptr.To[int32](0440),
					},
				},
			},
		},
		{
			name:    "Only TLS Secret no serviceID",
			service: &Service{SecretName: "cert-secret"},
			want: corev1.Volume{
				Name: "default-tls-certs",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "cert-secret",
						DefaultMode: ptr.To[int32](0440),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			volume := tt.service.CreateVolume(tt.id)
			g.Expect(volume).To(Equal(tt.want))
		})
	}
}

func TestCACreateVolumeMounts(t *testing.T) {
	tests := []struct {
		name          string
		ca            *Ca
		caBundleMount *string
		want          []corev1.VolumeMount
	}{
		{
			name: "Empty Ca",
			ca:   &Ca{},
			want: []corev1.VolumeMount{},
		},
		{
			name: "Only CaBundleSecretName no caBundleMount",
			ca: &Ca{
				CaBundleSecretName: "ca-secret",
			},
			want: []corev1.VolumeMount{
				{
					MountPath: "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem",
					Name:      "combined-ca-bundle",
					ReadOnly:  true,
					SubPath:   "tls-ca-bundle.pem",
				},
			},
		},
		{
			name: "CaBundleSecretName and caBundleMount",
			ca: &Ca{
				CaBundleSecretName: "ca-secret",
			},
			caBundleMount: ptr.To("/mount/my/ca.crt"),
			want: []corev1.VolumeMount{
				{
					MountPath: "/mount/my/ca.crt",
					Name:      "combined-ca-bundle",
					ReadOnly:  true,
					SubPath:   "tls-ca-bundle.pem",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			mounts := tt.ca.CreateVolumeMounts(tt.caBundleMount)
			g.Expect(mounts).To(HaveLen(len(tt.want)))
			g.Expect(mounts).To(Equal(tt.want))
		})
	}
}

func TestCaCreateVolume(t *testing.T) {
	tests := []struct {
		name string
		ca   *Ca
		want corev1.Volume
	}{
		{
			name: "Empty Ca",
			ca:   &Ca{},
			want: corev1.Volume{},
		},
		{
			name: "Set CaBundleSecretName",
			ca: &Ca{
				CaBundleSecretName: "ca-secret",
			},
			want: corev1.Volume{
				Name: "combined-ca-bundle",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  "ca-secret",
						DefaultMode: ptr.To[int32](0444),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			volume := tt.ca.CreateVolume()
			g.Expect(volume).To(Equal(tt.want))
		})
	}
}

func TestGenerateTLSConnectionConfig(t *testing.T) {
	tests := []struct {
		name         string
		services     map[string]Service // Updated to be a map
		ca           *Ca
		wantStmts    []string
		excludeStmts []string
	}{
		{
			name:         "No Secrets",
			services:     map[string]Service{}, // Empty map
			ca:           &Ca{},
			wantStmts:    []string{},
			excludeStmts: []string{"ssl=1", "ssl-cert=", "ssl-key=", "ssl-ca="},
		},
		{
			name:         "Only TLS Secret",
			services:     map[string]Service{"service1": {SecretName: "test-tls-secret"}},
			ca:           &Ca{},
			wantStmts:    []string{"ssl=1", "ssl-cert=", "ssl-key="},
			excludeStmts: []string{"ssl-ca="},
		},
		{
			name:         "Only CA Secret",
			services:     map[string]Service{},
			ca:           &Ca{CaBundleSecretName: "test-ca1"},
			wantStmts:    []string{"ssl=1", "ssl-ca="},
			excludeStmts: []string{"ssl-cert=", "ssl-key="},
		},
		{
			name:         "TLS and CA Secrets",
			services:     map[string]Service{"service1": {SecretName: "test-tls-secret"}},
			ca:           &Ca{CaBundleSecretName: "test-ca1"},
			wantStmts:    []string{"ssl=1", "ssl-cert=", "ssl-key=", "ssl-ca="},
			excludeStmts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			tlsInstance := &TLS{Service: tt.services, Ca: tt.ca}
			configStr := tlsInstance.CreateDatabaseClientConfig(nil)

			for _, stmt := range tt.wantStmts {
				g.Expect(configStr).To(ContainSubstring(stmt))
			}
			for _, stmt := range tt.excludeStmts {
				g.Expect(configStr).ToNot(ContainSubstring(stmt))
			}
		})
	}
}
