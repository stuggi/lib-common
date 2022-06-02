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

	"github.com/openstack-k8s-operators/lib-common/pkg/common"
	"github.com/openstack-k8s-operators/lib-common/pkg/helper"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewDatabase returns an initialized DB.
func NewDatabase(
	databaseHostname string,
	databaseName string,
	databaseUser string,
	secret string,
	labels map[string]string,
) *Database {
	return &Database{
		databaseHostname: databaseHostname,
		databaseName:     databaseName,
		databaseUser:     databaseUser,
		secret:           secret,
		labels:           labels,
	}
}

//
// CreateOrPatchDB - create or patch the service DB instance
//
func (d *Database) CreateOrPatchDB(
	ctx context.Context,
	h *helper.Helper,
) (controllerutil.OperationResult, ctrl.Result, error) {
	db := &mariadbv1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.GetBeforeObject().GetName(),
			Namespace: h.GetBeforeObject().GetNamespace(),
		},
		Spec: mariadbv1.MariaDBDatabaseSpec{
			// the DB name must not change, therefore specify it outside the mutuate function
			Name: d.databaseName,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), db, func() error {
		// TODO Labels
		db.Labels = common.MergeStringMaps(
			db.GetLabels(),
			d.labels,
		)

		db.Spec.Secret = d.secret

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), db, h.GetScheme())
		if err != nil {
			// TODO error conditions
			return err
		}

		return nil
	})
	if err != nil && !k8s_errors.IsNotFound(err) {
		return op, ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		// TODO: error conditions
		return op, ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	return op, ctrl.Result{}, nil
}

//
// GetDBWithName - get DB object with name in namespace
//
func (d *Database) GetDBWithName(
	ctx context.Context,
	h *helper.Helper,
) (*mariadbv1.MariaDBDatabase, error) {
	db := &mariadbv1.MariaDBDatabase{}
	err := h.GetClient().Get(
		ctx,
		types.NamespacedName{
			Name:      d.databaseName,
			Namespace: h.GetBeforeObject().GetNamespace(),
		},
		db)
	if err != nil {
		msg := fmt.Sprintf("Failed to get %s %s ", d.databaseName, h.GetBeforeObject().GetNamespace())
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
