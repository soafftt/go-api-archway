package utils

import "unsafe"

func ToBytesFromString(s string) []byte {
	if s == "" {
		return nil
	}

	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func ToStringFromBytes(bytes []byte) string {
	if bytes == nil || len(bytes) == 0 {
		return ""
	}

	return unsafe.String(unsafe.SliceData(bytes), len(bytes))
}
