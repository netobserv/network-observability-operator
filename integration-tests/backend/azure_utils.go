package e2etests

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	azRuntime "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	azTo "github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	azblobv2 "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/google/uuid"
	o "github.com/onsi/gomega"
	"github.com/tidwall/gjson"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

func getAzureStorageAccountFromCluster() (string, string, error) {
	deploy, err := k8sClient.AppsV1().Deployments("openshift-image-registry").Get(context.Background(), "image-registry", metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}
	var accountName string
	for _, c := range deploy.Spec.Template.Spec.Containers {
		for _, env := range c.Env {
			if env.Name == "REGISTRY_STORAGE_AZURE_ACCOUNTNAME" {
				accountName = env.Value
				break
			}
		}
	}
	if accountName == "" {
		return "", "", fmt.Errorf("REGISTRY_STORAGE_AZURE_ACCOUNTNAME not found on image-registry deployment")
	}

	secret, err := k8sClient.CoreV1().Secrets("openshift-image-registry").Get(context.Background(), "image-registry-private-configuration", metav1.GetOptions{})
	if err != nil {
		return accountName, "", err
	}

	return accountName, string(secret.Data["REGISTRY_STORAGE_AZURE_ACCOUNTKEY"]), nil
}

// To read Azure subscription json file from local disk.
// Also injects ENV vars needed to perform certain operations on Managed Identities.
func readAzureCredentials() bool {
	var azureCredFile string
	envDir, present := os.LookupEnv("CLUSTER_PROFILE_DIR")
	if present {
		azureCredFile = filepath.Join(envDir, "osServicePrincipal.json")
	} else {
		authFileLocation, present := os.LookupEnv("AZURE_AUTH_LOCATION")
		if present {
			azureCredFile = authFileLocation
		}
	}
	if len(azureCredFile) > 0 {
		fileContent, err := os.ReadFile(azureCredFile)
		o.Expect(err).NotTo(o.HaveOccurred())

		subscriptionID := gjson.Get(string(fileContent), `azure_subscription_id`).String()
		if subscriptionID == "" {
			subscriptionID = gjson.Get(string(fileContent), `subscriptionId`).String()
		}
		os.Setenv("AZURE_SUBSCRIPTION_ID", subscriptionID)

		tenantID := gjson.Get(string(fileContent), `azure_tenant_id`).String()
		if tenantID == "" {
			tenantID = gjson.Get(string(fileContent), `tenantId`).String()
		}
		os.Setenv("AZURE_TENANT_ID", tenantID)

		clientID := gjson.Get(string(fileContent), `azure_client_id`).String()
		if clientID == "" {
			clientID = gjson.Get(string(fileContent), `clientId`).String()
		}
		os.Setenv("AZURE_CLIENT_ID", clientID)

		clientSecret := gjson.Get(string(fileContent), `azure_client_secret`).String()
		if clientSecret == "" {
			clientSecret = gjson.Get(string(fileContent), `clientSecret`).String()
		}
		os.Setenv("AZURE_CLIENT_SECRET", clientSecret)
		return true
	}
	return false
}

// Creates a new default Azure credential
func createNewDefaultAzureCredential() *azidentity.DefaultAzureCredential {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	o.Expect(err).NotTo(o.HaveOccurred(), "Failed to obtain a credential")
	return cred
}

// Function to create a managed identity on Azure
func createManagedIdentityOnAzure(defaultAzureCred *azidentity.DefaultAzureCredential, azureSubscriptionID, lokiStackName, resourceGroup, region string) (string, string) {
	// Create the MSI client
	client, err := armmsi.NewUserAssignedIdentitiesClient(azureSubscriptionID, defaultAzureCred, nil)
	o.Expect(err).NotTo(o.HaveOccurred(), "Failed to create MSI client")

	// Configure the managed identity
	identity := armmsi.Identity{
		Location: &region,
	}

	// Create the identity
	result, err := client.CreateOrUpdate(context.Background(), resourceGroup, lokiStackName, identity, nil)
	o.Expect(err).NotTo(o.HaveOccurred(), "Failed to create or update the identity")
	return *result.Properties.ClientID, *result.Properties.PrincipalID
}

