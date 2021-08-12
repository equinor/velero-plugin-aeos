package main

import (
	"github.com/equinor/velero-plugin-aeos/internal/plugin"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/velero/pkg/plugin/framework"
)

func main() {
	framework.NewServer().
		RegisterObjectStore("equinor/velero-plugin-aeos", newObjectStorePlugin).
		Serve()
}

func newObjectStorePlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugin.NewFileObjectStore(logger), nil
}
