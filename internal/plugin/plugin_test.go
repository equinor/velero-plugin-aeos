package plugin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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
//   "storageAccount": "",
//   "credentialsFile": "",
//   "containerName": "",
//   "testBlobName": "",
//   "testFilePath": ""
// }
// credentials file schema
// AZURE_STORAGE_ACCOUNT_ACCESS_KEY=value
// AZURE_STORAGE_ACCOUNT_ENCRYPTION_KEY=value
// AZURE_STORAGE_ACCOUNT_ENCRYPTION_HASH=value

func loadtestConfigfile() (map[string]string, error) {
	var allowedKeys []string = []string{
		"storageAccount",
		"credentialsFile",
	}

	b64Config, ok := os.LookupEnv(testConfigEnvVar)
	if !ok {
		return nil, fmt.Errorf("could not find env var: %s", testConfigEnvVar)
	}

	bytes, err := base64.RawStdEncoding.DecodeString(b64Config)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
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

func TestInit(t *testing.T) {
	config, err := loadtestConfigfile()
	if err != nil {
		t.Fatal(err)
	}

	objectStore := NewFileObjectStore(logrus.New())

	err = objectStore.Init(config)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPutObject(t *testing.T) {
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

func TestListObjects(t *testing.T) {
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

func TestObjectExists(t *testing.T) {
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

func TestGetObject(t *testing.T) {
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

	var outputPath = testConfig["testFilePath"] + "-output"
	fd, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		t.Error(err)
	}

	bw, err := io.Copy(fd, rc)
	if err != nil {
		t.Error(err)
	}

	os.Remove(outputPath) // failing to delete file should not fail the test
	t.Logf("bytes written: %d", bw)
}
func TestDeleteObject(t *testing.T) {
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
