package test

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"strings"
	"testing"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gruntwork-io/terratest/modules/helm"
)

const (
	WaitTimerRetries = 60
	WaitTimerSleep   = 5 * time.Second
	NumPodsExpected  = 3  // TODO
)

func TestElkIntegration(t *testing.T) {
	t.Parallel()
	
	helmChartPath, err := filepath.Abs(filepath.Join("..", "charts", "elk"))
	require.NoError(t, err)

	kubectlOptions := k8s.NewKubectlOptions("", "", "")
	uniqueID := random.UniqueId()
	testNamespace := fmt.Sprintf("elk-stack-%s", strings.ToLower(uniqueID))
	k8s.CreateNamespace(t, kubectlOptions, testNamespace)
	defer k8s.DeleteNamespace(t, kubectlOptions, testNamespace)
	kubectlOptions.Namespace = testNamespace

	releaseName := fmt.Sprintf("elk-stack-%s", strings.ToLower(uniqueID))
	options := &helm.Options{
		KubectlOptions: kubectlOptions,
		ValuesFiles: []string{
			filepath.Join("..", "examples", "elk", "values.yaml"),
		},
	}

	defer helm.Delete(t, options, releaseName, true)
	helm.Install(t, options, helmChartPath, releaseName)

	// Uses label filters including gruntwork.io/deployment-type=canary to ensure the correct pods were created
	verifyElkPodsCreatedSuccessfully(t, kubectlOptions, "canary-test", releaseName)
}

func verifyElkPodsCreatedSuccessfully(
	t *testing.T,
	kubectlOptions *k8s.KubectlOptions,
	appName string,
	releaseName string,
) {

	// TODO
	filters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s,gruntwork.io/deployment-type=canary", appName, releaseName), 
	}

	k8s.WaitUntilNumPodsCreated(t, kubectlOptions, filters, NumPodsExpected, WaitTimerRetries, WaitTimerSleep)
	pods := k8s.ListPods(t, kubectlOptions, filters)

	for _, pod := range pods {
		k8s.WaitUntilPodAvailable(t, kubectlOptions, pod.Name, WaitTimerRetries, WaitTimerSleep)
	}

	// TODO
	mainDeploymentFilters := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s,gruntwork.io/deployment-type=main", appName, releaseName),
	}

	k8s.WaitUntilNumPodsCreated(t, kubectlOptions, mainDeploymentFilters, NumPodsExpected, WaitTimerRetries, WaitTimerSleep)
	mainPods := k8s.ListPods(t, kubectlOptions, mainDeploymentFilters)

	for _, mainPod := range mainPods {
		k8s.WaitUntilPodAvailable(t, kubectlOptions, mainPod.Name, WaitTimerRetries, WaitTimerSleep)
	}

}
