/*
Copyright 2024.

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

package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha1 "github.com/giantswarm/organization-operator/api/v1alpha1"
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	// Debug: Print current working directory
	pwd, _ := os.Getwd()
	fmt.Printf("Current working directory: %s\n", pwd)

	// Debug: Print KUBEBUILDER_ASSETS
	fmt.Printf("KUBEBUILDER_ASSETS: %s\n", os.Getenv("KUBEBUILDER_ASSETS"))

	// Check if KUBEBUILDER_ASSETS is set, if not, try to set it
	if os.Getenv("KUBEBUILDER_ASSETS") == "" {
		possiblePaths := []string{
			filepath.Join("..", "..", "testbin", "bin"),
			filepath.Join("..", "..", "bin"),
			"/usr/local/kubebuilder/bin",
			os.Getenv("HOME") + "/go/bin",
		}

		for _, path := range possiblePaths {
			fmt.Printf("Checking path: %s\n", path)
			if _, err := os.Stat(filepath.Join(path, "etcd")); err == nil {
				os.Setenv("KUBEBUILDER_ASSETS", path)
				fmt.Printf("Set KUBEBUILDER_ASSETS to: %s\n", path)
				break
			}
		}
	}

	// Debug: Print contents of KUBEBUILDER_ASSETS
	if assetPath := os.Getenv("KUBEBUILDER_ASSETS"); assetPath != "" {
		fmt.Printf("Contents of KUBEBUILDER_ASSETS (%s):\n", assetPath)
		files, _ := os.ReadDir(assetPath)
		for _, file := range files {
			fmt.Println(file.Name())
		}
	} else {
		fmt.Println("KUBEBUILDER_ASSETS is not set")
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = corev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	if testEnv != nil {
		err := testEnv.Stop()
		Expect(err).NotTo(HaveOccurred())
	}
})
