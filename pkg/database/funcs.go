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

package database

import (
	"context"
	"fmt"
	"time"

	common "github.com/openstack-k8s-operators/lib-common/pkg/common"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//
// CreateOrPatchDB - create or patch the service DB instance
//
func CreateOrPatchDB(
	ctx context.Context,
	c client.Client,
	obj client.Object,
	schema *runtime.Scheme,
	dbOptions Options,
) (*mariadbv1.MariaDBDatabase, controllerutil.OperationResult, ctrl.Result, error) {

	db := &mariadbv1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, c, db, func() error {
		// TODO Labels
		db.Labels = common.MergeStringMaps(
			db.GetLabels(),
			dbOptions.Labels,
		)

		db.Spec.Name = dbOptions.DatabaseName
		db.Spec.Secret = dbOptions.Secret

		err := controllerutil.SetControllerReference(obj, db, schema)
		if err != nil {
			// TODO error conditions
			return err
		}

		return nil
	})
	if err != nil && !k8s_errors.IsNotFound(err) {
		return db, op, ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		// TODO: error conditions
		return db, op, ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	return db, op, ctrl.Result{}, nil
}

//
// GetDBWithName - get DB object with name in namespace
//
func GetDBWithName(
	ctx context.Context,
	c client.Client,
	name string,
	namespace string,
) (*mariadbv1.MariaDBDatabase, error) {
	db := &mariadbv1.MariaDBDatabase{}
	err := c.Get(
		ctx,
		types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
		db)
	if err != nil {
		msg := fmt.Sprintf("Failed to get %s %s ", db.GetObjectKind(), db.Name)
		// TODO condition ???
		//cond.Message = fmt.Sprintf("Failed to get persitent volume claim %s ", baseImageName)
		//cond.Reason = shared.VMSetCondReasonPersitentVolumeClaimError
		//cond.Type = shared.CommonCondTypeError
		//err = common.WrapErrorForObject(cond.Message, instance, err)

		return db, fmt.Errorf(msg)
	}

	return db, nil
}

//
// TODO WaitForDBInitialized
//
