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

// +kubebuilder:object:generate:=true

package tls

import (
	"context"
	"fmt"
	"strings"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

const (
	// CABundleLabel added to the CA bundle secret for the namespace
	CABundleLabel = "combined-ca-bundle"
)

// Service contains server-specific TLS secret
type Service struct {
	// +kubebuilder:validation:Optional
	// SecretName - holding the cert, key for the service
	SecretName string `json:"secretName,omitempty"`
	// +kubebuilder:validation:Optional
	// CertMount - dst location to mount the service tls.crt cert. Can be used to override the default location which is /etc/tls/<service key>/tls.crt
	CertMount *string `json:"certMount,omitempty"`
	// +kubebuilder:validation:Optional
	// KeyMount - dst location to mount the service tls.key  key. Can be used to override the default location which is /etc/tls/<service key>/tls.key
	KeyMount *string `json:"keyMount,omitempty"`
	// +kubebuilder:validation:Optional
	// CaMount - dst location to mount the CA cert ca.crt to. Can be used if the service CA cert should be mounted specifically, e.g. to be set in a service config for validation, instead of the env wide bundle.
	CaMount *string `json:"caMount,omitempty"`
	// +kubebuilder:validation:Optional
	// DisableNonTLSListeners - disable non TLS listeners of the service (if supported)
	DisableNonTLSListeners bool `json:"disableNonTLSListeners,omitempty"`
}

// Ca contains CA-specific settings, which could be used both by services (to define their own CA certificates)
// and by clients (to verify the server's certificate)
type Ca struct {
	// +kubebuilder:validation:Optional
	// CaBundleSecretName - holding the CA certs in a pre-created bundle file
	CaBundleSecretName string `json:"caBundleSecretName"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"
	// CaBundleMount - dst location to mount the CA cert bundle
	CaBundleMount string `json:"caBundleMount"`
}

// TLS - a generic type, which encapsulates both the service and CA configurations
// Service is for the services with a single endpoint
// TypedSecretName handles multiple service endpoints with respective secrets
type TLS struct {
	// certificate configuration for API service certs
	APIService map[service.Endpoint]Service `json:"APIService"`
	// certificate configuration for additional arbitrary certs
	Service map[string]Service `json:"service"`
	// CA bundle configuration
	Ca *Ca `json:"ca"`
}

// NewTLS - initialize and return a TLS struct
func NewTLS(ctx context.Context, h *helper.Helper, namespace string, serviceMap map[string]Service, endpointMap map[string]service.Endpoint, ca *Ca) (*TLS, error) {

	apiService := make(map[service.Endpoint]Service)

	// Ensure service SecretName exists for each service in the map or return an error
	for serviceName, service := range serviceMap {
		if service.SecretName != "" {
			secretData, _, err := secret.GetSecret(ctx, h, service.SecretName, namespace)
			if err != nil {
				return nil, fmt.Errorf("error ensuring secret %s exists for service '%s': %w", service.SecretName, serviceName, err)
			}

			_, keyOk := secretData.Data["tls.key"]
			_, certOk := secretData.Data["tls.crt"]
			if !keyOk || !certOk {
				return nil, fmt.Errorf("secret %s for service '%s' does not contain both tls.key and tls.crt", service.SecretName, serviceName)
			}
		}

		// Use the endpointMap to get the correct Endpoint type for the apiService key
		endpoint, ok := endpointMap[serviceName]
		if !ok {
			return nil, fmt.Errorf("no endpoint defined for service '%s'", serviceName)
		}
		apiService[endpoint] = service
	}

	return &TLS{
		APIService: apiService,
		Service:    serviceMap,
		Ca:         ca,
	}, nil
}

// CreateVolumeMounts - add volume mount for TLS certificates and CA certificate for the service
func (s *Service) CreateVolumeMounts() []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount

	if s.SecretName != "" {
		certMountPath := "/etc/pki/tls/certs/tls.crt"
		if s.CertMount != nil {
			certMountPath = *s.CertMount
		}

		keyMountPath := "/etc/pki/tls/private/tls.key"
		if s.KeyMount != nil {
			keyMountPath = *s.KeyMount
		}

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "tls-crt",
			MountPath: certMountPath,
			SubPath:   "tls.crt",
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      "tls-key",
			MountPath: keyMountPath,
			SubPath:   "tls.key",
			ReadOnly:  true,
		})
	}

	if s.CaMount != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "ca-certs",
			MountPath: *s.CaMount,
			SubPath:   "ca.crt",
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

// CreateVolumes - add volume for TLS certificates and CA certificate for the service
func (s *Service) CreateVolumes() []corev1.Volume {
	var volumes []corev1.Volume

	if s.SecretName != "" {
		volume := corev1.Volume{
			Name: "tls-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  s.SecretName,
					DefaultMode: ptr.To[int32](0440),
				},
			},
		}
		volumes = append(volumes, volume)
	}

	if s.CaMount != nil {
		caVolume := corev1.Volume{
			Name: "ca-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  *s.CaMount,
					DefaultMode: ptr.To[int32](0444),
				},
			},
		}
		volumes = append(volumes, caVolume)
	}

	return volumes
}

// CreateVolumeMounts creates volume mounts for CA bundle file
func (c *Ca) CreateVolumeMounts() []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount

	if c.CaBundleMount != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      CABundleLabel,
			MountPath: c.CaBundleMount,
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

// CreateVolumes creates volumes for CA bundle file
func (c *Ca) CreateVolumes() []corev1.Volume {
	var volumes []corev1.Volume

	volume := corev1.Volume{
		Name: CABundleLabel,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  c.CaBundleSecretName,
				DefaultMode: ptr.To[int32](0444),
			},
		},
	}
	volumes = append(volumes, volume)

	return volumes
}

// CreateDatabaseClientConfig - connection flags for the MySQL client
// Configures TLS connections for clients that use TLS certificates
// returns a string of mysql config statements
// (vfisarov): Note dciabrin to recheck this after updates
func (t *TLS) CreateDatabaseClientConfig() string {
	conn := []string{}

	// This assumes certificates are always injected in
	// a common directory for all services
	for _, service := range t.Service {
		if service.SecretName != "" {
			certPath := "/etc/pki/tls/certs/tls.crt"
			keyPath := "/etc/pki/tls/private/tls.key"

			// Override paths if custom mounts are defined
			if service.CertMount != nil {
				certPath = *service.CertMount
			}
			if service.KeyMount != nil {
				keyPath = *service.KeyMount
			}

			conn = append(conn,
				fmt.Sprintf("ssl-cert=%s", certPath),
				fmt.Sprintf("ssl-key=%s", keyPath),
			)
		}
	}

	// Client uses a CA certificate that gets merged
	// into the pod's CA bundle by kolla_start
	if t.Ca != nil && t.Ca.CaBundleSecretName != "" {
		caPath := "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"
		if t.Ca.CaBundleMount != "" {
			caPath = t.Ca.CaBundleMount
		}
		conn = append(conn, fmt.Sprintf("ssl-ca=%s", caPath))
	}

	if len(conn) > 0 {
		conn = append([]string{"ssl=1"}, conn...)
	}

	return strings.Join(conn, "\n")
}
