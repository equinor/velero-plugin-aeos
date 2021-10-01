package plugin

const (
	resourceGroupConfigKey   = "resourceGroup"
	storageAccountConfigKey  = "storageAccount"
	subscriptionIDConfigKey  = "subscriptionId"
	blockSizeConfigKey       = "blockSizeInBytes"
	credentialsFileConfigKey = "credentialsFile"
	subscriptionIDEnvVar     = "AZURE_SUBSCRIPTION_ID"
	storageAccountKeyEnvVar  = "AZURE_STORAGE_ACCOUNT_ACCESS_KEY"
	blobDomainNameEnvVar     = "AZURE_BLOB_DOMAIN_NAME"
	encryptionKeyEnvVar      = "AZURE_STORAGE_ACCOUNT_ENCRYPTION_KEY"
	encryptionHashEnvVar     = "AZURE_STORAGE_ACCOUNT_ENCRYPTION_HASH"
	encryptionScopeEnvVar    = "AZURE_STORAGE_ACCOUNT_ENCRYPTION_SCOPE"
	secretsFileEnvVar        = "AZURE_CREDENTIALS_FILE"
	clientIDEnvVar           = "AZURE_CLIENT_ID"
	defaultBlobDomain        = "blob.core.windows.net"
	publicResourceManager    = "https://management.azure.com/"
)
