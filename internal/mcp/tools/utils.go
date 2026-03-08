package tools

import (
	"github.com/mitchellh/mapstructure"
)

// DecodeParams safely decodes a dynamic map[string]interface{} into a strongly-typed Go struct.
// It uses mapstructure with WeaklyTypedInput enabled to automatically handle common JSON
// unmarshaling quirks, such as decoding float64 into int fields.
func DecodeParams(input map[string]interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,   // Automatically converts float64 (JSON numbers) to int
		TagName:          "json", // Use json tags for mapping
		ErrorUnused:      false,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
