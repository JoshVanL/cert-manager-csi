package options

import (
	"github.com/spf13/cobra"
)

type Options struct {
	NodeID     string
	DriverName string

	// CSI driver endpoint.
	Endpoint string

	// Root directory to write data and mount from.
	DataRoot string

	// Endpoint that Kubelet should connect to driver.
	KubeletRegistrationEndpoint string

	// Size in Mbytes to create the tmpfs file system to write and mount from.
	TmpfsSize string

	// Optional Webhook configuration
	Webhook Webhook
}

// Optional Webhook configuration
type Webhook struct {
	// URL to server to consume Webhook Create, Renew, Destroy
	NetHost string
}

func AddFlags(cmd *cobra.Command) *Options {
	var opts Options

	cmd.PersistentFlags().StringVar(&opts.NodeID, "node-id", "", "node ID")
	cmd.MarkPersistentFlagRequired("node-id")

	cmd.PersistentFlags().StringVar(&opts.Endpoint, "endpoint", "", "CSI endpoint")
	cmd.MarkPersistentFlagRequired("endpoint")

	cmd.PersistentFlags().StringVar(&opts.DriverName, "driver-name",
		"csi.cert-manager.io", "name of the driver")

	cmd.PersistentFlags().StringVar(&opts.DataRoot, "data-root",
		"/csi-data-dir", "directory to store ephemeral data")

	cmd.PersistentFlags().StringVar(&opts.TmpfsSize, "tmpfs-size",
		"100", "size in Mbytes to create the tmpfs file system to store ephemeral data")

	addWebhookFlags(cmd, &opts)

	return &opts
}

func addWebhookFlags(cmd *cobra.Command, opts *Options) {
	cmd.PersistentFlags().StringVar(&opts.Webhook.NetHost, "webhook-net-host",
		"", "optional URL to a server to consume Create,Renew,Destroy webhooks for certificates")
}
