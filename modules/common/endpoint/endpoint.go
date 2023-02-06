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

package endpoint

import (
	"context"
	"net/url"
	"strings"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/route"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Endpoint - typedef to enumerate Endpoint verbs
type Endpoint string

const (
	// EndpointAdmin - admin endpoint
	EndpointAdmin Endpoint = "admin"
	// EndpointInternal - internal endpoint
	EndpointInternal Endpoint = "internal"
	// EndpointPublic - public endpoint
	EndpointPublic Endpoint = "public"
)

// Data - information for generation of K8S services and Keystone endpoint URLs
type Data struct {
	// Used in k8s service definition
	Port int32
	// An optional path suffix to append to route hostname when forming Keystone endpoint URLs
	Path string
	// details for metallb service generation
	MetalLB *MetalLBData
}

// MetalLBData - information specific to creating the MetalLB service
type MetalLBData struct {
	// Name of the metallb IpAddressPool
	IPAddressPool string
	// use shared IP for the service
	SharedIP bool
	// if set request these IPs via MetalLBLoadBalancerIPs, using a list for dual stack (ipv4/ipv6)
	LoadBalancerIPs []string
}

// ExposeEndpoints - creates services, routes and returns a map of created openstack endpoint
func ExposeEndpoints(
	ctx context.Context,
	h *helper.Helper,
	serviceName string,
	endpointSelector map[string]string,
	endpoints map[Endpoint]Data,
) (map[string]string, ctrl.Result, error) {
	endpointMap := make(map[string]string)

	for endpointType, data := range endpoints {

		endpointName := serviceName + "-" + string(endpointType)
		exportLabels := util.MergeStringMaps(
			endpointSelector,
			map[string]string{
				string(endpointType): "true",
			},
		)

		// Create metallb service if specified, otherwise create a route
		var hostname string
		if data.MetalLB != nil {
			annotations := map[string]string{
				service.MetalLBAddressPoolAnnotation: data.MetalLB.IPAddressPool,
			}
			if len(data.MetalLB.LoadBalancerIPs) > 0 {
				annotations[service.MetalLBLoadBalancerIPs] = strings.Join(data.MetalLB.LoadBalancerIPs, ",")
			}
			if data.MetalLB.SharedIP {
				annotations[service.MetalLBAllowSharedIPAnnotation] = data.MetalLB.IPAddressPool + "-vip"
			}

			svc := service.NewService(
				service.MetalLBService(&service.MetalLBServiceDetails{
					Name:        endpointName,
					Namespace:   h.GetBeforeObject().GetNamespace(),
					Annotations: annotations,
					Labels:      exportLabels,
					Selector:    endpointSelector,
					Port: service.GenericServicePort{
						Name:     endpointName,
						Port:     data.Port,
						Protocol: corev1.ProtocolTCP,
					},
				}),
				exportLabels,
				5,
			)
			ctrlResult, err := svc.CreateOrPatch(ctx, h)
			if err != nil {
				return endpointMap, ctrlResult, err
			} else if (ctrlResult != ctrl.Result{}) {
				return endpointMap, ctrlResult, nil
			}
			// create service - end

			hostname = svc.GetServiceHostnamePort()
		} else {

			// Create the service if none exists
			svc := service.NewService(
				service.GenericService(&service.GenericServiceDetails{
					Name:      endpointName,
					Namespace: h.GetBeforeObject().GetNamespace(),
					Labels:    exportLabels,
					Selector:  endpointSelector,
					Port: service.GenericServicePort{
						Name:     endpointName,
						Port:     data.Port,
						Protocol: corev1.ProtocolTCP,
					}}),
				exportLabels,
				5,
			)
			ctrlResult, err := svc.CreateOrPatch(ctx, h)
			if err != nil {
				return endpointMap, ctrlResult, err
			} else if (ctrlResult != ctrl.Result{}) {
				return endpointMap, ctrlResult, nil
			}
			// create service - end

			hostname = svc.GetServiceHostnamePort()

			// Create the route if it is public endpoint
			if endpointType == EndpointPublic {
				// Create the route if none exists
				// TODO TLS
				route := route.NewRoute(
					route.GenericRoute(&route.GenericRouteDetails{
						Name:           endpointName,
						Namespace:      h.GetBeforeObject().GetNamespace(),
						Labels:         exportLabels,
						ServiceName:    endpointName,
						TargetPortName: endpointName,
					}),
					exportLabels,
					5,
				)

				ctrlResult, err = route.CreateOrPatch(ctx, h)
				if err != nil {
					return endpointMap, ctrlResult, err
				} else if (ctrlResult != ctrl.Result{}) {
					return endpointMap, ctrlResult, nil
				}
				// create route - end

				hostname = route.GetHostname()
			}
		}

		// Update instance status with service endpoint url from route host information
		var protocol string

		// TODO: need to support https default here
		if !strings.HasPrefix(hostname, "http") {
			protocol = "http://"
		} else {
			protocol = ""
		}

		// Do not include data.Path in parsing check because %(project_id)s
		// is invalid without being encoded, but they should not be encoded in the actual endpoint
		apiEndpoint, err := url.Parse(protocol + hostname)
		if err != nil {
			return endpointMap, ctrl.Result{}, err
		}
		endpointMap[string(endpointType)] = apiEndpoint.String() + data.Path
	}

	return endpointMap, ctrl.Result{}, nil
}
