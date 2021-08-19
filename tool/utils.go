package tool

import (
	"encoding/json"
	"fmt"
	"os"
	"unicode"
)

// EncodeULEB128 appends v to b using unsigned LEB128 encoding.
func EncodeULEB128(v uint32, stream *Stream) (out []byte, err error) {
	for {
		c := uint8(v & 0x7f)
		v >>= 7
		if v > 0 {
			c |= 0x80
		}
		out = append(out, c)
		if v == 0 {
			break
		}
	}
	_, err = stream.Write(out)
	if err != nil {
		return nil, fmt.Errorf("EncodeULEB128 error: %w", err)
	}
	return out, nil
}

// EncodeSLEB128 appends v to b using signed LEB128 encoding.
func EncodeSLEB128(v int32, stream *Stream) (out []byte, err error) {
	for {
		c := uint8(v & 0x7f)
		s := uint8(v & 0x40)
		v >>= 7

		if (v != -1 || s == 0) && (v != 0 || s != 0) {
			c |= 0x80
		}

		out = append(out, c)

		if c&0x80 == 0 {
			break
		}
	}

	_, err = stream.Write(out)
	if err != nil {
		return nil, fmt.Errorf("EncodeSLEB128 error: %w", err)
	}
	return out, nil
}

// DecodeULEB128 decodes bytes from stream with unsigned LEB128 encoding.
func DecodeULEB128(stream *Stream) (u uint32, err error) {
	var shift uint
	for {
		b, err := stream.ReadByte()
		if err != nil {
			return 0, err
		}
		u |= uint32(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}

	return
}

// DecodeSLEB128 decodes bytes from stream with signed LEB128 encoding.
func DecodeSLEB128(stream *Stream) (s int32, err error) {
	var shift uint
	for {
		b, err := stream.ReadByte()
		if err != nil {
			return 0, err
		}
		s |= int32(b&0x7f) << shift
		shift += 7
		if b&0x80 == 0 {
			// If it's signed
			if b&0x40 != 0 {
				s |= ^0 << shift
			}
			break
		}
	}

	return
}

func ReadFromFile(path string) (JSON, error) {
	obj := make(JSON)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read from file error: %w", err)
	}

	jsonParser := json.NewDecoder(file)
	if err := jsonParser.Decode(&obj); err != nil {
		return nil, fmt.Errorf("read from file error: %w", err)
	}
	return obj, nil
}

// Camel-Case to underline
func Lcfirst(str string) string {
	//if z := rune(str[0]); unicode.IsLower(z) {
	//	return string(unicode.ToUpper(z)) + str[1:]
	//} else {
	//	return str
	//}
	var newStr string
	for i, b := range str {
		if unicode.IsUpper(b) {
			cm := string(unicode.ToLower(b))
			if i != 0 {
				cm = "_" + cm
			}
			newStr += cm
		} else {
			newStr += string(b)
		}
	}
	return newStr
}

// underline to Camel-Case
func Ucfirst(str string) string {
	//if z := rune(str[0]); unicode.IsUpper(z) {
	//	return string(unicode.ToLower(z)) + str[1:]
	//} else {
	//	return str
	//}
	var newStr string
	for i := 0; i < len(str); i++ {
		b := str[i]
		if unicode.IsLower(rune(b)) && i == 0 {
			newStr += string(unicode.ToUpper(rune(b)))
		} else if b == '_' {
			n := str[i+1]
			if unicode.IsLower(rune(n)) {
				newStr += string(unicode.ToUpper(rune(n)))
				i += 1
			}
		} else {
			newStr += string(b)
		}
	}
	return newStr
}

func Interface2Bytes(arr interface{}) (out []byte) {
	switch v := arr.(type) {
	case []interface{}:
		for _, b := range v {
			out = append(out, byte(b.(float64)))
		}
	case []byte:
		out = v
	case string:
		out = []byte(v)
	}
	return
}
