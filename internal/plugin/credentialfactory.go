package plugin

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/adal"
)

func credentialFactory(config PluginConfigMap) (azblob.Credential, error) {
	// try key based auth first
	if _, exists := os.LookupEnv(storageAccountKeyEnvVar); exists {
		return buildSharedKeyCredentialFromEnv(config)
	}

	return getOAuthToken(config)
}

func buildSharedKeyCredentialFromEnv(config PluginConfigMap) (*azblob.SharedKeyCredential, error) {
	secrets, err := getSecrets(false, storageAccountKeyEnvVar, encryptionKeyEnvVar, encryptionHashEnvVar)
	if err != nil {
		return nil, err
	}

	cred, err := azblob.NewSharedKeyCredential(config[storageAccountConfigKey], secrets[storageAccountKeyEnvVar])
	if err != nil {
		return nil, err
	}

	return cred, nil
}

func fetchMSIToken(config PluginConfigMap) (*adal.ServicePrincipalToken, error) {
	secrets, err := getSecrets(false, subscriptionIDEnvVar, storageAccountIdEnvVar, clientIDEnvVar)
	_ = secrets
	_ = err

	msiEndpoint, err := adal.GetMSIEndpoint()
	if err != nil {
		return nil, err
	}

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

func getOAuthToken(config PluginConfigMap) (azblob.Credential, error) {
	spt, err := fetchMSIToken(config)
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

	return tc, nil
}
