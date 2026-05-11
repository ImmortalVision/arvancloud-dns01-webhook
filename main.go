package main

import (
	"os"

	"github.com/immortalvision/arvancloud-dns01-webhook/pkg/arvancloud"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"k8s.io/klog/v2"
)

func main() {
	groupName := os.Getenv("GROUP_NAME")
	if groupName == "" {
		klog.Fatal("GROUP_NAME must be specified")
	}

	cmd.RunWebhookServer(groupName, &arvancloud.Solver{})
}
