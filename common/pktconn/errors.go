package pktconn

import (
	"errors"
	"io"
)

var (
	ErrPayloadTooLarge = io.ErrShortWrite
	ErrPayloadTooSmall = io.ErrUnexpectedEOF

	errPayloadTooLarge = errors.New("payload too large")
	errChecksumError   = errors.New("checksum error")
)

type timeoutError interface {
	Timeout() bool // Is it a timeout error
}

type temperaryError interface {
	Temporary() bool
}

// IsTimeout checks if the error is a timeout error
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	err = Cause(err)
	ne, ok := err.(timeoutError)
	return ok && ne.Timeout()
}

// IsTimeout checks if the error is a timeout error
func IsTemporary(err error) bool {
	if err == nil {
		return false
	}

	err = Cause(err)
	ne, ok := err.(temperaryError)
	return ok && ne.Temporary()
}

func Cause(err error) error {
	type causer interface {
		Cause() error
	}
	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
