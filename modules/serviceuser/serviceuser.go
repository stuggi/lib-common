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

// Package serviceuser provides a central registry of OpenStack service
// UIDs/GIDs. Go operators import the constants directly; non-Go consumers
// (e.g. container image builds) use the generated YAML file.
package serviceuser

// ServiceUser defines a service user for container image builds.
type ServiceUser struct {
	UID    int64    `yaml:"uid"`
	GID    int64    `yaml:"gid"`
	Home   string   `yaml:"home"`
	Groups []string `yaml:"groups,omitempty"`
}

// Service UID/GID constants.
//
// WARNING: Do not change existing values. Services that write to persistent
// storage (nova, glance, cinder, swift, galera) have files owned by these
// UIDs on PersistentVolumes. Changing a UID would require a manual chown
// migration on every PV in every deployment.
const (
	AodhUID            int64 = 42402
	AodhGID            int64 = 42402
	BarbicanUID        int64 = 42403
	BarbicanGID        int64 = 42403
	CeilometerUID      int64 = 42405
	CeilometerGID      int64 = 42405
	CloudkittyUID      int64 = 42406
	CloudkittyGID      int64 = 42406
	CinderUID          int64 = 42407
	CinderGID          int64 = 42407
	DesignateUID       int64 = 42411
	DesignateGID       int64 = 42411
	GlanceUID          int64 = 42415
	GlanceGID          int64 = 42415
	HeatUID            int64 = 42418
	HeatGID            int64 = 42418
	HorizonUID         int64 = 42420
	HorizonGID         int64 = 42420
	IronicUID          int64 = 42422
	IronicGID          int64 = 42422
	IronicInspectorUID int64 = 42461
	IronicInspectorGID int64 = 42461
	KeystoneUID        int64 = 42425
	KeystoneGID        int64 = 42425
	ManilaUID          int64 = 42429
	ManilaGID          int64 = 42429
	MemcachedUID       int64 = 42457
	MemcachedGID       int64 = 42457
	MysqlUID           int64 = 42434
	MysqlGID           int64 = 42434
	NeutronUID         int64 = 42435
	NeutronGID         int64 = 42435
	NovaUID            int64 = 42436
	NovaGID            int64 = 42436
	OctaviaUID         int64 = 42437
	OctaviaGID         int64 = 42437
	PlacementUID       int64 = 42482
	PlacementGID       int64 = 42482
	RabbitmqUID        int64 = 42439
	RabbitmqGID        int64 = 42439
	RedisUID           int64 = 42460
	RedisGID           int64 = 42460
	SwiftUID           int64 = 42445
	SwiftGID           int64 = 42445
	WatcherUID         int64 = 42451
	WatcherGID         int64 = 42451
)

// ApacheGID is the GID of the "apache" system group baked into the
// RHEL/CentOS httpd package used by OpenStack service images. Some
// RPM-shipped httpd conf.d files (e.g. mod_auth_openidc's
// auth_openidc.conf) are group-owned by apache with restrictive
// permissions (0640). These files are baked into the container image
// rather than volume-mounted, so FSGroup does not apply to them.
//
// Container image builds do not currently add OpenStack service users to
// the apache group, so httpd-based workloads running under a restrictive,
// non-root SecurityContext need apache granted as a supplemental group
// explicitly — see pod.RestrictivePodSecurityContextWithGroups in
// modules/common/pod.
const ApacheGID int64 = 48

// Registry maps service names to their user definitions. This is the
// source of truth — the YAML generator (gen_uid_gid_yaml.go) marshals
// this to zz_generated_uid_gid.yaml for container image builds.
// Read-only at runtime; do not modify.
var Registry = map[string]ServiceUser{
	"aodh":             {UID: AodhUID, GID: AodhGID, Home: "/var/lib/aodh"},
	"barbican":         {UID: BarbicanUID, GID: BarbicanGID, Home: "/var/lib/barbican"},
	"ceilometer":       {UID: CeilometerUID, GID: CeilometerGID, Home: "/var/lib/ceilometer"},
	"cloudkitty":       {UID: CloudkittyUID, GID: CloudkittyGID, Home: "/var/lib/cloudkitty"},
	"cinder":           {UID: CinderUID, GID: CinderGID, Home: "/var/lib/cinder"},
	"designate":        {UID: DesignateUID, GID: DesignateGID, Home: "/var/lib/designate"},
	"glance":           {UID: GlanceUID, GID: GlanceGID, Home: "/var/lib/glance"},
	"heat":             {UID: HeatUID, GID: HeatGID, Home: "/var/lib/heat"},
	"horizon":          {UID: HorizonUID, GID: HorizonGID, Home: "/var/lib/horizon"},
	"ironic":           {UID: IronicUID, GID: IronicGID, Home: "/var/lib/ironic"},
	"ironic-inspector": {UID: IronicInspectorUID, GID: IronicInspectorGID, Home: "/var/lib/ironic-inspector"},
	"keystone":         {UID: KeystoneUID, GID: KeystoneGID, Home: "/var/lib/keystone", Groups: []string{"apache"}},
	"manila":           {UID: ManilaUID, GID: ManilaGID, Home: "/var/lib/manila"},
	"memcached":        {UID: MemcachedUID, GID: MemcachedGID, Home: "/var/lib/memcached"},
	"mysql":            {UID: MysqlUID, GID: MysqlGID, Home: "/var/lib/mysql"},
	"neutron":          {UID: NeutronUID, GID: NeutronGID, Home: "/var/lib/neutron"},
	"nova":             {UID: NovaUID, GID: NovaGID, Home: "/var/lib/nova", Groups: []string{"libvirt", "qemu"}},
	"octavia":          {UID: OctaviaUID, GID: OctaviaGID, Home: "/var/lib/octavia"},
	"placement":        {UID: PlacementUID, GID: PlacementGID, Home: "/var/lib/placement"},
	"rabbitmq":         {UID: RabbitmqUID, GID: RabbitmqGID, Home: "/var/lib/rabbitmq"},
	"redis":            {UID: RedisUID, GID: RedisGID, Home: "/var/lib/redis"},
	"swift":            {UID: SwiftUID, GID: SwiftGID, Home: "/var/lib/swift"},
	"watcher":          {UID: WatcherUID, GID: WatcherGID, Home: "/var/lib/watcher"},
}
