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

package annotations

import (
	"encoding/json"
	"fmt"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
)

// GetNADAnnotation returns pod annotation for network-attachment-definition
// e.g. k8s.v1.cni.cncf.io/networks: '[{"name": "internalapi", "namespace": "openstack"},{"name": "storage", "namespace": "openstack"}]'
func GetNADAnnotation(namespace string, nads []string) (map[string]string, error) {

	netAnnotations := []networkv1.NetworkSelectionElement{}
	for _, nad := range nads {
		netAnnotations = append(
			netAnnotations,
			networkv1.NetworkSelectionElement{
				Name:      nad,
				Namespace: namespace,
			},
		)
	}

	networks, err := json.Marshal(netAnnotations)
	if err != nil {
		return nil, fmt.Errorf("failed to encode networks %s into json: %w", nads, err)
	}

	return map[string]string{networkv1.NetworkAttachmentAnnot: string(networks)}, nil
}

// GetNetworkStatusFromAnnotation returns NetworkStatus list with networking details the pods are attached to
func GetNetworkStatusFromAnnotation(annotations map[string]string) ([]networkv1.NetworkStatus, error) {

	var netStatus []networkv1.NetworkStatus

	if netStatusAnnotation, ok := annotations[networkv1.NetworkStatusAnnot]; ok {
		err := json.Unmarshal([]byte(netStatusAnnotation), &netStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to decode networks status %s: %w", netStatusAnnotation, err)
		}
	}

	return netStatus, nil
}