// Function to create Federated Credentials on Azure
func createFederatedCredentialforLoki(defaultAzureCred *azidentity.DefaultAzureCredential, azureSubscriptionID, managedIdentityName, lokiServiceAccount, lokiStackNS, federatedCredentialName, serviceAccountIssuer, resourceGroup string) {
	subjectName := "system:serviceaccount:" + lokiStackNS + ":" + lokiServiceAccount

	// Create the Federated Identity Credentials client
	client, err := armmsi.NewFederatedIdentityCredentialsClient(azureSubscriptionID, defaultAzureCred, nil)
	o.Expect(err).NotTo(o.HaveOccurred(), "Failed to create federated identity credentials client")

	// Create or update the federated identity credential
	result, err := client.CreateOrUpdate(
		context.Background(),
		resourceGroup,
		managedIdentityName,
		federatedCredentialName,
		armmsi.FederatedIdentityCredential{
			Properties: &armmsi.FederatedIdentityCredentialProperties{
				Issuer:    &serviceAccountIssuer,
				Subject:   &subjectName,
				Audiences: []*string{azTo.Ptr("api://AzureADTokenExchange")},
			},
		},
		nil,
	)
	o.Expect(err).NotTo(o.HaveOccurred(), "Failed to create or update the federated credential: "+federatedCredentialName)
	e2e.Logf("Federated credential created/updated successfully: %s\n", *result.Name)
}

// Assigns role to a Azure Managed Identity on subscription level scope
func createRoleAssignmentForManagedIdentity(defaultAzureCred *azidentity.DefaultAzureCredential, azureSubscriptionID, identityPrincipalID string) {
	clientFactory, err := armauthorization.NewClientFactory(azureSubscriptionID, defaultAzureCred, nil)
	o.Expect(err).NotTo(o.HaveOccurred(), "Failed to create instance of ClientFactory")

	scope := "/subscriptions/" + azureSubscriptionID
	// Below is standard role definition ID for Storage Blob Data Contributor built-in role
	roleDefinitionID := scope + "/providers/Microsoft.Authorization/roleDefinitions/ba92f5b4-2d11-453d-a403-e96b0029c9fe"

	// Create or update a role assignment by scope and name
	_, err = clientFactory.NewRoleAssignmentsClient().Create(context.Background(), scope, uuid.NewString(), armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      azTo.Ptr(identityPrincipalID),
			PrincipalType:    azTo.Ptr(armauthorization.PrincipalTypeServicePrincipal),
			RoleDefinitionID: azTo.Ptr(roleDefinitionID),
		},
	}, nil)
	o.Expect(err).NotTo(o.HaveOccurred(), "Role Assignment operation failure....")
}

// Creates Azure storage account
func createStorageAccountOnAzure(defaultAzureCred *azidentity.DefaultAzureCredential, azureSubscriptionID, resourceGroup, region string) string {
	storageAccountName := "aosqelogging" + getRandomString()
	// Create the storage account
	storageClient, err := armstorage.NewAccountsClient(azureSubscriptionID, defaultAzureCred, nil)
	o.Expect(err).NotTo(o.HaveOccurred())
	result, err := storageClient.BeginCreate(context.Background(), resourceGroup, storageAccountName, armstorage.AccountCreateParameters{
		Location: azTo.Ptr(region),
		SKU: &armstorage.SKU{
			Name: azTo.Ptr(armstorage.SKUNameStandardLRS),
		},
		Kind: azTo.Ptr(armstorage.KindStorageV2),
	}, nil)
	o.Expect(err).NotTo(o.HaveOccurred())

	// Poll until the Storage account is ready
	_, err = result.PollUntilDone(context.Background(), &azRuntime.PollUntilDoneOptions{
		Frequency: 10 * time.Second,
	})
	o.Expect(err).NotTo(o.HaveOccurred(), "Storage account is not ready...")
	os.Setenv("LOKI_OBJECT_STORAGE_STORAGE_ACCOUNT", storageAccountName)
	return storageAccountName
}

