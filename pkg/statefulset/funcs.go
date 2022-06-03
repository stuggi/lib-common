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
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/pkg/common"
	"github.com/openstack-k8s-operators/lib-common/pkg/helper"
	appsv1 "k8s.io/api/apps/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewStatefulSet returns an initialized NewStatefulset.
func NewStatefulSet(
	statefulset *appsv1.StatefulSet,
	labels map[string]string,
	timeout int,
) *StatefulSet {
	return &StatefulSet{
		statefulset: statefulset,
		timeout:     timeout,
	}
}

/*
// GenericRoute func
func GenericRoute(routeInfo *GenericRouteDetails) *routev1.Route {
	serviceRef := routev1.RouteTargetReference{
		Kind: "Service",
		Name: routeInfo.ServiceName,
	}
	routePort := &routev1.RoutePort{
		TargetPort: intstr.FromString(routeInfo.TargetPortName),
	}

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeInfo.Name,
			Namespace: routeInfo.Namespace,
			Labels:    routeInfo.Labels,
		},
		Spec: routev1.RouteSpec{
			To:   serviceRef,
			Port: routePort,
		},
	}
}
*/

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
		statefulset.Labels = common.MergeStringMaps(statefulset.Labels, statefulset.Labels)
		statefulset.Annotations = s.statefulset.Annotations
		statefulset.Spec = s.statefulset.Spec

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), statefulset, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("StatefulSet %s not found, reconcile in %ds", statefulset.Name, s.timeout))
			return ctrl.Result{RequeueAfter: time.Duration(s.timeout) * time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("StatefulSet %s - %s", statefulset.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete a statefulset.
func (s *StatefulSet) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	err := h.GetClient().Delete(ctx, s.statefulset)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("Error deleting statefulset %s: %v", s.statefulset.Name, err)
		return err
	}

	return nil
}
