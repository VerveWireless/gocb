package gocb

import (
	"encoding/json"
	"reflect"
)

type Transcoder interface {
	Decode([]byte, uint32, interface{}) error
	Encode(interface{}) ([]byte, uint32, error)
}

type DefaultTranscoder struct {
}

func (t DefaultTranscoder) Decode(bytes []byte, flags uint32, out interface{}) (err error) {
	// Check for legacy flags
	if flags&cfMask == 0 {
		// Legacy Flags
		if flags == lfJson {
			// Legacy JSON
			flags = cfFmtJson
		} else {
			return clientError{"Unexpected legacy flags value"}
		}
	}

	// Make sure compression is disabled
	if flags&cfCmprMask != cfCmprNone {
		return clientError{"Unexpected value compression"}
	}

	format := flags & cfFmtMask

	switch format {
	case cfFmtBinary, cfFmtString:
		switch out.(type) {
		case *[]byte:
			*(out.(*[]byte)) = bytes
			return nil
		case *string:
			*(out.(*string)) = string(bytes)
			return nil
		case *interface{}:
			*(out.(*interface{})) = interface{}(bytes)
			return nil
		default:
			return clientError{"You must encoding binary in a string or byte array, not " + reflect.ValueOf(out).Type().String()}
		}
	case cfFmtJson:
		err := json.Unmarshal(bytes, &out)
		return err
	default:
		return clientError{"Unexpected flags value"}
	}
	return clientError{"Library error. This should never happen."}
}

func (t DefaultTranscoder) Encode(value interface{}) ([]byte, uint32, error) {
	var bytes []byte
	var flags uint32
	var err error

	switch value.(type) {
	case []byte:
		bytes = value.([]byte)
		flags = cfFmtBinary
	case *[]byte:
		bytes = *value.(*[]byte)
		flags = cfFmtBinary
	case string:
		bytes = []byte(value.(string))
		flags = cfFmtString
	case *string:
		bytes = []byte(*value.(*string))
		flags = cfFmtString
	default:
		bytes, err = json.Marshal(value)
		if err != nil {
			return nil, 0, err
		}
		flags = cfFmtJson
	}

	// No compression supported currently

	return bytes, flags, nil
}
