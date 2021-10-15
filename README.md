# Azure Encrypted Object Store - Velero Plugin
![example branch main](https://github.com/equinor/velero-plugin-aeos/actions/workflows/build.yml/badge.svg)
![example branch main](https://github.com/equinor/velero-plugin-aeos/actions/workflows/docker-publish.yml/badge.svg) 

## Overview and Features
This plugin provides an object store that encrypts all k8s objects with a client provided key when storing them in Azure. It can be used either as a replacement for or a supplement to the official velero azure plugin.

## Installation and Usage
Usage is similar to official azure plugin. Docs

Currently, the only access keys and aad-pod-identity based authentication are supported by this plugin. SP based auth may come in the future.

1. Create a secrets file named 'credentials-velero' with the following keys. This secrets file can be shared with the official azure plugin.
```
AZURE_STORAGE_ACCOUNT_ACCESS_KEY=
AZURE_STORAGE_ACCOUNT_ENCRYPTION_KEY=
AZURE_STORAGE_ACCOUNT_ENCRYPTION_HASH=
```

Or for aad-pod-identities:
```
AZURE_SUBSCRIPTION_ID=
AZURE_STORAGE_ACCOUNT_ID=
AZURE_STORAGE_ACCOUNT_ENCRYPTION_KEY=
AZURE_STORAGE_ACCOUNT_ENCRYPTION_HASH=
```


2. Generate a random encryption key and hash using the hack/genkey.py script and insert them into the secrets file. 
```
/bin/python3 hack/keygen.py bits256
```

3. If installing using the velero cli:
```
velero install \
    --provider equinor/velero-plugin-aeos \
    --plugins ghcr.io/equinor/velero-plugin-aeos:latest \
    --bucket $BLOB_CONTAINER \
    --secret-file ./credentials-velero \
    --backup-location-config storageAccount=$AZURE_STORAGE_ACCOUNT_ID,storageAccountKeyEnvVar=AZURE_STORAGE_ACCOUNT_ACCESS_KEY \
    --use-volume-snapshots=false
```
## Using AEOS with the official Azure Plugin
Coming soon.

## Unit Tests
Builds are tested against a real azure storage account. See the 'Build' badge for the current status of the build and test workflow on the main branch.
