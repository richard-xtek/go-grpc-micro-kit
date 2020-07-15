package kafka

import (
	"encoding/json"

	uuid "github.com/satori/go.uuid"
)

// JSONMarshaler ...
type JSONMarshaler struct {
	NewUUID      func() string
	GenerateName func(v interface{}) string
}

// Marshal ...
func (m JSONMarshaler) Marshal(v interface{}) (*Message, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	msg := NewMessage(
		m.newUUID(),
		b,
	)
	msg.Metadata.Set("name", m.Name(v))

	return msg, nil
}

func (m JSONMarshaler) newUUID() string {
	if m.NewUUID != nil {
		return m.NewUUID()
	}

	// default
	return uuid.NewV4().String()
}

// Unmarshal ...
func (JSONMarshaler) Unmarshal(msg *Message, v interface{}) (err error) {
	return json.Unmarshal(msg.Payload, v)
}

// Name ...
func (m JSONMarshaler) Name(cmdOrEvent interface{}) string {
	if m.GenerateName != nil {
		return m.GenerateName(cmdOrEvent)
	}

	return FullyQualifiedStructName(cmdOrEvent)
}

// NameFromMessage ...
func (m JSONMarshaler) NameFromMessage(msg *Message) string {
	return msg.Metadata.Get("name")
}
