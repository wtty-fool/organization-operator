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

//nolint:gosec
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

//nolint:gosec
package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/organization-operator/test/utils"
)

const (
	namespace    = "organization-operator-system"
	projectImage = "example.com/organization-operator:v0.0.1"
)

var _ = Describe("Organization Operator", Ordered, func() {
	BeforeAll(func() {
		By("Installing Prometheus operator")
		Expect(utils.InstallPrometheusOperator()).To(Succeed())

		By("Installing the cert-manager")
		Expect(utils.InstallCertManager()).To(Succeed())

		By("Creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		By("Uninstalling the Prometheus manager bundle")
		utils.UninstallPrometheusOperator()

		By("Uninstalling the cert-manager bundle")
		utils.UninstallCertManager()

		By("Removing manager namespace")
		cmd := exec.Command("kubectl", "delete", "ns", namespace)
		_, err := utils.Run(cmd)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Operator Deployment", func() {
		It("should deploy and run successfully", func() {
			By("Building the manager (Operator) image")
			cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectImage))
			_, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("Loading the manager (Operator) image on Kind")
			err = utils.LoadImageToKindClusterWithName(projectImage)
			Expect(err).NotTo(HaveOccurred())

			By("Installing CRDs")
			cmd = exec.Command("make", "install")
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("Deploying the controller-manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectImage))
			_, err = utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("Validating that the controller-manager pod is running as expected")
			Eventually(verifyControllerUp, 2*time.Minute, 5*time.Second).Should(Succeed())
		})
	})
})

func verifyControllerUp() error {
	// Get pod name
	cmd := exec.Command("kubectl", "get",
		"pods", "-l", "control-plane=controller-manager",
		"-o", "go-template={{ range .items }}"+
			"{{ if not .metadata.deletionTimestamp }}"+
			"{{ .metadata.name }}"+
			"{{ \"\\n\" }}{{ end }}{{ end }}",
		"-n", namespace,
	)
	podOutput, err := utils.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to get controller pod: %v", err)
	}

	podNames := utils.GetNonEmptyLines(string(podOutput))
	if len(podNames) != 1 {
		return fmt.Errorf("expected 1 controller pod running, but got %d", len(podNames))
	}

	controllerPodName := podNames[0]
	if !strings.Contains(controllerPodName, "controller-manager") {
		return fmt.Errorf("controller pod name %s doesn't contain 'controller-manager'", controllerPodName)
	}

	// Validate pod status
	cmd = exec.Command("kubectl", "get",
		"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
		"-n", namespace,
	)
	status, err := utils.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to get controller pod status: %v", err)
	}

	if string(status) != "Running" {
		return fmt.Errorf("controller pod in %s status", status)
	}

	return nil
}
