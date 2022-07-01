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

package openstack

import (
	"fmt"

	"github.com/go-logr/logr"
	services "github.com/gophercloud/gophercloud/openstack/identity/v3/services"

	appsv1 "k8s.io/api/apps/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

// Service -
type Service struct {
	Name        string
	Type        string
	Description string
	Enabled     bool
}

//
// CreateService - create service
//
func (o *OpenStack) CreateService(
	log logr.Logger,
	s Service,
) (string, error) {

	createOpts := services.CreateOpts{
		Type:    s.Type,
		Enabled: &s.Enabled,
		Extra: map[string]interface{}{
			"name":        s.Name,
			"description": s.Description,
		},
	}

	service, err := services.Create(o.GetOSClient(), createOpts).Extract()
	if err != nil {
		return "", err
	}

	return service.ID, nil
}

//
// GetService - get service with type and name
//
func (o *OpenStack) GetService(
	log logr.Logger,
	serviceType string,
	serviceName string,
) (*services.Service, error) {
	listOpts := services.ListOpts{
		ServiceType: serviceType,
		Name:        serviceName,
	}

	allPages, err := services.List(o.osclient, listOpts).AllPages()
	if err != nil {
		return nil, err
	}
	allServices, err := services.ExtractServices(allPages)
	if err != nil {
		return nil, err
	}

	if len(allServices) == 0 {
		return nil, k8s_errors.NewNotFound(
			appsv1.Resource("Services"),
			fmt.Sprintf("%s service not found in keystone", serviceName),
		)
	}

	return &allServices[0], nil
}

//
// UpdateService - update service with type and name
//
func (o *OpenStack) UpdateService(
	log logr.Logger,
	s Service,
	serviceID string,
) error {
	updateOpts := services.UpdateOpts{
		Type:    s.Type,
		Enabled: &s.Enabled,
		Extra: map[string]interface{}{
			"name":        s.Name,
			"description": s.Description,
		},
	}
	_, err := services.Update(o.GetOSClient(), serviceID, updateOpts).Extract()
	if err != nil {
		return err
	}
	return nil
}

//
// DeleteService - delete service with serviceID
//
func (o *OpenStack) DeleteService(
	log logr.Logger,
	serviceID string,
) error {
	log.Info(fmt.Sprintf("Delete service with id %s", serviceID))
	err := services.Delete(o.GetOSClient(), serviceID).ExtractErr()
	if err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}

	return nil
}
