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
	gophercloud "github.com/gophercloud/gophercloud"
	endpoints "github.com/gophercloud/gophercloud/openstack/identity/v3/endpoints"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

// Endpoint -
type Endpoint struct {
	Name         string
	ServiceID    string
	Availability gophercloud.Availability
	URL          string
}

//
// CreateEndpoint - create endpoint
//
func (o *OpenStack) CreateEndpoint(
	log logr.Logger,
	e Endpoint,
) (string, error) {

	// validate if endpoint already exist
	allEndpoints, err := o.GetEndpoints(
		log,
		e.ServiceID,
		string(e.Availability))
	if err != nil {
		return "", err
	}

	if len(allEndpoints) > 0 {
		return allEndpoints[0].ID, nil
	}

	// Create the endpoint
	createOpts := endpoints.CreateOpts{
		Availability: e.Availability,
		Name:         e.Name,
		Region:       o.region,
		ServiceID:    e.ServiceID,
		URL:          e.URL,
	}
	createdEndpoint, err := endpoints.Create(o.osclient, createOpts).Extract()
	if err != nil {
		return "", err
	}
	return createdEndpoint.ID, nil
}

//
// GetEndpoints - get endpoints for the registered service. if endpointInterface
// is provided, just return the endpoint for that type.
//
func (o *OpenStack) GetEndpoints(
	log logr.Logger,
	serviceID string,
	endpointInterface string,
) ([]endpoints.Endpoint, error) {
	log.Info(fmt.Sprintf("Getting Endpoints for service %s %s ", serviceID, endpointInterface))

	listOpts := endpoints.ListOpts{
		ServiceID: serviceID,
		RegionID:  o.region,
	}
	if endpointInterface != "" {
		availability, err := GetAvailability(endpointInterface)
		if err != nil {
			return nil, err
		}

		listOpts.Availability = availability
	}

	allPages, err := endpoints.List(o.osclient, listOpts).AllPages()
	if err != nil {
		return nil, err
	}
	allEndpoints, err := endpoints.ExtractEndpoints(allPages)
	if err != nil {
		return nil, err
	}

	log.Info("Getting Endpoint successfully")

	return allEndpoints, nil
}

//
// DeleteEndpoint - delete endpoint
//
func (o *OpenStack) DeleteEndpoint(
	log logr.Logger,
	e Endpoint,
) error {
	log.Info(fmt.Sprintf("Deleting Endpoint %s %s ", e.Name, e.Availability))

	// get all registered endpoints for the service/endpointInterface
	allEndpoints, err := o.GetEndpoints(log, e.ServiceID, string(e.Availability))
	if err != nil {
		return err
	}

	for _, endpt := range allEndpoints {
		log.Info(fmt.Sprintf("Delete endpoint %s %s - %s", endpt.Name, string(endpt.Availability), endpt.URL))
		err = endpoints.Delete(o.osclient, endpt.ID).ExtractErr()

		if err != nil && !k8s_errors.IsNotFound(err) {
			return err
		}
	}

	log.Info("Deleting Endpoint successfully")
	return nil
}

//
// UpdateEndpoint -
//
func (o *OpenStack) UpdateEndpoint(
	log logr.Logger,
	e Endpoint,
	endpointID string,
) (string, error) {
	log.Info(fmt.Sprintf("Updating Endpoint %s %s ", e.Name, e.Availability))

	// Update the endpoint
	updateOpts := endpoints.UpdateOpts{
		Availability: e.Availability,
		Name:         e.Name,
		Region:       o.region,
		ServiceID:    e.ServiceID,
		URL:          e.URL,
	}
	endpt, err := endpoints.Update(o.osclient, endpointID, updateOpts).Extract()
	if err != nil {
		return "", err
	}

	log.Info("Updating Endpoint successfully")
	return endpt.ID, nil
}
