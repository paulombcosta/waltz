package transfer

import "github.com/paulombcosta/waltz/provider"

type TransferClient struct {
	Provider provider.Provider
}

func Transfer(provider provider.Provider, tracks []string) TransferClient {
	return TransferClient{Provider: provider}
}

func (t TransferClient) To(provider provider.Provider) error {
	return nil
}
