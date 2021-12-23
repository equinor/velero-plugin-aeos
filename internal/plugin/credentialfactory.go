package plugin

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/sirupsen/logrus"
)

type credentialFactory struct {
	config PluginConfigMap
	log    logrus.FieldLogger
}

func newCredentialFactory(config PluginConfigMap, log logrus.FieldLogger) *credentialFactory {
	return &credentialFactory{config: config, log: log}
}

func (f *credentialFactory) Build() (azblob.Credential, error) {
	// try key based auth first
	if _, exists := os.LookupEnv(storageAccountKeyEnvVar); exists {
		f.log.Debugln("Building Shared Key Credential...")
		return f.buildSharedKeyCredentialFromEnv()
	}

	f.log.Debugln("Building MSI Credential...")
	return f.getOAuthToken()
}

func (f *credentialFactory) buildSharedKeyCredentialFromEnv() (*azblob.SharedKeyCredential, error) {
	secrets, err := getSecrets(false, storageAccountKeyEnvVar, encryptionKeyEnvVar, encryptionHashEnvVar)
	if err != nil {
		return nil, err
	}

	cred, err := azblob.NewSharedKeyCredential(f.config[storageAccountConfigKey], secrets[storageAccountKeyEnvVar])
	if err != nil {
		return nil, err
	}

	return cred, nil
}

func (f *credentialFactory) fetchMSIToken() (*adal.ServicePrincipalToken, error) {
	secrets, _ := getSecrets(false, f.config[subscriptionIDConfigKey], clientIDEnvVar)

	msiEndpoint, err := adal.GetMSIEndpoint()
	if err != nil {
		return nil, err
	}
	f.log.Debugf("MSI Endpoint: %s\n", msiEndpoint)

	var spToken *adal.ServicePrincipalToken
	if secrets[clientIDEnvVar] == "" {
		spToken, err = adal.NewServicePrincipalTokenFromMSI(msiEndpoint, publicResourceManager)
		if err != nil {
			return nil, fmt.Errorf("failed to get oauth token from MSI: %v", err)
		}
	} else {
		spToken, err = adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(msiEndpoint, publicResourceManager, secrets[clientIDEnvVar])
		if err != nil {
			return nil, fmt.Errorf("failed to get oauth token from MSI for user assigned identity: %v", err)
		}
	}

	return spToken, nil
}

func (f *credentialFactory) getOAuthToken() (azblob.Credential, error) {
	spt, err := f.fetchMSIToken()
	if err != nil {
		log.Fatal(err)
	}

	// Refresh obtains a fresh token
	err = spt.Refresh()
	if err != nil {
		log.Fatal(err)
	}

	tc := azblob.NewTokenCredential(spt.Token().AccessToken, func(tc azblob.TokenCredential) time.Duration {
		err := spt.Refresh()
		if err != nil {
			// something went wrong, prevent the refresher from being triggered again
			return 0
		}

		// set the new token value
		tc.SetToken(spt.Token().AccessToken)

		// get the next token slightly before the current one expires
		return time.Until(spt.Token().Expires()) - 10*time.Second
	})

	f.log.Debugf("AccessToken: %s, ExpiresIn: %s\n", spt.Token().AccessToken, spt.Token().ExpiresIn)
	return tc, nil
}
