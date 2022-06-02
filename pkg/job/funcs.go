/*
Copyright 2021 Red Hat

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

package job

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/openstack-k8s-operators/lib-common/pkg/common"
	"github.com/openstack-k8s-operators/lib-common/pkg/helper"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"errors"
)

// NewJob returns an initialized Job.
func NewJob(
	job *batchv1.Job,
	jobType string,
	preserve bool,
	timeout int,
	beforeHash string,
) *Job {

	return &Job{
		job:        job,
		jobType:    jobType,
		preserve:   preserve,
		timeout:    timeout,
		beforeHash: beforeHash,
		changed:    false,
	}
}

// createJob - creates job, reconciles after Xs if object won't exist.
func (j *Job) createJob(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), j.job, func() error {
		err := controllerutil.SetControllerReference(h.GetBeforeObject(), j.job, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Job %s not found, reconcile in %ds", j.job.Name, j.timeout))
			return ctrl.Result{RequeueAfter: time.Duration(j.timeout) * time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Job %s %s - %s", j.jobType, j.job.Name, op))
		return ctrl.Result{RequeueAfter: time.Duration(j.timeout) * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

//
// DoJob - run a job if the hashBefore and hash is different. If there is an existing job, the job gets deleted
// and re-created. If the job finished successful and preserve flag is not set it gets deleted.
//
func (j *Job) DoJob(
	ctx context.Context,
	h *helper.Helper,
	//	hashMap map[string]string,
	//) (map[string]string, ctrl.Result, error) {
) (ctrl.Result, error) {
	var ctrlResult ctrl.Result
	var err error

	j.hash, err = common.ObjectHash(j.job)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error calculating %s hash: %v", j.jobType, err)
	}

	// if the hash changed the job should run
	if j.beforeHash != j.hash {
		j.changed = true
	}

	//
	// Check if this job already exists
	//
	err = h.GetClient().Get(ctx, types.NamespacedName{Name: j.job.Name, Namespace: j.job.Namespace}, j.job)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if k8s_errors.IsNotFound(err) {
		if j.changed {
			ctrlResult, err = j.createJob(ctx, h)
			if err != nil {
				return ctrlResult, err
			}
		}
	} else {
		if j.changed {
			err = j.DeleteJob(ctx, h)
			if err != nil {
				return ctrl.Result{}, err
			}

			ctrlResult, err = j.createJob(ctx, h)
			if err != nil {
				return ctrlResult, err
			}
		}

		requeue, err := j.WaitOnJob(ctx, h)
		if err != nil {
			return ctrl.Result{}, err
		} else if requeue {
			h.GetLogger().Info(fmt.Sprintf("Waiting on %s Job %s", j.jobType, j.job.Name))
			return ctrl.Result{RequeueAfter: time.Second * time.Duration(j.timeout)}, err
		}

		// delete the job if PreserveJobs is not enabled
		if !j.preserve {
			err = j.DeleteJob(ctx, h)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

	}

	return ctrl.Result{}, nil
}

// DeleteJob func
// kclient required to properly cleanup the job depending pods with DeleteOptions
func (j *Job) DeleteJob(
	ctx context.Context,
	h *helper.Helper,
) error {
	foundJob, err := h.GetKClient().BatchV1().Jobs(j.job.Namespace).Get(ctx, j.job.Name, metav1.GetOptions{})
	if err == nil {
		h.GetLogger().Info("Deleting Job", "Job.Namespace", j.job.Namespace, "Job.Name", j.job.Name)
		background := metav1.DeletePropagationBackground
		err = h.GetKClient().BatchV1().Jobs(foundJob.Namespace).Delete(
			ctx, foundJob.Name, metav1.DeleteOptions{PropagationPolicy: &background})
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

// WaitOnJob func
func (j *Job) WaitOnJob(
	ctx context.Context,
	h *helper.Helper,
) (bool, error) {
	// Check if this Job already exists
	foundJob := &batchv1.Job{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: j.job.Name, Namespace: j.job.Namespace}, foundJob)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Error(err, "Job was not found.")
			return true, err
		}
		h.GetLogger().Info("WaitOnJob err")
		return true, err
	}

	if foundJob.Status.Active > 0 {
		h.GetLogger().Info("Job Status Active... requeuing")
		return true, err
	} else if foundJob.Status.Failed > 0 {
		h.GetLogger().Info("Job Status Failed")
		return true, k8s_errors.NewInternalError(errors.New("Job Failed. Check job logs"))
	} else if foundJob.Status.Succeeded > 0 {
		h.GetLogger().Info("Job Status Successful")
	} else {
		h.GetLogger().Info("Job Status incomplete... requeuing")
		return true, err
	}

	return false, nil

}

// HasChanged func
func (j *Job) HasChanged() bool {
	return j.changed
}

// GetHash func
func (j *Job) GetHash() string {
	return j.hash
}
