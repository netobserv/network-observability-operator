package e2etests

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	filePath "path/filepath"
	"reflect"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

type User struct {
	Username string
	Password string
}

func getCoStatus(coName string, statusToCompare map[string]string) map[string]string {
	newStatusToCompare := make(map[string]string)
	obj, err := getDynamicResource("clusteroperator", coName, "")
	o.Expect(err).NotTo(o.HaveOccurred())
	conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	for key := range statusToCompare {
		for _, c := range conditions {
			cond, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if t, _ := cond["type"].(string); t == key {
				newStatusToCompare[key], _ = cond["status"].(string)
				break
			}
		}
	}
	return newStatusToCompare
}

func waitCoBecomes(coName string, waitTime int, expectedStatus map[string]string) error {
	errCo := wait.PollUntilContextTimeout(context.Background(), 10*time.Second, time.Duration(waitTime)*time.Second, false, func(context.Context) (bool, error) {
		gottenStatus := getCoStatus(coName, expectedStatus)
		eq := reflect.DeepEqual(expectedStatus, gottenStatus)
		if eq {
			eq := reflect.DeepEqual(expectedStatus, map[string]string{"Available": "True", "Progressing": "False", "Degraded": "False"})
			if eq {
				// For True False False, we want to wait some bit more time and double check, to ensure it is stably healthy
				time.Sleep(25 * time.Second)
				gottenStatus := getCoStatus(coName, expectedStatus)
				eq := reflect.DeepEqual(expectedStatus, gottenStatus)
				if eq {
					e2e.Logf("Given operator %s becomes available/non-progressing/non-degraded +%v", coName, gottenStatus)
					return true, nil
				}
			} else {
				e2e.Logf("Given operator %s becomes %s", coName, gottenStatus)
				return true, nil
			}
		}
		return false, nil
	})
	if errCo != nil {
		coGVR, _ := resolveGVR("clusteroperator")
		coList, err := k8sDynClient.Resource(coGVR).List(context.Background(), metav1.ListOptions{})
		if err == nil {
			var coNames []string
			for _, item := range coList.Items {
				name := item.GetName()
				status, _, _ := getConditionStatus(&item, "Available")
				progressing, _, _ := getConditionStatus(&item, "Progressing")
				degraded, _, _ := getConditionStatus(&item, "Degraded")
				coNames = append(coNames, fmt.Sprintf("%s(Available=%s,Progressing=%s,Degraded=%s)", name, status, progressing, degraded))
			}
			e2e.Logf("ClusterOperators: %v", strings.Join(coNames, ", "))
		}
	}
	return errCo
}

