package plugin

const (
	resourceGroupConfigKey           = "resourceGroup"
	storageAccountConfigKey          = "storageAccount"
	storageAccountKeyEnvVarConfigKey = "storageAccountKeyEnvVar"
	subscriptionIDConfigKey          = "subscriptionId"
	blockSizeConfigKey               = "blockSizeInBytes"
	credentialsFileEnvVar            = "AZURE_CREDENTIALS_FILE"
	encryptionKeyEnvVar              = "AZURE_ENCRYPTION_KEY"
	encryptionHashEnvVar             = "AZURE_ENCRYPTION_HASH"

	// see https://docs.microsoft.com/en-us/rest/api/storageservices/put-block#uri-parameters
	defaultBlockSize = 100 * 1024 * 1024
	blob_url_suffix  = "https://%s.blob.core.windows.net"
)
