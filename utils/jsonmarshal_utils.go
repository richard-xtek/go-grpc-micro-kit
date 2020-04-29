package utils

import (
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap/zapcore"
)

var (
	// JsonPbMarshaller is the marshaller used for serializing protobuf messages.
	jsonPbMarshaller = &jsonpb.Marshaler{}
)

// JsonpbObjectMarshaler ...
type JsonpbObjectMarshaler struct {
	Pb proto.Message
}

// MarshalLogObject ...
func (j JsonpbObjectMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected("msg", j)
}

// MarshalJSON ...
func (j JsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := jsonPbMarshaller.Marshal(b, j.Pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}
