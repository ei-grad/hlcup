package models

type Writer interface {
	// Write is expected to write the whole passed []byte and return length and
	// nil error
	Write([]byte) (int, error)
}
