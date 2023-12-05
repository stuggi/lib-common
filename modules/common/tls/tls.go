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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// CABundleLabel added to the CA bundle secret for the namespace
	CABundleLabel = "combined-ca-bundle"
	// CABundleKey - key of the secret entry holding the ca bundle
	CABundleKey = "tls-ca-bundle.pem"

	// CertKey - key of the secret entry holding the cert
	CertKey = "tls.crt"
	// PrivateKey - key of the secret entry holding the cert private key
	PrivateKey = "tls.key"
	// CAKey - key of the secret entry holding the ca
	CAKey = "ca.crt"

	// TLSHashName - Name of the hash of hashes of all cert resources used to indentify a change
	TLSHashName = "certs"
)

// API defines the observed state of TLS with API only
type API struct {
	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// API tls type which encapsulates for API services
	API APIService `json:"api,omitempty"`

	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// Secret containing CA bundle
	Ca `json:",inline"`
}

// APIService - API tls type which encapsulates for API services
type APIService struct {
	// +kubebuilder:validation:Optional
	// Disabled TLS for the deployment of the service
	Disabled *bool `json:"disabled,omitempty"`

	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// The key must be the endpoint type (public, internal)
	Public GenericService `json:"public,omitempty"`

	// +kubebuilder:validation:optional
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// The key must be the endpoint type (public, internal)
	Internal GenericService `json:"internal,omitempty"`
}

// GenericService contains server-specific TLS secret or issuer
type GenericService struct {
	// +kubebuilder:validation:Optional
	// SecretName - holding the cert, key for the service
	SecretName *string `json:"secretName,omitempty"`
}

// Ca contains CA-specific settings, which could be used both by services (to define their own CA certificates)
// and by clients (to verify the server's certificate)
type Ca struct {
	// CaBundleSecretName - holding the CA certs in a pre-created bundle file
	CaBundleSecretName string `json:"caBundleSecretName,omitempty"`
}

// Service contains server-specific TLS secret
// +kubebuilder:object:generate:=false
type Service struct {
	// SecretName - holding the cert, key for the service
	SecretName string `json:"secretName"`

	// CertMount - dst location to mount the service tls.crt cert. Can be used to override the default location which is /etc/tls/certs/<service id>.crt
	CertMount *string `json:"certMount,omitempty"`

	// KeyMount - dst location to mount the service tls.key  key. Can be used to override the default location which is /etc/tls/private/<service id>.key
	KeyMount *string `json:"keyMount,omitempty"`

	// CaMount - dst location to mount the CA cert ca.crt to. Can be used if the service CA cert should be mounted specifically, e.g. to be set in a service config for validation, instead of the env wide bundle.
	CaMount *string `json:"caMount,omitempty"`
}

// Enabled - returns true if tls is configured for the public and internal
func (a *APIService) Enabled(endpt service.Endpoint) bool {
	switch endpt {
	case service.EndpointPublic:
		return (a.Disabled == nil || (a.Disabled != nil && !*a.Disabled)) && a.Public.SecretName != nil
	case service.EndpointInternal:
		return (a.Disabled == nil || (a.Disabled != nil && !*a.Disabled)) && a.Internal.SecretName != nil
	}

	return false
}

// ValidateCertSecrets - validates the content of the cert secrets to make sure "tls-ca-bundle.pem" key exist
func (a *APIService) ValidateCertSecrets(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
) (string, ctrl.Result, error) {
	var svc GenericService
	certHashes := map[string]env.Setter{}
	for _, endpt := range []service.Endpoint{service.EndpointInternal, service.EndpointPublic} {
		switch endpt {
		case service.EndpointPublic:
			if !a.Enabled(service.EndpointPublic) {
				continue
			}

			svc = a.Public

		case service.EndpointInternal:
			if !a.Enabled(service.EndpointInternal) {
				continue
			}

			svc = a.Public
		}

		endptTLSCfg, err := svc.ToService()
		if err != nil {
			return "", ctrl.Result{}, err
		}

		if endptTLSCfg.SecretName != "" {
			// validate the cert secret has the expected keys
			hash, ctrlResult, err := endptTLSCfg.ValidateCertSecret(ctx, h, namespace)
			if err != nil {
				return "", ctrlResult, err
			} else if (ctrlResult != ctrl.Result{}) {
				return "", ctrlResult, nil
			}

			certHashes["cert-"+endpt.String()] = env.SetValue(hash)
		}
	}

	certsHash, err := util.HashOfInputHashes(certHashes)
	if err != nil {
		return "", ctrl.Result{}, err
	}
	return certsHash, ctrl.Result{}, nil
}

// ToService - convert tls.APIService to tls.Service
func (s *GenericService) ToService() (*Service, error) {
	toS := &Service{}

	sBytes, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("error marshalling api service: %w", err)
	}

	err = json.Unmarshal(sBytes, toS)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling tls service: %w", err)
	}

	return toS, nil
}

// ValidateCACertSecret - validates the content of the cert secret to make sure "tls-ca-bundle.pem" key exist
func ValidateCACertSecret(
	ctx context.Context,
	c client.Client,
	caSecret types.NamespacedName,
) (string, ctrl.Result, error) {
	hash, ctrlResult, err := secret.VerifySecret(
		ctx,
		caSecret,
		[]string{CABundleKey},
		c,
		5*time.Second)
	if err != nil {
		return "", ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return "", ctrlResult, nil
	}

	return hash, ctrl.Result{}, nil
}

