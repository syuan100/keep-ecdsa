package tss

import (
	"encoding/json"
	"fmt"
	"github.com/keep-network/keep-tecdsa/pkg/net"
)

// MessageRouting holds the information required to route a message. It determines
// a type of routing based on the receiver. If receiver is `nil` it will assume
// to broadcast the message. If receiver is provided it will send direct unicast
// message to the receiver. It holds the message itself as well.
type MessageRouting struct {
	ReceiverID string
	Message    net.TaggedMarshaler
}

// ProtocolMessage is a network message used to transport messages generated in
// TSS protocol execution. It is a wrapper over a message generated by underlying
// implementation of the protocol.
type ProtocolMessage struct {
	SenderPublicKey []byte
	Payload         []byte
	IsBroadcast     bool
}

// Type returns a string type of the `TSSMessage` so that it conforms to
// `net.Message` interface.
func (m *ProtocolMessage) Type() string {
	return fmt.Sprintf("%T", m)
}

// Marshal converts this message to a byte array suitable for network communication.
func (m *ProtocolMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal converts a byte array produced by Marshal to a message.
func (m *ProtocolMessage) Unmarshal(bytes []byte) error {
	var message ProtocolMessage
	if err := json.Unmarshal(bytes, &message); err != nil {
		return err
	}

	m.SenderPublicKey = message.SenderPublicKey
	m.Payload = message.Payload
	m.IsBroadcast = message.IsBroadcast

	return nil
}

// JoinMessage is a network message used to notify peer members about readiness
// to start protool execution.
type JoinMessage struct {
	SenderPublicKey []byte
}

// Type returns a string type of the `JoinMessage`.
func (m *JoinMessage) Type() string {
	return fmt.Sprintf("%T", m)
}

// Marshal converts this message to a byte array suitable for network communication.
func (m *JoinMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal converts a byte array produced by Marshal to a message.
func (m *JoinMessage) Unmarshal(bytes []byte) error {
	var message JoinMessage
	if err := json.Unmarshal(bytes, &message); err != nil {
		return err
	}

	m.SenderPublicKey = message.SenderPublicKey

	return nil
}