func getAzureResourceGroupFromCluster() (string, error) {
	obj, err := getDynamicResource("infrastructures", "cluster", "")
	if err != nil {
		return "", err
	}

	resourceGroup, found := getNestedField(obj.Object, ".status.platformStatus.azure.resourceGroupName")
	if !found || resourceGroup == "" {
		return "", fmt.Errorf("failed to get resource group name: empty value")
	}

	return resourceGroup, nil
}

// Returns the Azure environment and storage account URI suffixes
func getStorageAccountURISuffixAndEnvForAzure() (string, string) {
	obj, err := getDynamicResource("infrastructures", "cluster", "")
	if err != nil {
		return "AzureGlobal", ".blob.core.windows.net"
	}

	cloudName, _ := getNestedField(obj.Object, ".status.platformStatus.azure.cloudName")

	storageAccountURISuffix := ".blob.core.windows.net"
	environment := "AzureGlobal"
	// Currently we don't have template support for STS/WIF on Azure Government
	// The below code should be ok to run when support is added for WIF
	if strings.ToLower(cloudName) == "azureusgovernmentcloud" {
		storageAccountURISuffix = ".blob.core.usgovcloudapi.net"
		environment = "AzureUSGovernment"
	}
	if strings.ToLower(cloudName) == "azurechinacloud" {
		storageAccountURISuffix = ".blob.core.chinacloudapi.cn"
		environment = "AzureChinaCloud"
	}
	if strings.ToLower(cloudName) == "azuregermancloud" {
		environment = "AzureGermanCloud"
		storageAccountURISuffix = ".blob.core.cloudapi.de"
	}
	return environment, storageAccountURISuffix
}

// Creates a blob container under the provided storageAccount
func createBlobContaineronAzure(defaultAzureCred *azidentity.DefaultAzureCredential, storageAccountName, storageAccountURISuffix, containerName string) {
	blobServiceClient, err := azblobv2.NewClient(fmt.Sprintf("https://%s%s", storageAccountName, storageAccountURISuffix), defaultAzureCred, nil)
	o.Expect(err).NotTo(o.HaveOccurred())
	_, err = blobServiceClient.CreateContainer(context.Background(), containerName, nil)
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("%s container created successfully: ", containerName)
}

// Creates Loki object storage secret required on Azure STS/WIF clusters
func createLokiObjectStorageSecretForWIF(lokiStackNS, objectStorageSecretName, environment, containerName, storageAccountName string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectStorageSecretName,
			Namespace: lokiStackNS,
		},
		StringData: map[string]string{
			"environment":  environment,
			"container":    containerName,
			"account_name": storageAccountName,
		},
	}
	_, err := k8sClient.CoreV1().Secrets(lokiStackNS).Create(context.Background(), secret, metav1.CreateOptions{})
	return err
}

// Deletes a storage account in Microsoft Azure
func deleteAzureStorageAccount(defaultAzureCred *azidentity.DefaultAzureCredential, azureSubscriptionID, resourceGroupName, storageAccountName string) {
	clientFactory, err := armstorage.NewClientFactory(azureSubscriptionID, defaultAzureCred, nil)
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to create instance of ClientFactory for storage account deletion")

	_, err = clientFactory.NewAccountsClient().Delete(context.Background(), resourceGroupName, storageAccountName, nil)
	if err != nil {
		e2e.Logf("Error while deleting storage account. %v", err.Error())
	} else {
		e2e.Logf("storage account deleted successfully..")
	}
}

// Deletes the Azure Managed identity
func deleteManagedIdentityOnAzure(defaultAzureCred *azidentity.DefaultAzureCredential, azureSubscriptionID, resourceGroupName, identityName string) {
	client, err := armmsi.NewUserAssignedIdentitiesClient(azureSubscriptionID, defaultAzureCred, nil)
	o.Expect(err).NotTo(o.HaveOccurred(), "failed to create MSI client for identity deletion")

	_, err = client.Delete(context.Background(), resourceGroupName, identityName, nil)
	if err != nil {
		e2e.Logf("Error deleting identity. %v", err.Error())
	} else {
		e2e.Logf("managed identity deleted successfully...")
	}
}

