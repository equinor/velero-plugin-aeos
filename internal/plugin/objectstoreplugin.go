/*
Copyright 2017, 2019 the Velero contributors.

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

package plugin

import (
	"context"
	"errors"
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
	f.log.Infof("FileObjectStore.Init called")
	if err := veleroplugin.ValidateObjectStoreConfigKeys(config,
		storageAccountConfigKey,
		credentialsFileConfigKey,
		subscriptionIDConfigKey,
		resourceGroupConfigKey,
	); err != nil {
		return err
	}

	if err := loadSecretsFile(config[credentialsFileConfigKey]); err != nil {
		return err
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

	return make([]string, 0), nil // This function is not implemented.
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
		marker = listBlob.NextMarker

		for _, blobInfo := range listBlob.Segment.BlobItems {
			objects = append(objects, blobInfo.Name)
		}
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
	sasQueryParams, err := azblob.BlobSASSignatureValues{
		Protocol:      azblob.SASProtocolHTTPS,
		ExpiryTime:    time.Now().UTC().Add(ttl),
		ContainerName: bucket,
		BlobName:      key,
		Permissions:   azblob.BlobSASPermissions{Add: false, Read: true, Write: false}.String()}.NewSASQueryParameters(f.credential)
	if err != nil {
		log.Fatal(err)
	}

	qp := sasQueryParams.Encode()
	f.service.URL()
	SasUri := fmt.Sprintf("%s/%s/%s?%s",
		f.service.URL().RawPath, bucket, key, qp)

	return SasUri, errors.New("not implemented")
}
