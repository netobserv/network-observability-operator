package e2etests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// processTemplate processes OpenShift template files
// Parameters format: "-f", "<file>", "-p", "KEY1=VALUE1", "KEY2=VALUE2", ...
// Returns path to processed YAML file containing the template objects
func processTemplate(namespace string, parameters ...string) (string, error) {
	if namespace == "" {
		namespace = "default"
	}
	var configFile string
	err := wait.PollUntilContextTimeout(context.Background(), 3*time.Second, 15*time.Second, false, func(context.Context) (bool, error) {
		outputFile := filepath.Join(os.TempDir(), getRandomString()+".json")
		args := []string{"process"}
		args = append(args, "-n", namespace)
		args = append(args, parameters...)
		cmd := exec.Command("oc", args...)
		output, err := cmd.Output()
		if err != nil {
			e2e.Logf("oc process failed: %v, and try next round", err)
			return false, nil
		}
		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			e2e.Logf("failed to write output file: %v", err)
			return false, nil
		}
		configFile = outputFile
		return true, nil
	})
	return configFile, err
}

// applyResourceFromTemplateByAdmin processes template and applies it
func applyResourceFromTemplateByAdmin(parameters ...string) error {
	return resourceFromTemplate("", parameters...)
}

// applyNsResourceFromTemplateByAdmin processes template and applies it to the specified namespace
func applyNsResourceFromTemplateByAdmin(namespace string, parameters ...string) error {
	return resourceFromTemplate(namespace, parameters...)
}

func resourceFromTemplate(namespace string, parameters ...string) error {
	configFile, err := processTemplate(namespace, parameters...)
	if err != nil {
		return fmt.Errorf("failed to process template %v: %w", parameters, err)
	}
	defer os.Remove(configFile)

	e2e.Logf("the file of resource is %s", configFile)

	var resourceErr error
	if namespace != "" {
		cmd := exec.Command("oc", "apply", "-f", configFile, "-n", namespace)
		output, err := cmd.CombinedOutput()
		if err != nil {
			resourceErr = fmt.Errorf("oc apply failed: %w, output: %s", err, string(output))
		}
	} else {
		cmd := exec.Command("oc", "apply", "-f", configFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			resourceErr = fmt.Errorf("oc apply failed: %w, output: %s", err, string(output))
		}
	}
	return resourceErr
}
