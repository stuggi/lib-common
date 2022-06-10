/*
Copyright 2020 Red Hat

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

package common

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// GetGroupLabel -
func GetGroupLabel(serviceName string) string {
	return serviceName + ".openstack.org"
}

// GetOwnerUIDLabelSelector -
func GetOwnerUIDLabelSelector(groupLabel string) string {
	return groupLabel + "/uid"
}

// GetOwnerNameSpaceLabelSelector -
func GetOwnerNameSpaceLabelSelector(groupLabel string) string {
	return groupLabel + "/namespace"
}

// GetOwnerNameLabelSelector -
func GetOwnerNameLabelSelector(groupLabel string) string {
	return groupLabel + "/name"
}

// GetLabels - create default labels map, additional custom labels can be passed
func GetLabels(
	obj metav1.Object,
	groupLabel string,
	custom map[string]string,
) map[string]string {
	ownerUIDLabelSelector := GetOwnerUIDLabelSelector(groupLabel)
	ownerNameSpaceLabelSelector := GetOwnerNameSpaceLabelSelector(groupLabel)
	ownerNameLabelSelector := GetOwnerNameLabelSelector(groupLabel)

	// Labels for all objects
	labelSelector := map[string]string{
		ownerUIDLabelSelector:       string(obj.GetUID()),
		ownerNameSpaceLabelSelector: obj.GetNamespace(),
		ownerNameLabelSelector:      obj.GetName(),
	}

	return MergeStringMaps(labelSelector, custom)
}
