package app

import "errors"

var (
	errEmptyInt               = errors.New("empty integer")
	errUnexpectedFirstChar    = errors.New("unexpected first char found. Expecting 0-9")
	errUnexpectedTrailingChar = errors.New("unexpected traling char found. Expecting 0-9")
	errTooLongInt             = errors.New("too long int")
)

var maxIntChars = 10

func parseUint32(b []byte) (uint32, error) {
	n := len(b)
	if n == 0 {
		return 0, errEmptyInt
	}
	var v uint32
	for i := 0; i < n; i++ {
		c := b[i]
		k := c - '0'
		if k > 9 {
			if i == 0 {
				return 0, errUnexpectedFirstChar
			}
			return 0, errUnexpectedTrailingChar
		}
		if i >= maxIntChars {
			return 0, errTooLongInt
		}
		v = 10*v + uint32(k)
	}
	if n != len(b) {
		return 0, errUnexpectedTrailingChar
	}
	return v, nil
}
