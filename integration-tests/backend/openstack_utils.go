package e2etests

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	tokens3 "github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/objects"
	"github.com/gophercloud/gophercloud/pagination"
	o "github.com/onsi/gomega"
	yamlv3 "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// openstackCredentials represents OpenStack credentials extracted from cluster
type openstackCredentials struct {
	Clouds struct {
		Openstack struct {
			Auth struct {
				AuthURL                     string `yaml:"auth_url"`
				Password                    string `yaml:"password"`
				ProjectID                   string `yaml:"project_id"`
				ProjectName                 string `yaml:"project_name"`
				UserDomainName              string `yaml:"user_domain_name"`
				Username                    string `yaml:"username"`
				ApplicationCredentialID     string `yaml:"application_credential_id"`
				ApplicationCredentialSecret string `yaml:"application_credential_secret"`
			} `yaml:"auth"`
			EndpointType       string `yaml:"endpoint_type"`
			IdentityAPIVersion string `yaml:"identity_api_version"`
			RegionName         string `yaml:"region_name"`
			Verify             bool   `yaml:"verify"`
		} `yaml:"openstack"`
	} `yaml:"clouds"`
}

// getOpenStackCredentials gets credentials from cluster secret
func getOpenStackCredentials() (*openstackCredentials, error) {
	cred := &openstackCredentials{}
	secret, err := k8sClient.CoreV1().Secrets("kube-system").Get(context.Background(), "openstack-credentials", metav1.GetOptions{})
	if err != nil {
		return cred, err
	}
	err = yamlv3.Unmarshal(secret.Data["clouds.yaml"], cred)
	return cred, err
}

// newOpenStackClient creates a new OpenStack service client
func newOpenStackClient(cred *openstackCredentials, serviceType string) *gophercloud.ServiceClient {
	var client *gophercloud.ServiceClient
	var opts gophercloud.AuthOptions

	if cred.Clouds.Openstack.Auth.ApplicationCredentialID != "" && cred.Clouds.Openstack.Auth.ApplicationCredentialSecret != "" {
		opts = gophercloud.AuthOptions{
			IdentityEndpoint:            cred.Clouds.Openstack.Auth.AuthURL,
			ApplicationCredentialID:     cred.Clouds.Openstack.Auth.ApplicationCredentialID,
			ApplicationCredentialSecret: cred.Clouds.Openstack.Auth.ApplicationCredentialSecret,
		}
	} else {
		opts = gophercloud.AuthOptions{
			IdentityEndpoint: cred.Clouds.Openstack.Auth.AuthURL,
			Username:         cred.Clouds.Openstack.Auth.Username,
			Password:         cred.Clouds.Openstack.Auth.Password,
			TenantID:         cred.Clouds.Openstack.Auth.ProjectID,
			DomainName:       cred.Clouds.Openstack.Auth.UserDomainName,
		}
	}

	provider, err := openstack.AuthenticatedClient(opts)
	o.Expect(err).NotTo(o.HaveOccurred())

	switch serviceType {
	case "identity":
		client, err = openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{Region: cred.Clouds.Openstack.RegionName})
	case "object-store":
		client, err = openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{Region: cred.Clouds.Openstack.RegionName})
	case "compute":
		client, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{Region: cred.Clouds.Openstack.RegionName})
	default:
		o.Expect(fmt.Errorf("unsupported OpenStack service type: %s", serviceType)).NotTo(o.HaveOccurred())
	}
	o.Expect(err).NotTo(o.HaveOccurred())
	return client
}

// getAuthenticatedUserID gets current user ID from auth response
func getAuthenticatedUserID(providerClient *gophercloud.ProviderClient) (string, error) {
	res := providerClient.GetAuthResult()
	if res == nil {
		return "", fmt.Errorf("no AuthResult available")
	}
	switch r := res.(type) {
	case tokens3.CreateResult:
		u, err := r.ExtractUser()
		if err != nil {
			return "", err
		}
		return u.ID, nil
	default:
		return "", fmt.Errorf("got unexpected AuthResult type %t", r)
	}
}

// getOpenStackUserIDAndDomainID returns the user ID and domain ID
func getOpenStackUserIDAndDomainID(cred *openstackCredentials) (string, string) {
	client := newOpenStackClient(cred, "identity")
	userID, err := getAuthenticatedUserID(client.ProviderClient)
	o.Expect(err).NotTo(o.HaveOccurred())
	user, err := users.Get(client, userID).Extract()
	o.Expect(err).NotTo(o.HaveOccurred())
	return userID, user.DomainID
}

// createOpenStackContainer creates a storage container in OpenStack
func createOpenStackContainer(client *gophercloud.ServiceClient, name string) error {
	pager := containers.List(client, &containers.ListOpts{Full: true, Prefix: name})
	exist := false

	// Check if the container exists
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		containerNames, err := containers.ExtractNames(page)
		o.Expect(err).NotTo(o.HaveOccurred())
		for _, n := range containerNames {
			if n == name {
				exist = true
				break
			}
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	if exist {
		err = emptyOpenStackContainer(client, name)
		o.Expect(err).NotTo(o.HaveOccurred())
	}

	// Create the container
	res := containers.Create(client, name, containers.CreateOpts{})
	_, err = res.Extract()
	return err
}

// deleteOpenStackContainer deletes the storage container from OpenStack
func deleteOpenStackContainer(client *gophercloud.ServiceClient, name string) error {
	err := emptyOpenStackContainer(client, name)
	if err != nil {
		return err
	}

	response := containers.Delete(client, name)
	_, err = response.Extract()
	if err != nil {
		return fmt.Errorf("error deleting container %s: %v", name, err)
	}
	e2e.Logf("Container %s is deleted", name)
	return nil
}

// emptyOpenStackContainer clears all the objects in storage container
func emptyOpenStackContainer(client *gophercloud.ServiceClient, name string) error {
	pager := objects.List(client, name, &objects.ListOpts{Full: true})
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		objectNames, err := objects.ExtractNames(page)
		if err != nil {
			return false, fmt.Errorf("error getting object names: %v", err)
		}
		for _, obj := range objectNames {
			result := objects.Delete(client, name, obj, objects.DeleteOpts{})
			_, err := result.Extract()
			if err != nil {
				return false, fmt.Errorf("hit error when deleting object %s: %v", obj, err)
			}
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("error deleting objects in container %s: %v", name, err)
	}
	e2e.Logf("All objects in container %s are removed", name)
	return nil
}