// getAzureCloudStorageURISuffix returns the blob storage URI suffix based on cloud type
func getAzureCloudStorageURISuffix() string {
	obj, err := getDynamicResource("infrastructures", "cluster", "")
	if err != nil {
		e2e.Logf("error getting infrastructure cluster: %v", err)
		return ".blob.core.windows.net"
	}

	cloudName, _ := getNestedField(obj.Object, ".status.platformStatus.azure.cloudName")
	if strings.ToLower(cloudName) == "azureusgovernmentcloud" {
		return ".blob.core.usgovcloudapi.net"
	}
	// Add other clouds as needed:
	// "azurechinacloud" -> ".blob.core.chinacloudapi.cn"
	// "azuregermancloud" -> ".blob.core.cloudapi.de"

	return ".blob.core.windows.net" // AzurePublicCloud default
}

// patches CLIENT_ID, SUBSCRIPTION_ID, TENANT_ID AND REGION into Loki subscription on Azure WIF clusters
func patchLokiConfigIntoLokiSubscription(azureSubscriptionID, identityClientID, region string) {
	patchData := fmt.Sprintf(`{
		"spec": {
			"config": {
				"env": [
					{
						"name": "CLIENTID",
						"value": "%s"
					},
					{
						"name": "TENANTID",
						"value": "%s"
					},
					{
						"name": "SUBSCRIPTIONID",
						"value": "%s"
					},
					{
						"name": "REGION",
						"value": "%s"
					}
				]
			}
		}
	}`, identityClientID, os.Getenv("AZURE_TENANT_ID"), azureSubscriptionID, region)

	err := patchDynamicResource("subscription", "loki-operator", loNS, types.MergePatchType, []byte(patchData))
	o.Expect(err).NotTo(o.HaveOccurred(), "Patching Loki Operator failed...")

	WaitForPodsReadyWithLabel("openshift-operators-redhat", "name=loki-operator-controller-manager")
}

// Performs creation of Managed Identity, Associated Federated credentials, Role assignment to the managed identity and object storage creation on Azure
func performManagedIdentityAndSecretSetupForAzureWIF(lokistackName, lokiStackNS, azureContainerName, lokiStackStorageSecretName string) {
	region, err := getAzureClusterRegion()
	o.Expect(err).NotTo(o.HaveOccurred())
	serviceAccountIssuer, err := getOIDC()
	o.Expect(err).NotTo(o.HaveOccurred())
	resourceGroup, err := getResourceGroupOnAzure()
	o.Expect(err).NotTo(o.HaveOccurred())

	azureSubscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	cred := createNewDefaultAzureCredential()

	identityClientID, identityPrincipalID := createManagedIdentityOnAzure(cred, azureSubscriptionID, lokistackName, resourceGroup, region)
	createFederatedCredentialforLoki(cred, azureSubscriptionID, lokistackName, lokistackName, lokiStackNS, "openshift-logging-"+lokistackName, serviceAccountIssuer, resourceGroup)
	createFederatedCredentialforLoki(cred, azureSubscriptionID, lokistackName, lokistackName+"-ruler", lokiStackNS, "openshift-logging-"+lokistackName+"-ruler", serviceAccountIssuer, resourceGroup)
	createRoleAssignmentForManagedIdentity(cred, azureSubscriptionID, identityPrincipalID)
	patchLokiConfigIntoLokiSubscription(azureSubscriptionID, identityClientID, region)
	storageAccountName := createStorageAccountOnAzure(cred, azureSubscriptionID, resourceGroup, region)
	environment, storageAccountURISuffix := getStorageAccountURISuffixAndEnvForAzure()
	createBlobContaineronAzure(cred, storageAccountName, storageAccountURISuffix, azureContainerName)
	err = createLokiObjectStorageSecretForWIF(lokiStackNS, lokiStackStorageSecretName, environment, azureContainerName, storageAccountName)
	o.Expect(err).NotTo(o.HaveOccurred())
}

