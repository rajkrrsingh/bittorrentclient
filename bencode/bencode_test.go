package bencode

import (
	"reflect"
	"testing"
)

func TestEncodeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test", "4:test"},
		{"", "0:"},
		{"hello world", "11:hello world"},
		{"BitTorrent protocol", "19:BitTorrent protocol"},
	}

	for _, test := range tests {
		result, err := Encode(test.input)
		if err != nil {
			t.Errorf("Encode(%q) returned error: %v", test.input, err)
		}
		if string(result) != test.expected {
			t.Errorf("Encode(%q) = %q, want %q", test.input, string(result), test.expected)
		}
	}
}

func TestEncodeInteger(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{100, "i100e"},
		{0, "i0e"},
		{-5, "i-5e"},
		{42, "i42e"},
	}

	for _, test := range tests {
		result, err := Encode(test.input)
		if err != nil {
			t.Errorf("Encode(%d) returned error: %v", test.input, err)
		}
		if string(result) != test.expected {
			t.Errorf("Encode(%d) = %q, want %q", test.input, string(result), test.expected)
		}
	}
}

func TestEncodeList(t *testing.T) {
	tests := []struct {
		input    []interface{}
		expected string
	}{
		{[]interface{}{"Test", "Data"}, "l4:Test4:Datae"},
		{[]interface{}{}, "le"},
		{[]interface{}{42, "test"}, "li42e4:teste"},
		{[]interface{}{1, 2, 3}, "li1ei2ei3ee"},
	}

	for _, test := range tests {
		result, err := Encode(test.input)
		if err != nil {
			t.Errorf("Encode(%v) returned error: %v", test.input, err)
		}
		if string(result) != test.expected {
			t.Errorf("Encode(%v) = %q, want %q", test.input, string(result), test.expected)
		}
	}
}

func TestEncodeDict(t *testing.T) {
	tests := []struct {
		input    map[string]interface{}
		expected string
	}{
		{
			map[string]interface{}{
				"Status":  "Good",
				"site": "example.com",
			},
			"d6:Status4:Good4:site11:example.come",
		},
		{
			map[string]interface{}{},
			"de",
		},
		{
			map[string]interface{}{
				"number": 42,
				"string": "test",
			},
			"d6:numberi42e6:string4:teste",
		},
	}

	for _, test := range tests {
		result, err := Encode(test.input)
		if err != nil {
			t.Errorf("Encode(%v) returned error: %v", test.input, err)
		}
		if string(result) != test.expected {
			t.Errorf("Encode(%v) = %q, want %q", test.input, string(result), test.expected)
		}
	}
}

func TestEncodeNestedDict(t *testing.T) {
	input := map[string]interface{}{
		"Test Data": map[string]interface{}{
			"Status":  "Good",
			"site": "example.com",
		},
	}
	expected := "d9:Test Datad6:Status4:Good4:site11:example.comee"

	result, err := Encode(input)
	if err != nil {
		t.Errorf("Encode(%v) returned error: %v", input, err)
	}
	if string(result) != expected {
		t.Errorf("Encode(%v) = %q, want %q", input, string(result), expected)
	}
}

func TestDecodeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"4:test", "test"},
		{"0:", ""},
		{"11:hello world", "hello world"},
		{"19:BitTorrent protocol", "BitTorrent protocol"},
	}

	for _, test := range tests {
		result, err := Decode([]byte(test.input))
		if err != nil {
			t.Errorf("Decode(%q) returned error: %v", test.input, err)
		}
		if result != test.expected {
			t.Errorf("Decode(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestDecodeInteger(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"i100e", 100},
		{"i0e", 0},
		{"i-5e", -5},
		{"i42e", 42},
	}

	for _, test := range tests {
		result, err := Decode([]byte(test.input))
		if err != nil {
			t.Errorf("Decode(%q) returned error: %v", test.input, err)
		}
		if result != test.expected {
			t.Errorf("Decode(%q) = %d, want %d", test.input, result, test.expected)
		}
	}
}

func TestDecodeList(t *testing.T) {
	tests := []struct {
		input    string
		expected []interface{}
	}{
		{"l4:Test4:Datae", []interface{}{"Test", "Data"}},
		{"le", []interface{}{}},
		{"li42e4:teste", []interface{}{42, "test"}},
		{"li1ei2ei3ee", []interface{}{1, 2, 3}},
	}

	for _, test := range tests {
		result, err := Decode([]byte(test.input))
		if err != nil {
			t.Errorf("Decode(%q) returned error: %v", test.input, err)
		}
		resultList, ok := result.([]interface{})
		if !ok {
			t.Errorf("Decode(%q) did not return a list", test.input)
			continue
		}
		if len(resultList) != len(test.expected) {
			t.Errorf("Decode(%q) length = %d, want %d", test.input, len(resultList), len(test.expected))
			continue
		}
		for i, item := range resultList {
			if !reflect.DeepEqual(item, test.expected[i]) {
				t.Errorf("Decode(%q)[%d] = %v, want %v", test.input, i, item, test.expected[i])
			}
		}
	}
}

func TestDecodeDict(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]interface{}
	}{
		{
			"d6:Status4:Good4:site11:example.come",
			map[string]interface{}{
				"Status":  "Good",
				"site": "example.com",
			},
		},
		{
			"de",
			map[string]interface{}{},
		},
		{
			"d6:numberi42e6:string4:teste",
			map[string]interface{}{
				"number": 42,
				"string": "test",
			},
		},
	}

	for _, test := range tests {
		result, err := Decode([]byte(test.input))
		if err != nil {
			t.Errorf("Decode(%q) returned error: %v", test.input, err)
		}
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Decode(%q) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestDecodeNestedDict(t *testing.T) {
	input := "d9:Test Datad6:Status4:Good4:site11:example.comee"
	expected := map[string]interface{}{
		"Test Data": map[string]interface{}{
			"Status":  "Good",
			"site": "example.com",
		},
	}

	result, err := Decode([]byte(input))
	if err != nil {
		t.Errorf("Decode(%q) returned error: %v", input, err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Decode(%q) = %v, want %v", input, result, expected)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	testCases := []interface{}{
		"hello world",
		42,
		[]interface{}{"test", 123, "another"},
		map[string]interface{}{
			"key1": "value1",
			"key2": 456,
			"key3": []interface{}{"nested", "list"},
		},
	}

	for _, original := range testCases {
		encoded, err := Encode(original)
		if err != nil {
			t.Errorf("Encode(%v) failed: %v", original, err)
			continue
		}

		decoded, err := Decode(encoded)
		if err != nil {
			t.Errorf("Decode(Encode(%v)) failed: %v", original, err)
			continue
		}

		if !reflect.DeepEqual(decoded, original) {
			t.Errorf("Round trip failed: %v != %v", decoded, original)
		}
	}
}

func TestDecodeErrors(t *testing.T) {
	errorCases := []string{
		"",       // empty input
		"i",      // incomplete integer
		"i42",    // incomplete integer (missing 'e')
		"5:abc",  // string too short
		"l",      // incomplete list
		"d",      // incomplete dict
		"d1:a",   // incomplete dict (missing value)
		"x",      // invalid character
		"i12x3e", // invalid integer
	}

	for _, input := range errorCases {
		_, err := Decode([]byte(input))
		if err == nil {
			t.Errorf("Decode(%q) should have failed but didn't", input)
		}
	}
}
