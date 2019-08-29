package local

import (
	"fmt"

	"github.com/keep-network/keep-tecdsa/pkg/chain/eth"
)

type localKeep struct {
	publicKey [64]byte

	signatureRequestedHandlers map[int]func(event *eth.SignatureRequestedEvent)
}

func (c *localChain) requestSignature(keepAddress eth.KeepAddress, digest [32]byte) error {
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()

	keep, ok := c.keeps[keepAddress]
	if !ok {
		return fmt.Errorf(
			"keep not found for address [%s]",
			keepAddress.String(),
		)
	}

	signatureRequestedEvent := &eth.SignatureRequestedEvent{
		Digest: digest,
	}

	for _, handler := range keep.signatureRequestedHandlers {
		go func(handler func(event *eth.SignatureRequestedEvent), signatureRequestedEvent *eth.SignatureRequestedEvent) {
			handler(signatureRequestedEvent)
		}(handler, signatureRequestedEvent)
	}

	return nil
}
