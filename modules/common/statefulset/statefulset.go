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

package statefulset

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/pod"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	appsv1 "k8s.io/api/apps/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewStatefulSet returns an initialized NewStatefulset.
func NewStatefulSet(
	statefulset *appsv1.StatefulSet,
	timeout time.Duration,
) *StatefulSet {
	return &StatefulSet{
		statefulset: statefulset,
		timeout:     timeout,
	}
}

// CreateOrPatch - creates or patches a statefulset, reconciles after Xs if object won't exist.
func (s *StatefulSet) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.statefulset.Name,
			Namespace: s.statefulset.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), statefulset, func() error {
		// selector is immutable so we set this value only if
		// a new object is going to be created
		if statefulset.ObjectMeta.CreationTimestamp.IsZero() {
			statefulset.Spec.Selector = s.statefulset.Spec.Selector
		}

		statefulset.Annotations = util.MergeStringMaps(statefulset.Annotations, s.statefulset.Annotations)
		statefulset.Labels = util.MergeStringMaps(statefulset.Labels, s.statefulset.Labels)
		// We need to copy the Spec field by field as Selector is not updatable
		// This list needs to be synced StatefulSet to gain ability to set
		// those new fields via lib-common
		statefulset.Spec.Replicas = s.statefulset.Spec.Replicas
		statefulset.Spec.Template = s.statefulset.Spec.Template
		statefulset.Spec.VolumeClaimTemplates = s.statefulset.Spec.VolumeClaimTemplates
		statefulset.Spec.ServiceName = s.statefulset.Spec.ServiceName
		statefulset.Spec.PodManagementPolicy = s.statefulset.Spec.PodManagementPolicy
		statefulset.Spec.UpdateStrategy = s.statefulset.Spec.UpdateStrategy
		statefulset.Spec.RevisionHistoryLimit = s.statefulset.Spec.RevisionHistoryLimit
		statefulset.Spec.MinReadySeconds = s.statefulset.Spec.MinReadySeconds
		statefulset.Spec.PersistentVolumeClaimRetentionPolicy = s.statefulset.Spec.PersistentVolumeClaimRetentionPolicy

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), statefulset, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("StatefulSet %s not found, reconcile in %s", statefulset.Name, s.timeout))
			return ctrl.Result{RequeueAfter: s.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	// update the deployment object of the deployment type
	s.statefulset = statefulset

	h.GetLogger().Info(fmt.Sprintf("StatefulSet %s %s", statefulset.Name, op))
	// Only poll on Deployment updates, not on initial create.
	if op != controllerutil.OperationResultCreated {
		// only poll if replicas > 0
		if s.statefulset.Spec.Replicas != nil && *s.statefulset.Spec.Replicas > 0 {
			// Ignore context.DeadlineExceeded when PollUntilContextTimeout reached
			// the poll timeout. d.rolloutStatus as information on the
			// replica rollout, the consumer can evaluate the rolloutStatus and
			// retry/reconcile until RolloutComplete, or ProgressDeadlineExceeded.
			if err := s.PollRolloutStatus(ctx, h); err != nil && !errors.Is(err, context.DeadlineExceeded) &&
				!strings.Contains(err.Error(), "would exceed context deadline") {
				return ctrl.Result{}, fmt.Errorf("poll rollout error: %w", err)
			}
		}
	}

	return ctrl.Result{}, nil
}

// PollRolloutStatus - will poll the statefulset rollout to verify its status for Complet, Failed or polling until timeout.
//
// - Complete - all replicas updated using RolloutComplete()
//
// - Failed   - rollout of new config failed and the new pod is stuck in ProgressDeadlineExceeded using ProgressDeadlineExceeded()
func (s *StatefulSet) PollRolloutStatus(
	ctx context.Context,
	h *helper.Helper,
) error {
	if s.rolloutPollInterval == nil {
		s.rolloutPollInterval = ptr.To(DefaultPollInterval)
	}
	if s.rolloutPollTimeout == nil {
		s.rolloutPollTimeout = ptr.To(DefaultPollTimeout)
	}

	err := wait.PollUntilContextTimeout(ctx, *s.rolloutPollInterval, *s.rolloutPollTimeout, true, func(ctx context.Context) (bool, error) {
		// Fetch deployment object
		depl, err := GetStatefulSetWithName(ctx, h, s.statefulset.Name, s.statefulset.Namespace)
		if err != nil {
			return false, err
		}
		s.statefulset = depl

		// Check if rollout is complete
		if Complete(s.statefulset.Status, s.statefulset.Generation) {
			s.rolloutStatus = ptr.To(DeploymentPollCompleted)
			s.rolloutMessage = fmt.Sprintf(DeploymentPollCompletedMessage, s.statefulset.Name)
			h.GetLogger().Info(s.rolloutMessage)
			// If rollout is complete, return true to stop polling
			return true, nil
		}

		// statefulset does not have deployment conditions on
		// the status itself. have to check the pods
		podList, err := pod.GetPodListWithLabel(ctx, h, s.statefulset.Namespace, s.statefulset.Spec.Template.Labels)
		if err != nil {
			return false, err
		}

		if ready, msg := pod.StatusPodList(*podList); !ready {
			s.rolloutStatus = ptr.To(DeploymentPollProgressing)
			s.rolloutMessage = fmt.Sprintf(DeploymentPollProgressingMessage, s.statefulset.Name,
				s.statefulset.Status.UpdatedReplicas, s.statefulset.Status.Replicas, msg)
			return false, nil
		}

		// If rollout reached complete while checking the pods, return true to stop polling
		return true, nil
	})

	return err
}

// RolloutComplete -
func (s *StatefulSet) RolloutComplete() bool {
	return s.GetRolloutStatus() != nil && *s.GetRolloutStatus() == DeploymentPollCompleted
}

// Complete -
func Complete(status appsv1.StatefulSetStatus, generation int64) bool {
	return status.UpdatedReplicas == status.Replicas &&
		status.Replicas == status.AvailableReplicas &&
		status.ObservedGeneration == generation
}

// GetRolloutStatus - get rollout status of the deployment.
func (s *StatefulSet) GetRolloutStatus() *string {
	return s.rolloutStatus
}

// GetRolloutMessage - get rollout message of the deployment.
func (s *StatefulSet) GetRolloutMessage() string {
	return s.rolloutMessage
}

// GetStatefulSet - get the statefulset object.
func (s *StatefulSet) GetStatefulSet() appsv1.StatefulSet {
	return *s.statefulset
}

// GetStatefulSetWithName func
func GetStatefulSetWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*appsv1.StatefulSet, error) {

	depl := &appsv1.StatefulSet{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, depl)
	if err != nil {
		return depl, err
	}

	return depl, nil
}

// Delete - delete a statefulset.
func (s *StatefulSet) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	err := h.GetClient().Delete(ctx, s.statefulset)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("Error deleting statefulset %s: %w", s.statefulset.Name, err)
		return err
	}

	return nil
}
