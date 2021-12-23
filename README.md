# Velero Plugin: Azure Encrypted Object Store
![example branch main](https://github.com/equinor/velero-plugin-aeos/actions/workflows/build.yml/badge.svg)
![example branch main](https://github.com/equinor/velero-plugin-aeos/actions/workflows/docker-publish.yml/badge.svg)

## Overview
This plugin provides a velero object store that creates [client-side encrypted backups](https://docs.microsoft.com/en-us/azure/storage/common/storage-client-side-encryption?toc=%2Fazure%2Fstorage%2Fblobs%2Ftoc.json&tabs=dotnet) for Microsoft Azure Storage Accounts and fills a feature gap in the official [Azure plugin](https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure).

## FAQ
**Q: Why do I need this plugin?**

A: This plugin is useful for environments with strict compliance requirements or for preventing secret leakage via unencrypted backups.

**Q: Why is this feature missing from the official plugin?**

A: The official Azure plugin uses the older [azure-sdk-for-go](https://github.com/Azure/azure-sdk-for-go) which does not have the ability to create client-side encrypted backups. This plugin uses the newer (but still in preview) [azure-storage-blob-go](https://github.com/Azure/azure-storage-blob-go) to implement the velero object store. For this reason, client-side encrypted backups are unlikely to be implemented in the official plugin in the near future.

**Q: This plugin only contains an object store. How do I create volume snapshots?**

A: This plugin can be used with the official plugin. See the instructions below.

## Installation and Usage
Usage is similar to official azure plugin.

Currently, only access key and MSI / pod-identity based authentication are supported by this plugin. Service Principal based auth may come in the future, if requested.

1. Generate a random encryption key and hash using the hack/genkey.py script.
```
/bin/python3 hack/keygen.py bits256
Key    (ASCII)  :       iOsvHDudvugULWyAKnhtvakBBgqkjSNk
Key     (B64)   :       aU9zdkhEdWR2dWdVTFd5QUtuaHR2YWtCQmdxa2pTTms=
KeyHash (B64)   :       LFe0unbGu/arNngmQpJm3edzq+nmy0wRrQReup9DLVY=
```

2. Create a secrets file named 'credentials-velero' with the following keys. This secrets file can be shared with the official azure plugin.
```
AZURE_STORAGE_ACCOUNT_ACCESS_KEY=
AZURE_STORAGE_ACCOUNT_ENCRYPTION_KEY=aU9zdkhEdWR2dWdVTFd5QUtuaHR2YWtCQmdxa2pTTms
AZURE_STORAGE_ACCOUNT_ENCRYPTION_HASH=LFe0unbGu/arNngmQpJm3edzq+nmy0wRrQReup9DLVY=
```

Or for MSI / pod-identities:
```
AZURE_STORAGE_ACCOUNT_ENCRYPTION_KEY=aU9zdkhEdWR2dWdVTFd5QUtuaHR2YWtCQmdxa2pTTms
AZURE_STORAGE_ACCOUNT_ENCRYPTION_HASH=LFe0unbGu/arNngmQpJm3edzq+nmy0wRrQReup9DLVY=
```

3. If installing using the velero cli:
```
velero install \
    --provider equinor/velero-plugin-aeos \
    --plugins ghcr.io/equinor/velero-plugin-aeos:latest \
    --bucket velero \
    --secret-file ./credentials-velero \
    --backup-location-config storageAccount=mySA,resourceGroup=myRG,subscriptionId=00000000-0000-0000-0000-000000000000
    --use-volume-snapshots=false
```
## Using AEOS with the official Azure Plugin
A suggested use case for this plugin is to configure velero to use the AEOS plugin to encypt k8s secrets and the official plugin for the remaining k8s resources and persistent volumes. You will need a storage account with two containers: one for the official plugin and the other for AEOS.

1. Install velero using the official plugin

2. Install AEOS with the following command:
```
velero plugin add ghcr.io/equinor/velero-plugin-aeos:main
```

3. Modify the secrets file according the AEOS installation instructions. The official plugin and AEOS can share the same secrets file.

4. Create an additional BackupStorageLocation resource that points to the second storage account container. You will also need to change the provider field to "equinor/velero-plugin-aeos". See the example below.
```
apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: velero
  namespace: velero
spec:
  config:
    resourceGroup: myRg
    storageAccount: mySA
    subscriptionId: 00000000-0000-0000-0000-000000000000
  default: true
  objectStorage:
    bucket: velero
  provider: azure
---
apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: velero-secrets
  namespace: velero
spec:
  config:
    resourceGroup: myRg
    storageAccount: mySA
    subscriptionId: 00000000-0000-0000-0000-000000000000
  default: true
  objectStorage:
    bucket: velero-secrets
  provider: equinor/velero-plugin-aeos
```

5. (Recommended) Create Schedule resources with filters to determine how api resources are backed up. You will need a schedule for each BackupStorageLocation. Example:
```
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: default
  namespace: velero
spec:
  schedule: '@every 24h'
  template:
    includeClusterResources: true
    excludedResources:
      - secrets
    snapshotVolumes: true
    storageLocation: velero
    ttl: 1440h0m0s
---
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: default-secrets
  namespace: velero
spec:
  schedule: '@every 24h'
  template:
    includeClusterResources: false
    excludedResources:
      - *
    includedResources:
      - secrets
    snapshotVolumes: false
    storageLocation: velero-secrets
    ttl: 1440h0m0s
```
## Unit Tests
Builds are tested against a real azure storage account. See the 'Build' badge for the current status of the build and test workflow on the main branch.

