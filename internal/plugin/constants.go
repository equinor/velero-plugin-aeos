package plugin

const (
	storageAccountConfigKey          = "storageAccount"
	storageAccountKeyEnvVarConfigKey = "storageAccountKeyEnvVar"
	subscriptionIDConfigKey          = "subscriptionId"
	blockSizeConfigKey               = "blockSizeInBytes"

	// blocks must be less than/equal to 100MB in size
	// ref. https://docs.microsoft.com/en-us/rest/api/storageservices/put-block#uri-parameters
	defaultBlockSize = 100 * 1024 * 1024
	blob_url_suffix  = "https://%s.blob.core.windows.net"
)
