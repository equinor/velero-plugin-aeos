package plugin

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/sirupsen/logrus"
	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"
)

type FileObjectStore struct {
	log        logrus.FieldLogger
	credential *azblob.SharedKeyCredential
	pipeline   *pipeline.Pipeline
	service    *azblob.ServiceURL
	cpk        *azblob.ClientProvidedKeyOptions
}

// NewFileObjectStore instantiates a FileObjectStore.
func NewFileObjectStore(log logrus.FieldLogger) *FileObjectStore {
	return &FileObjectStore{log: log}
}

// Init initializes the plugin. After v0.10.0, this can be called multiple times.
func (f *FileObjectStore) Init(config map[string]string) error {
	f.log.Infof("Init")

	if err := veleroplugin.ValidateObjectStoreConfigKeys(config,
		storageAccountConfigKey,
		credentialsFileConfigKey,
	); err != nil {
		return err
	}

	// make best effort to find a valid secret file either from the config or the environment.
	// if one is found, load it. if not, assume secret vars are loaded into the environment already.
	secretsFilePath, ok := tryResolveSecretsFile(config[credentialsFileConfigKey])
	if ok {
		if err := loadSecretsFile(secretsFilePath); err != nil {
			return err
		}
	}

	secrets, err := getRequiredSecrets(storageAccountKeyEnvVar, encryptionKeyEnvVar, encryptionHashEnvVar)
	if err != nil {
		return err
	}

	key := secrets[encryptionKeyEnvVar]
	hash := secrets[encryptionHashEnvVar]
	scope := os.Getenv(encryptionScopeEnvVar)
	cpk := azblob.NewClientProvidedKeyOptions(&key, &hash, &scope)

	cred, err := azblob.NewSharedKeyCredential(config[storageAccountConfigKey], secrets[storageAccountKeyEnvVar])
	if err != nil {
		return err
	}

	blobDN := parseBlobDomainName(os.Getenv(blobDomainNameEnvVar))
	if blobDN == "" {
		blobDN = defaultBlobDomain
	}

	u, _ := url.Parse(fmt.Sprintf("https://%s.%s", config[storageAccountConfigKey], blobDN))
	if err != nil {
		return err
	}

	pipeline := azblob.NewPipeline(cred, azblob.PipelineOptions{})
	service := azblob.NewServiceURL(*u, pipeline)

	f.credential = cred
	f.pipeline = &pipeline
	f.service = &service
	f.cpk = &cpk

	return nil
}

func (f *FileObjectStore) PutObject(bucket string, key string, body io.Reader) error {
	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
	})
	log.Infof("PutObject")

	container := f.service.NewContainerURL(bucket)
	blobURL := container.NewBlockBlobURL(key)
	r, err := azblob.UploadStreamToBlockBlob(context.Background(), body, blobURL, azblob.UploadStreamToBlockBlobOptions{ClientProvidedKeyOptions: *f.cpk})

	_ = r

	if err != nil {
		return err
	}
	return nil
}

func (f *FileObjectStore) ObjectExists(bucket, key string) (bool, error) {
	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
	})
	log.Infof("ObjectExists")
	ctx := context.Background()
	container := f.service.NewContainerURL(bucket)
	blob := container.NewBlobURL(key)
	_, err := blob.GetProperties(ctx, azblob.BlobAccessConditions{}, *f.cpk)

	if err == nil {
		return true, err
	}

	if storageErr, ok := err.(azblob.StorageError); ok {
		if storageErr.Response().StatusCode == 404 {
			return false, nil
		}
	}

	return false, err
}

func (f *FileObjectStore) GetObject(bucket, key string) (io.ReadCloser, error) {
	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
	})
	log.Infof("GetObject")

	container := f.service.NewContainerURL(bucket)
	blobURL := container.NewBlockBlobURL(key)
	response, err := blobURL.Download(context.TODO(), 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, *f.cpk)
	if err != nil {
		return nil, err
	}

	return response.Body(azblob.RetryReaderOptions{}), nil
}

func (f *FileObjectStore) ListCommonPrefixes(bucket, prefix, delimiter string) ([]string, error) {
	log := f.log.WithFields(logrus.Fields{
		"bucket":    bucket,
		"delimiter": delimiter,
		"prefix":    prefix,
	})
	log.Infof("ListCommonPrefixes")

	var prefixes []string
	container := f.service.NewContainerURL(bucket)
	marker := azblob.Marker{}

	for marker.NotDone() {
		listBlob, err := container.ListBlobsHierarchySegment(context.Background(), marker, delimiter, azblob.ListBlobsSegmentOptions{Prefix: prefix})

		if err != nil {
			return nil, err
		}

		for _, blobInfo := range listBlob.Segment.BlobPrefixes {
			prefixes = append(prefixes, blobInfo.Name)
		}

		marker = listBlob.NextMarker
	}

	return prefixes, nil // This function is not implemented.
}

func (f *FileObjectStore) ListObjects(bucket, prefix string) ([]string, error) {
	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"prefix": prefix,
	})
	log.Infof("ListObjects")

	var objects []string
	container := f.service.NewContainerURL(bucket)
	marker := azblob.Marker{}

	for marker.NotDone() {
		listBlob, err := container.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{Prefix: prefix})

		if err != nil {
			return nil, err
		}

		for _, blobInfo := range listBlob.Segment.BlobItems {
			objects = append(objects, blobInfo.Name)
		}

		marker = listBlob.NextMarker
	}
	return objects, nil
}

func (f *FileObjectStore) DeleteObject(bucket, key string) error {
	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
	})
	log.Infof("DeleteObject")

	container := f.service.NewContainerURL(bucket)
	blobURL := container.NewBlockBlobURL(key)
	_, err := blobURL.Delete(context.Background(), azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	if err != nil {
		return err
	}
	return nil
}

func (f *FileObjectStore) CreateSignedURL(bucket, key string, ttl time.Duration) (string, error) {
	log := f.log.WithFields(logrus.Fields{
		"bucket": bucket,
		"key":    key,
	})
	log.Infof("CreateSignedURL")

	credential, err := azblob.NewSharedKeyCredential(bucket, key)
	if err != nil {
		return "", err
	}

	sasQueryParams, err := azblob.AccountSASSignatureValues{
		Protocol:      azblob.SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(ttl),
		Permissions:   azblob.AccountSASPermissions{Read: true, List: true}.String(),
		Services:      azblob.AccountSASServices{Blob: true}.String(),
		ResourceTypes: azblob.AccountSASResourceTypes{Container: true, Object: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		return "", err
	}

	qp := sasQueryParams.Encode()
	return fmt.Sprintf("https://%s.%s?%s", bucket, defaultBlobDomain, qp), nil
}
