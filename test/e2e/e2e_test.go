package e2e

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/giantswarm/organization-operator/test/utils"
)

const (
	namespace    = "new-system"
	makeCmd      = "make"
	kubectlCmd   = "kubectl"
	projectImage = "example.com/new:v0.0.1"
)

var validImageNameRegex = regexp.MustCompile(`^[\w.\-/:]+$`)

func safeExecCommand(name string, arg ...string) *exec.Cmd {
	// Validate the command name
	if !filepath.IsAbs(name) {
		name, _ = exec.LookPath(name)
	}
	return exec.Command(name, arg...)
}
func validateImageName(name string) error {
	if !validImageNameRegex.MatchString(name) {
		return fmt.Errorf("invalid image name: %s", name)
	}
	return nil
}

var _ = Describe("controller", Ordered, func() {
	BeforeAll(func() {
		By("installing prometheus operator")
		Expect(utils.InstallPrometheusOperator()).To(Succeed())

		By("installing the cert-manager")
		Expect(utils.InstallCertManager()).To(Succeed())

		By("creating manager namespace")
		cmd := safeExecCommand(kubectlCmd, "create", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	AfterAll(func() {
		By("uninstalling the Prometheus manager bundle")
		utils.UninstallPrometheusOperator()

		By("uninstalling the cert-manager bundle")
		utils.UninstallCertManager()

		By("removing manager namespace")
		cmd := safeExecCommand(kubectlCmd, "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	Context("Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error

			By("validating the project image name")
			err = validateImageName(projectImage)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("building the manager(Operator) image")
			cmd := safeExecCommand(makeCmd, "docker-build", fmt.Sprintf("IMG=%s", projectImage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the manager(Operator) image on Kind")
			err = utils.LoadImageToKindClusterWithName(projectImage)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("installing CRDs")
			cmd = safeExecCommand(makeCmd, "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = safeExecCommand(makeCmd, "deploy", fmt.Sprintf("IMG=%s", projectImage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				cmd = safeExecCommand(kubectlCmd, "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pod running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = safeExecCommand(kubectlCmd, "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())
		})
	})
})
