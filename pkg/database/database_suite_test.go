/*
Copyright 2018 The Kubernetes Authors.
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

package database_test

import (
	"context"
	"os"
	goruntime "runtime"
	"strings"
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func TestADatabase(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteName := "Database Suite"
	RunSpecsWithDefaultAndCustomReporters(t, suiteName, []Reporter{printer.NewlineReporter{}, printer.NewProwReporter(suiteName)})
	if err := c.Create(ctx, ns); err != nil {
		t.Fatalf("Create namespace error: (%v)", err)
	}
}

var env *envtest.Environment
var cfg *rest.Config
var c client.Client
var ctx context.Context

var ns = &corev1.Namespace{
	ObjectMeta: metav1.ObjectMeta{
		Name: "openstack",
	},
}

var _ = BeforeSuite(func() {
	Expect(os.Setenv("TEST_ASSET_KUBE_APISERVER", "/usr/local/kubebuilder/bin/kube-apiserver")).To(Succeed())
	Expect(os.Setenv("TEST_ASSET_ETCD", "/usr/local/kubebuilder/bin/etcd")).To(Succeed())
	Expect(os.Setenv("TEST_ASSET_KUBECTL", "/usr/local/kubebuilder/bin/kubectl")).To(Succeed())

	klog.InitFlags(nil)
	logger := klogr.New()
	log.SetLogger(logger)
	ctrl.SetLogger(logger)
	klog.SetOutput(ginkgo.GinkgoWriter)

	var err error

	//logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	env = &envtest.Environment{}
	ctx = ctrl.SetupSignalHandler()

	cfg, err = env.Start()
	Expect(err).NotTo(HaveOccurred())

	host := "localhost"
	if strings.ToLower(os.Getenv("USE_EXISTING_CLUSTER")) == "true" {
		// 0.0.0.0 is required on Linux when using kind because otherwise the kube-apiserver running in kind
		// is unable to reach the webhook, because the webhook would be only listening on 127.0.0.1.
		// Somehow that's not an issue on MacOS.
		if goruntime.GOOS == "linux" {
			host = "0.0.0.0"
		}
	}

	options := manager.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
		//CertDir:               env.WebhookInstallOptions.LocalServingCertDir,
		//Port:                  env.WebhookInstallOptions.LocalServingPort,
		//ClientDisableCacheFor: objs,
		Host: host,
	}

	mgr, err := ctrl.NewManager(env.Config, options)
	if err != nil {
		klog.Fatalf("Failed to start testenv manager: %v", err)
	}

	mgr.GetConfig()

	c = mgr.GetClient()
	//c, err = client.New(cfg, client.Options{})
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(env.Stop()).To(Succeed())
})