// ValidateCertSecret - validates the content of the cert secret to make sure "tls.key", "tls.crt" and optional "ca.crt" keys exist
func (s *Service) ValidateCertSecret(ctx context.Context, h *helper.Helper, namespace string) (string, ctrl.Result, error) {
	// define keys to expect in cert secret
	keys := []string{PrivateKey, CertKey}
	if s.CaMount != nil {
		keys = append(keys, CAKey)
	}

	hash, ctrlResult, err := secret.VerifySecret(
		ctx,
		types.NamespacedName{Name: s.SecretName, Namespace: namespace},
		keys,
		h.GetClient(),
		5*time.Second)
	if err != nil {
		return "", ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return "", ctrlResult, nil
	}

	return hash, ctrl.Result{}, nil
}

// ValidateEndpointCerts - validates all services from an endpointCfgs and
// returns the hash of hashes for all the certificates
func ValidateEndpointCerts(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	endpointCfgs map[service.Endpoint]Service,
) (string, ctrl.Result, error) {
	certHashes := map[string]env.Setter{}
	for endpt, endpointTLSCfg := range endpointCfgs {
		if endpointTLSCfg.SecretName != "" {
			// validate the cert secret has the expected keys
			hash, ctrlResult, err := endpointTLSCfg.ValidateCertSecret(ctx, h, namespace)
			if err != nil {
				return "", ctrlResult, err
			} else if (ctrlResult != ctrl.Result{}) {
				return "", ctrlResult, nil
			}

			certHashes["cert-"+endpt.String()] = env.SetValue(hash)
		}
	}

	certsHash, err := util.HashOfInputHashes(certHashes)
	if err != nil {
		return "", ctrl.Result{}, err
	}
	return certsHash, ctrl.Result{}, nil
}

// getCertMountPath - return certificate mount path
func (s *Service) getCertMountPath(serviceID string) string {
	if serviceID == "" {
		serviceID = "default"
	}

	certMountPath := fmt.Sprintf("/etc/pki/tls/certs/%s.crt", serviceID)
	if s.CertMount != nil {
		certMountPath = *s.CertMount
	}

	return certMountPath
}

// getKeyMountPath - return key mount path
func (s *Service) getKeyMountPath(serviceID string) string {
	if serviceID == "" {
		serviceID = "default"
	}

	keyMountPath := fmt.Sprintf("/etc/pki/tls/private/%s.key", serviceID)
	if s.KeyMount != nil {
		keyMountPath = *s.KeyMount
	}

	return keyMountPath
}

// CreateVolumeMounts - add volume mount for TLS certificates and CA certificate for the service
func (s *Service) CreateVolumeMounts(serviceID string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}
	if serviceID == "" {
		serviceID = "default"
	}
	if s.SecretName != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      serviceID + "-tls-certs",
			MountPath: s.getCertMountPath(serviceID),
			SubPath:   CertKey,
			ReadOnly:  true,
		}, corev1.VolumeMount{
			Name:      serviceID + "-tls-certs",
			MountPath: s.getKeyMountPath(serviceID),
			SubPath:   PrivateKey,
			ReadOnly:  true,
		})

		if s.CaMount != nil {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      serviceID + "-tls-certs",
				MountPath: *s.CaMount,
				SubPath:   CAKey,
				ReadOnly:  true,
			})
		}
	}

	return volumeMounts
}

// CreateVolume - add volume for TLS certificates and CA certificate for the service
func (s *Service) CreateVolume(serviceID string) corev1.Volume {
	volume := corev1.Volume{}
	if serviceID == "" {
		serviceID = "default"
	}
	if s.SecretName != "" {
		volume = corev1.Volume{
			Name: serviceID + "-tls-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  s.SecretName,
					DefaultMode: ptr.To[int32](0440),
				},
			},
		}
	}

	return volume
}

// CreateVolumeMounts creates volume mounts for CA bundle file
func (c *Ca) CreateVolumeMounts(caBundleMount *string) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{}

	if caBundleMount == nil {
		caBundleMount = ptr.To("/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem")
	}

	if c.CaBundleSecretName != "" {
		volumeMounts = []corev1.VolumeMount{
			{
				Name:      CABundleLabel,
				MountPath: *caBundleMount,
				SubPath:   CABundleKey,
				ReadOnly:  true,
			},
		}
	}

	return volumeMounts
}

// CreateVolume creates volumes for CA bundle file
func (c *Ca) CreateVolume() corev1.Volume {
	volume := corev1.Volume{}

	if c.CaBundleSecretName != "" {
		volume = corev1.Volume{
			Name: CABundleLabel,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  c.CaBundleSecretName,
					DefaultMode: ptr.To[int32](0444),
				},
			},
		}
	}

	return volume
}

// CreateDatabaseClientConfig - connection flags for the MySQL client
// Configures TLS connections for clients that use TLS certificates
// returns a string of mysql config statements
// With the serviceID it is possible to control which certificate
// to be use if there are multiple mounted to the deployment.
func (s *Service) CreateDatabaseClientConfig(serviceID string) string {
	conn := []string{}

	if serviceID != "" || (s.CertMount != nil && s.KeyMount != nil) {
		certPath := s.getCertMountPath(serviceID)
		keyPath := s.getKeyMountPath(serviceID)

		conn = append(conn,
			fmt.Sprintf("ssl-cert=%s", certPath),
			fmt.Sprintf("ssl-key=%s", keyPath),
		)
	}

	// Client uses a CA certificate
	caPath := "/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"
	if s.CaMount != nil {
		caPath = *s.CaMount
	}
	conn = append(conn, fmt.Sprintf("ssl-ca=%s", caPath))

	if len(conn) > 0 {
		conn = append([]string{"ssl=1"}, conn...)
	}

	return strings.Join(conn, "\n")
}
