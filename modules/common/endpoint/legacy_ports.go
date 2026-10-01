/*
Copyright 2026 Red Hat

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

// LegacyPublicTLSPorts holds the RHOSP/director public TLS front ports
// (tripleo-heat-templates wallaby environments/ssl/tls-endpoints-public-ip.yaml),
// used for RHOSP 17.1 -> RHOSO 18 endpoint-URL preservation. Keyed by the RHOSO
// service ServiceName const (confirmed against each operator's const.go, not
// the THT endpoint name - e.g. "heat-cfnapi", not "heat-cfn"; "nova-novncproxy",
// not "novncproxy").
var LegacyPublicTLSPorts = map[string]int32{
	"keystone":         13000,
	"placement":        13778,
	"nova":             13774,
	"glance":           13292,
	"cinder":           13776,
	"neutron":          13696,
	"heat-api":         13004,
	"heat-cfnapi":      13005,
	"swift":            13808,
	"barbican":         13311,
	"octavia":          13876,
	"manila":           13786,
	"ironic":           13385,
	"ironic-inspector": 13050,
	"designate":        13001,
	"aodh":             13042,
	"nova-novncproxy":  13080,
	"horizon":          443,
}

// GetLegacyPublicTLSPort returns the director-era public TLS port for
// serviceName and whether an entry exists.
func GetLegacyPublicTLSPort(serviceName string) (int32, bool) {
	port, ok := LegacyPublicTLSPorts[serviceName]
	return port, ok
}