func generateUsersHtpasswd(passwdFile *string, users []*User) error {
	for i := 0; i < len(users); i++ {
		// Generate new username and password
		username := fmt.Sprintf("testuser-%v-%v", i, getRandomString())
		password := getRandomString()
		users[i] = &User{Username: username, Password: password}

		// Add new user to htpasswd file in the temp directory
		cmd := fmt.Sprintf("htpasswd -b %v %v %v", *passwdFile, users[i].Username, users[i].Password)
		err := exec.Command("bash", "-c", cmd).Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func getNewUser(count int) ([]*User, string, string) {
	usersDirPath := "/tmp/" + getRandomString()
	usersHTpassFile := usersDirPath + "/htpasswd"
	err := os.MkdirAll(usersDirPath, 0o755)
	o.Expect(err).NotTo(o.HaveOccurred())

	htPassSecret, err := getOAuthHTPasswdSecretName()
	o.Expect(err).NotTo(o.HaveOccurred())
	users := make([]*User, count)
	if htPassSecret == "" {
		htPassSecret = "htpass-secret"
		_, _ = os.Create(usersHTpassFile)
		err = generateUsersHtpasswd(&usersHTpassFile, users)
		o.Expect(err).NotTo(o.HaveOccurred())
		err = createSecretFromFile("openshift-config", htPassSecret, "htpasswd", usersHTpassFile)
		o.Expect(err).NotTo(o.HaveOccurred())
		err = patchOAuthAddHTPasswdIdentityProvider(htPassSecret)
		o.Expect(err).NotTo(o.HaveOccurred())
	} else {
		err = extractSecretToFile("openshift-config", htPassSecret, "htpasswd", usersHTpassFile)
		o.Expect(err).NotTo(o.HaveOccurred())
		err = generateUsersHtpasswd(&usersHTpassFile, users)
		o.Expect(err).NotTo(o.HaveOccurred())
		// Update htpass-secret with the modified htpasswd file
		err = updateSecretFromFile("openshift-config", htPassSecret, "htpasswd", usersHTpassFile)
		o.Expect(err).NotTo(o.HaveOccurred())
	}

	g.By("Checking authentication operator should be in Progressing in 180 seconds")
	err = waitCoBecomes("authentication", 180, map[string]string{"Progressing": "True"})
	assertWaitPollNoErr(err, "authentication operator did not start progressing in 180 seconds")
	e2e.Logf("Checking authentication operator should be Available in 600 seconds")
	err = waitCoBecomes("authentication", 600, map[string]string{"Available": "True", "Progressing": "False", "Degraded": "False"})
	assertWaitPollNoErr(err, "authentication operator did not become available in 600 seconds")

	return users, usersHTpassFile, htPassSecret
}

func userCleanup(users []*User, usersHTpassFile string, htPassSecret string) {
	defer os.RemoveAll(usersHTpassFile)
	for i := range users {
		// Add new user to htpasswd file in the temp directory
		cmd := fmt.Sprintf("htpasswd -D %v %v", usersHTpassFile, users[i].Username)
		err := exec.Command("bash", "-c", cmd).Run()
		o.Expect(err).NotTo(o.HaveOccurred())
	}

	// Update htpass-secret with the modified htpasswd file
	err := updateSecretFromFile("openshift-config", htPassSecret, "htpasswd", usersHTpassFile)
	o.Expect(err).NotTo(o.HaveOccurred())

	g.By("Checking authentication operator should be in Progressing in 180 seconds")
	err = waitCoBecomes("authentication", 180, map[string]string{"Progressing": "True"})
	assertWaitPollNoErr(err, "authentication operator did not start progressing in 180 seconds")
	e2e.Logf("Checking authentication operator should be Available in 600 seconds")
	err = waitCoBecomes("authentication", 600, map[string]string{"Available": "True", "Progressing": "False", "Degraded": "False"})
	assertWaitPollNoErr(err, "authentication operator did not become available in 600 seconds")
}

func addUserAsReader(username string) {
	baseDir, _ := filePath.Abs("testdata")
	readerCRBPath := filePath.Join(baseDir, "netobserv-loki-reader-multitenant-crb.yaml")
	parameters := []string{"--ignore-unknown-parameters=true", "-f", readerCRBPath, "-p", "USERNAME=" + username}
	err := applyResourceFromTemplateByAdmin(parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

func removeUserAsReader(username string) {
	cmd := exec.Command("oc", "adm", "policy", "remove-cluster-role-from-user", "netobserv-loki-reader", username)
	err := cmd.Run()
	o.Expect(err).NotTo(o.HaveOccurred())
}

func addTemplatePermissions(username string) {
	baseDir, _ := filePath.Abs("testdata")
	readerCRBPath := filePath.Join(baseDir, "testuser-template-crb.yaml")
	parameters := []string{"--ignore-unknown-parameters=true", "-f", readerCRBPath, "-p", "USERNAME=" + username}
	err := applyResourceFromTemplateByAdmin(parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

func generateOAuthTokenPair() (string, string) {
	const sha256Prefix = "sha256~"
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	o.Expect(err).NotTo(o.HaveOccurred())
	randomToken := base64.RawURLEncoding.EncodeToString(randomBytes)
	hashed := sha256.Sum256([]byte(randomToken))
	return sha256Prefix + randomToken, sha256Prefix + base64.RawURLEncoding.EncodeToString(hashed[:])
}

func changeUser(user *User, namespace string) string {
	serverURL, err := getServerURL()
	o.Expect(err).NotTo(o.HaveOccurred())

	userUIDBytes, err := exec.Command("oc", "get", "user", user.Username, "-o", "jsonpath={.metadata.uid}").Output()
	if err != nil {
		_, err = exec.Command("oc", "create", "user", user.Username).Output()
		o.Expect(err).NotTo(o.HaveOccurred(), fmt.Sprintf("failed to create user %s", user.Username))
		userUIDBytes, err = exec.Command("oc", "get", "user", user.Username, "-o", "jsonpath={.metadata.uid}").Output()
		o.Expect(err).NotTo(o.HaveOccurred())
	}
	userUID := strings.TrimSpace(string(userUIDBytes))

	oauthClientName := "e2e-client-" + namespace
	oauthClientJSON := fmt.Sprintf(`{"apiVersion":"oauth.openshift.io/v1","kind":"OAuthClient","metadata":{"name":"%s"},"grantMethod":"auto"}`, oauthClientName)
	cmd := exec.Command("oc", "create", "-f", "-")
	cmd.Stdin = strings.NewReader(oauthClientJSON)
	_ = cmd.Run()

	privToken, pubToken := generateOAuthTokenPair()

	tokenJSON := fmt.Sprintf(`{"apiVersion":"oauth.openshift.io/v1","kind":"OAuthAccessToken","metadata":{"name":"%s"},"clientName":"%s","userName":"%s","userUID":"%s","scopes":["user:full"],"redirectURI":"https://localhost:8443/oauth/token/implicit"}`, pubToken, oauthClientName, user.Username, userUID)
	cmd = exec.Command("oc", "create", "-f", "-")
	cmd.Stdin = strings.NewReader(tokenJSON)
	output, err := cmd.CombinedOutput()
	o.Expect(err).NotTo(o.HaveOccurred(), fmt.Sprintf("failed to create OAuthAccessToken: %s", string(output)))

	err = exec.Command("oc", "config", "set-credentials", user.Username, "--token="+privToken).Run()
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to set credentials for user "+user.Username)

	clusterNameBytes, err := exec.Command("oc", "config", "view", "-o", "jsonpath={.clusters[0].name}").Output()
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to get cluster name")
	clusterName := strings.TrimSpace(string(clusterNameBytes))

	userContext := fmt.Sprintf("%s/%s", user.Username, serverURL)
	err = exec.Command("oc", "config", "set-context", userContext, "--cluster="+clusterName, "--user="+user.Username, "--namespace="+namespace).Run()
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to set context for user "+user.Username)
	err = exec.Command("oc", "config", "use-context", userContext).Run()
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to switch to context "+userContext)

	e2e.Logf("Changed to user %s with context %s", user.Username, userContext)
	return userContext
}

func removeTemplatePermissions(username string) {
	baseDir, _ := filePath.Abs("testdata")
	readerCRBPath := filePath.Join(baseDir, "testuser-template-crb.yaml")
	parameters := []string{"-f", readerCRBPath, "-p", "USERNAME=" + username}
	configFile, err := processTemplate("", parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
	cmd := exec.Command("oc", "delete", "-f", configFile)
	err = cmd.Run()
	o.Expect(err).NotTo(o.HaveOccurred())
}
