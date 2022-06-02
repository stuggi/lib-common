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
	// DatabaseUserPasswordKey - key in secret which holds the service user DB password
	DatabaseUserPasswordKey = "DatabasePassword"
	// DatabaseAdminPasswordKey - key in secret which holds the admin user password
	DatabaseAdminPasswordKey = "AdminPassword"
)

// Database -
type Database struct {
	databaseHostname string
	databaseName     string
	databaseUser     string
	secret           string
	labels           map[string]string
}

/*
// DBSyncOptions -
type DBSyncOptions struct {
	ServiceName      string
	Namespace        string
	ServiceAccount   string // e.g. keystone-operator-keystone
	ContainerImage   string
	VolumeMounts     []corev1.VolumeMount
	InitVolumeMounts []corev1.VolumeMount
	Volumes          []corev1.Volume
	DBOptions        Database

	DatabaseHostname string
	DatabaseName     string
	Secret           string
	Labels           map[string]string // app: keystone-api
}
*/
