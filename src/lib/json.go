package lib

import (
	"encoding/json"
	"io"
)

func JSONEncode(data any, indent bool) ([]byte, error) {
	if indent {
		// MarshalIndent is like Marshal but applies Indent to format the output.
		// Each JSON element in the output will begin on a new line beginning with prefix
		// followed by one or more copies of indent according to the indentation nesting.
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
// If v is nil or not a pointer, Unmarshal returns an [InvalidUnmarshalError].
func JSONDecode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func NewJSONEncoder(w io.Writer) *json.Encoder {
	return json.NewEncoder(w)
}

func NewJSONDecoder(r io.Reader) *json.Decoder {
	return json.NewDecoder(r)
}
