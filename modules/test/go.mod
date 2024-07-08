module github.com/openstack-k8s-operators/lib-common/modules/test

go 1.20

require (
	github.com/go-logr/logr v1.4.2
	github.com/onsi/gomega v1.33.1
	golang.org/x/mod v0.17.0
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/onsi/ginkgo/v2 v2.19.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/openstack-k8s-operators/lib-common/modules/common => ../common

replace github.com/openstack-k8s-operators/lib-common/modules/openstack => ../openstack

// mschuppert: map to latest commit from release-4.16 tag
// must consistent within modules and service operators
replace github.com/openshift/api => github.com/openshift/api v0.0.0-20240702171116-4b89b3a92a17