func getResourceGroupOnAzure() (string, error) {
	obj, err := getDynamicResource("infrastructures", "cluster", "")
	if err != nil {
		return "", err
	}

	resourceGroup, found := getNestedField(obj.Object, ".status.platformStatus.azure.resourceGroupName")
	if !found || resourceGroup == "" {
		return "", fmt.Errorf("failed to get resource group name: empty value")
	}

	return resourceGroup, nil
}

// Get region/location of cluster running on Azure Cloud
func getAzureClusterRegion() (string, error) {
	nodes, err := k8sClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no nodes found")
	}

	region := nodes.Items[0].Labels["topology.kubernetes.io/region"]
	if region == "" {
		return "", fmt.Errorf("region label not found on node")
	}

	return region, nil
}

// newAzureContainerClient initializes a new azure blob container client
func newAzureContainerClient(accountName, accountKey, azContainerName string) (azblob.ContainerURL, error) {
	storageAccountURISuffix := getAzureCloudStorageURISuffix()
	u, _ := url.Parse(fmt.Sprintf("https://%s%s", accountName, storageAccountURISuffix))
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	serviceURL := azblob.NewServiceURL(*u, p)
	return serviceURL.NewContainerURL(azContainerName), err
}

// createAzureStorageBlobContainer creates azure storage container
func createAzureStorageBlobContainer(accountName, accountKey, containerName string) error {
	container, err := newAzureContainerClient(accountName, accountKey, containerName)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// check if the container exists or not
	// if exists, then remove the blobs in the container, if not, create the container
	_, err = container.GetProperties(ctx, azblob.LeaseAccessConditions{})
	message := fmt.Sprintf("%v", err)
	if strings.Contains(message, "ContainerNotFound") {
		_, err = container.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
		return err
	}
	return emptyAzureBlobContainer(container)
}

// deleteAzureStorageBlobContainer deletes azure storage container with retry logic for transient errors
func deleteAzureStorageBlobContainer(accountName, accountKey, containerName string) error {
	container, err := newAzureContainerClient(accountName, accountKey, containerName)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		// Remove all blobs first
		err = emptyAzureBlobContainer(container)
		if err != nil {
			lastErr = err
			e2e.Logf("attempt %d/3: failed to empty container %s: %v", attempt, containerName, err)
			time.Sleep(time.Duration(attempt) * 5 * time.Second)
			continue
		}

		// Delete the container
		_, err = container.Delete(ctx, azblob.ContainerAccessConditions{})
		if err != nil {
			lastErr = err
			e2e.Logf("attempt %d/3: failed to delete container %s: %v", attempt, containerName, err)
			time.Sleep(time.Duration(attempt) * 5 * time.Second)
			continue
		}
		e2e.Logf("Azure storage container %s is deleted", containerName)
		return nil
	}
	return fmt.Errorf("failed to delete Azure storage container %s after 3 attempts: %v", containerName, lastErr)
}

// emptyAzureBlobContainer removes all the files in azure storage container
func emptyAzureBlobContainer(container azblob.ContainerURL) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var errs []string
	for marker := (azblob.Marker{}); marker.NotDone(); { // The parens around Marker{} are required to avoid compiler error.
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := container.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return fmt.Errorf("error listing blobs in container: %v", err)
		}

		// IMPORTANT: ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			blobURL := container.NewBlockBlobURL(blobInfo.Name)
			_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
			if err != nil {
				e2e.Logf("WARNING: failed to delete blob %q in container: %v", blobInfo.Name, err)
				errs = append(errs, fmt.Sprintf("Blob(%q).Delete: %v", blobInfo.Name, err))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to delete %d blob(s) in container: %s", len(errs), strings.Join(errs, "; "))
	}
	e2e.Logf("deleted all blob items in the container.")
	return nil
}
