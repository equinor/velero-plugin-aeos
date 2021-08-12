package plugin

const (
	resourceGroupConfigKey   = "resourceGroup"
	storageAccountConfigKey  = "storageAccount"
	subscriptionIDConfigKey  = "subscriptionId"
	blockSizeConfigKey       = "blockSizeInBytes"
	credentialsFileConfigKey = "credentialsFile"
	storageAccountKeyEnvVar  = "AZURE_STORAGE_ACCOUNT_ACCESS_KEY"
	blobDomainNameEnvVar     = "AZURE_BLOB_DOMAIN_NAME"
	encryptionKeyEnvVar      = "AZURE_STORAGE_ACCOUNT_ENCRYPTION_KEY"
	encryptionHashEnvVar     = "AZURE_STORAGE_ACCOUNT_ENCRYPTION_HASH"
	encryptionScopeEnvVar    = "AZURE_STORAGE_ACCOUNT_ENCRYPTION_SCOPE"
	secretsFileEnvVar        = "AZURE_CREDENTIALS_FILE"
	defaultBlobDomain        = "blob.core.windows.net"
)
