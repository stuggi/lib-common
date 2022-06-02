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

const (
	test = "test"
)

/*

import (
	"fmt"

	common "github.com/openstack-k8s-operators/lib-common/pkg/common"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DbSyncJob func
//func DbSyncJob(cr *keystonev1beta1.KeystoneAPI, cmName string) (*batchv1.Job, error) {
func DbSyncJob(options DBSyncOptions) (*batchv1.Job, error) {

	runAsUser := int64(0)

	passwordInitCmd, err := common.ExecuteTemplateFile("password_init.sh", nil)
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-db-sync", options.ServiceName),
			Namespace: options.Namespace,
			Labels:    options.Labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:      "OnFailure",
					ServiceAccountName: options.ServiceAccount,
					Containers: []corev1.Container{
						{
							Name: fmt.Sprintf("%s-db-sync", options.ServiceName),
							//Command: []string{"/bin/sleep", "7000"},
							Image: options.ContainerImage,
							SecurityContext: &corev1.SecurityContext{
								RunAsUser: &runAsUser,
							},
							Env: []corev1.EnvVar{
								{
									Name:  "KOLLA_CONFIG_STRATEGY",
									Value: "COPY_ALWAYS",
								},
								{
									Name:  "KOLLA_BOOTSTRAP",
									Value: "TRUE",
								},
							},
							VolumeMounts: options.VolumeMounts,
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:    fmt.Sprintf("%s-secrets", options.ServiceName),
							Image:   options.ContainerImage,
							Command: []string{"/bin/sh", "-c", passwordInitCmd},
							Env: []corev1.EnvVar{
								{
									Name:  "DatabaseHost",
									Value: options.DBOptions.DatabaseHostname,
								},
								{
									Name:  "DatabaseUser",
									Value: options.DBOptions.DatabaseUser,
								},
								{
									Name:  "DatabaseSchema",
									Value: options.DBOptions.DatabaseName,
								},
								{
									Name: "DatabasePassword",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: options.DBOptions.Secret,
											},
											Key: DatabaseUserPasswordKey,
										},
									},
								},
								{
									Name: "AdminPassword",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: options.DBOptions.Secret,
											},
											Key: DatabaseAdminPasswordKey,
										},
									},
								},
							},
							VolumeMounts: options.InitVolumeMounts,
						},
					},
				},
			},
		},
	}
	job.Spec.Template.Spec.Volumes = options.Volumes
	return job, nil
}
*/
