package plugin

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

const (
	testConfigEnvVar = "VELERO_PLUGIN_AEOS_TEST_CONFIG"
)

var testConfig map[string]string = make(map[string]string)

// test config file schema
// {
// 	"storageAccount": "",
//  "credentialsFile": ""
// 	"containerName": "",
// 	"testBlobName": "",
// }
// credentials file schema
// AZURE_STORAGE_ACCOUNT_ACCESS_KEY=value
// AZURE_BLOB_DOMAIN_NAME=value
// AZURE_STORAGE_ACCOUNT_ENCRYPTION_KEY=value
// AZURE_STORAGE_ACCOUNT_ENCRYPTION_HASH=value
// AZURE_STORAGE_ACCOUNT_ENCRYPTION_SCOPE=value

func loadtestConfigfile() (map[string]string, error) {
	var allowedKeys []string = []string{
		"storageAccount",
		"credentialsFile",
	}

	path := os.Getenv(testConfigEnvVar)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return nil, err
	}

	config := make(map[string]string)
	for k, v := range data {
		config[k] = v.(string)
	}

	for k, v := range config {
		testConfig[k] = v
	}

	for key := range config {
		for _, allowedKey := range allowedKeys {
			if key == allowedKey {
				goto SKIP
			}
		}
		delete(config, key)
	SKIP:
	}

	println(testConfig)
	return config, nil
}

func TestPreviewInit(t *testing.T) {
	config, err := loadtestConfigfile()
	if err != nil {
		t.Error(err)
	}

	objectStore := NewFileObjectStore(logrus.New())

	err = objectStore.Init(config)
	if err != nil {
		t.Error(err)
	}
}

func TestPreviewPutObject(t *testing.T) {
	config, err := loadtestConfigfile()
	if err != nil {
		t.Error(err)
	}

	objectStore := NewFileObjectStore(logrus.New())

	err = objectStore.Init(config)
	if err != nil {
		t.Error(err)
	}

	fd, err := os.Open(testConfig["testFilePath"])
	if err != nil {
		t.Error(err)
	}
	defer fd.Close()

	err = objectStore.PutObject(testConfig["containerName"], testConfig["testBlobName"], fd)
	if err != nil {
		t.Error(err)
	}
}

func TestPreviewListObjects(t *testing.T) {
	config, err := loadtestConfigfile()
	if err != nil {
		t.Error(err)
	}

	objectStore := NewFileObjectStore(logrus.New())

	err = objectStore.Init(config)
	if err != nil {
		t.Error(err)
	}

	objects, err := objectStore.ListObjects(testConfig["containerName"], "")
	if err != nil {
		t.Error(err)
	}
	if len(objects) == 0 {
		t.Error("No objects found")
	}
}

func TestPreviewObjectExists(t *testing.T) {
	config, err := loadtestConfigfile()
	if err != nil {
		t.Error(err)
	}

	objectStore := NewFileObjectStore(logrus.New())

	err = objectStore.Init(config)
	if err != nil {
		t.Error(err)
	}

	exists, err := objectStore.ObjectExists(testConfig["containerName"], testConfig["testBlobName"])
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Fail()
	}
}

func TestPreviewGetObject(t *testing.T) {
	config, err := loadtestConfigfile()
	if err != nil {
		t.Error(err)
	}

	objectStore := NewFileObjectStore(logrus.New())

	err = objectStore.Init(config)
	if err != nil {
		t.Error(err)
	}
	rc, err := objectStore.GetObject(testConfig["containerName"], testConfig["testBlobName"])
	if err != nil {
		t.Error(err)
	}

	fd, err := os.OpenFile(testConfig["testFilePath"]+"-output", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		t.Error(err)
	}

	bw, err := io.Copy(fd, rc)
	if err != nil {
		t.Error(err)
	}
	t.Logf("bytes written: %d", bw)
}
func TestPreviewDeleteObject(t *testing.T) {
	config, err := loadtestConfigfile()
	if err != nil {
		t.Error(err)
	}

	objectStore := NewFileObjectStore(logrus.New())

	err = objectStore.Init(config)
	if err != nil {
		t.Error(err)
	}
	err = objectStore.DeleteObject(testConfig["containerName"], testConfig["testBlobName"])
	if err != nil {
		t.Error(err)
	}
}
