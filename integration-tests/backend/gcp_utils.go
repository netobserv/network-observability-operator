package e2etests

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

func getGcpProjectID() (string, error) {
	obj, err := getDynamicResource("infrastructures", "cluster", "")
	if err != nil {
		return "", err
	}
	projectID, found := getNestedField(obj.Object, ".status.platformStatus.gcp.projectID")
	if !found {
		return "", fmt.Errorf("gcp projectID not found in infrastructure status")
	}
	return projectID, nil
}

// listGCSBuckets lists all buckets in the project
func listGCSBuckets(client storage.Client, projectID string) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buckets []string
	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, battrs.Name)
	}
	return buckets, nil
}

// emptyGCSBucket removes all objects in a bucket
func emptyGCSBucket(client storage.Client, bucketName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bucket := client.Bucket(bucketName)
	it := bucket.Objects(ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err != nil && err != iterator.Done {
			return fmt.Errorf("can't get objects in bucket %s: %v", bucketName, err)
		}
		if err == iterator.Done {
			break
		}
		if err := bucket.Object(objAttrs.Name).Delete(ctx); err != nil {
			return fmt.Errorf("Object(%q).Delete: %v", objAttrs.Name, err)
		}
	}
	e2e.Logf("deleted all object items in the bucket %s.", bucketName)
	return nil
}

// createGCSBucket creates a GCS bucket or empties it if it exists
func createGCSBucket(projectID, bucketName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the GCS client, credentials from GOOGLE_APPLICATION_CREDENTIALS env var
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Check if the bucket exists
	exist := false
	buckets, err := listGCSBuckets(*client, projectID)
	if err != nil {
		return err
	}
	for _, bu := range buckets {
		if bu == bucketName {
			exist = true
			break
		}
	}

	if exist {
		// Bucket exists, empty it
		return emptyGCSBucket(*client, bucketName)
	}

	// Bucket doesn't exist, create it
	bucket := client.Bucket(bucketName)
	if err := bucket.Create(ctx, projectID, &storage.BucketAttrs{}); err != nil {
		return fmt.Errorf("Bucket(%q).Create: %v", bucketName, err)
	}
	e2e.Logf("GCS Bucket %v is created", bucketName)
	return nil
}

// deleteGCSBucket deletes a GCS bucket
func deleteGCSBucket(bucketName string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Remove all objects first
	err = emptyGCSBucket(*client, bucketName)
	if err != nil {
		return err
	}

	// Delete the bucket
	bucket := client.Bucket(bucketName)
	if err := bucket.Delete(ctx); err != nil {
		return fmt.Errorf("Bucket(%q).Delete: %v", bucketName, err)
	}
	e2e.Logf("GCS Bucket %v is deleted", bucketName)
	return nil
}
