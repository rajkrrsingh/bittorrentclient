package bencode

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strconv"
)

var (
	ErrInvalidInput  = errors.New("invalid bencode input")
	ErrUnexpectedEnd = errors.New("unexpected end of input")
)

func Encode(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := encode(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encode(buf *bytes.Buffer, data interface{}) error {
	switch v := data.(type) {
	case string:
		buf.WriteString(strconv.Itoa(len(v)))
		buf.WriteByte(':')
		buf.WriteString(v)
	case []byte:
		buf.WriteString(strconv.Itoa(len(v)))
		buf.WriteByte(':')
		buf.Write(v)
	case int:
		buf.WriteByte('i')
		buf.WriteString(strconv.Itoa(v))
		buf.WriteByte('e')
	case int64:
		buf.WriteByte('i')
		buf.WriteString(strconv.FormatInt(v, 10))
		buf.WriteByte('e')
	case []interface{}:
		buf.WriteByte('l')
		for _, item := range v {
			if err := encode(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	case map[string]interface{}:
		buf.WriteByte('d')

		// Sort keys lexicographically
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			if err := encode(buf, key); err != nil {
				return err
			}
			if err := encode(buf, v[key]); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	default:
		return fmt.Errorf("unsupported type: %T", data)
	}
	return nil
}

func Decode(data []byte) (interface{}, error) {
	result, _, err := decode(data, 0)
	return result, err
}

func decode(data []byte, start int) (interface{}, int, error) {
	if start >= len(data) {
		return nil, start, ErrUnexpectedEnd
	}

	switch data[start] {
	case 'i':
		return decodeInt(data, start)
	case 'l':
		return decodeList(data, start)
	case 'd':
		return decodeDict(data, start)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return decodeString(data, start)
	default:
		return nil, start, ErrInvalidInput
	}
}

func decodeInt(data []byte, start int) (interface{}, int, error) {
	if start >= len(data) || data[start] != 'i' {
		return nil, start, ErrInvalidInput
	}

	end := start + 1
	for end < len(data) && data[end] != 'e' {
		end++
	}

	if end >= len(data) {
		return nil, start, ErrUnexpectedEnd
	}

	numStr := string(data[start+1 : end])
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return nil, start, fmt.Errorf("invalid integer: %s", numStr)
	}

	return int(num), end + 1, nil
}

func decodeString(data []byte, start int) (interface{}, int, error) {
	colon := start
	for colon < len(data) && data[colon] != ':' {
		colon++
	}

	if colon >= len(data) {
		return nil, start, ErrUnexpectedEnd
	}

	lengthStr := string(data[start:colon])
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, start, fmt.Errorf("invalid string length: %s", lengthStr)
	}

	if colon+1+length > len(data) {
		return nil, start, ErrUnexpectedEnd
	}

	str := string(data[colon+1 : colon+1+length])
	return str, colon + 1 + length, nil
}

func decodeList(data []byte, start int) (interface{}, int, error) {
	if start >= len(data) || data[start] != 'l' {
		return nil, start, ErrInvalidInput
	}

	var list []interface{}
	pos := start + 1

	for pos < len(data) && data[pos] != 'e' {
		item, newPos, err := decode(data, pos)
		if err != nil {
			return nil, start, err
		}
		list = append(list, item)
		pos = newPos
	}

	if pos >= len(data) {
		return nil, start, ErrUnexpectedEnd
	}

	return list, pos + 1, nil
}

func decodeDict(data []byte, start int) (interface{}, int, error) {
	if start >= len(data) || data[start] != 'd' {
		return nil, start, ErrInvalidInput
	}

	dict := make(map[string]interface{})
	pos := start + 1

	for pos < len(data) && data[pos] != 'e' {
		// Decode key (must be a string)
		key, newPos, err := decode(data, pos)
		if err != nil {
			return nil, start, err
		}

		keyStr, ok := key.(string)
		if !ok {
			return nil, start, errors.New("dictionary key must be a string")
		}

		// Decode value
		value, newPos, err := decode(data, newPos)
		if err != nil {
			return nil, start, err
		}

		dict[keyStr] = value
		pos = newPos
	}

	if pos >= len(data) {
		return nil, start, ErrUnexpectedEnd
	}

	return dict, pos + 1, nil
}
