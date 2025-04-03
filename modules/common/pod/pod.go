/*
Copyright 2022 Red Hat

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
	"context"
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// GetPodListWithLabel - Get all pods in namespace of the obj matching label selector
func GetPodListWithLabel(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	labelSelectorMap map[string]string,
) (*corev1.PodList, error) {

	labelSelectorString := labels.Set(labelSelectorMap).String()

	// use kclient to not use a cached client to be able to list pods in namespace which are not cached
	// otherwise we hit "Error listing pods for labels: map[ ... ] - unable to get: default because of unknown namespace for the cache"
	podList, err := h.GetKClient().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelectorString})
	if err != nil {
		err = fmt.Errorf("error listing pods for labels: %v - %w", labelSelectorMap, err)
		return nil, err
	}

	return podList, nil
}

// GetPodFQDNList gets a list of pods matching the labels provided and returns a slice of pod FQDNs.
func GetPodFQDNList(ctx context.Context, h *helper.Helper, namespace string, labelSelector map[string]string) ([]string, error) {
	var podSvcNames []string
	var podList *corev1.PodList

	podList, err := GetPodListWithLabel(ctx, h, namespace, labelSelector)
	if err != nil {
		return nil, fmt.Errorf("error getting list of pods: %w", err)
	}

	for _, pod := range podList.Items {
		// Check for pod.Spec.Hostname and Subdomain
		if pod.Spec.Hostname == "" || pod.Spec.Subdomain == "" {
			return nil, fmt.Errorf("%w: Pod does not have the required Spec Hostname and Subdomain details to accurately form a FQDN", util.ErrNoPodSubdomain)
		}
		podSvcNames = append(podSvcNames, fmt.Sprintf("%s.%s", pod.Spec.Hostname, pod.Spec.Subdomain))
	}

	return podSvcNames, nil
}

// GetCondition - get pod condition with PodConditionType
func GetCondition(conditions []corev1.PodCondition, t corev1.PodConditionType) *corev1.PodCondition {
	for _, condition := range conditions {
		if condition.Type == t {
			return &condition
		}
	}

	return nil
}

// IsReady - when PodReady condition == True
func IsReady(status corev1.PodStatus) (bool, *corev1.PodCondition) {
	cond := GetCondition(status.Conditions, corev1.PodReady)

	if cond != nil {
		if cond.Status == corev1.ConditionTrue {
			return true, cond
		}
		return false, cond
	}

	return false, nil
}

// StatusPodList -
func StatusPodList(podList corev1.PodList) (bool, string) {
	var message string
	var ready bool
	var cond *corev1.PodCondition
	// check all pods if they are ready
	for _, p := range podList.Items {
		// when a pod is not ready, return Ready condition message, or a default message
		// as a fallback if no Ready condition found.
		if ready, cond = IsReady(p.Status); !ready {
			if cond != nil && cond.Message != "" {
				message = cond.Message
			}
			//message = "pod Ready condition missing"

			break
		}
		//message = cond.Message
	}

	return ready, message
}
