package parser

import "fmt"

type WrapError struct {
	msg string
	err error
}

func (e WrapError) Error() string {
	return fmt.Sprintf("%s: %v", e.msg, e.err)
}

func (e WrapError) Unwrap() error {
	return e.err
}

type Error struct {
	msg string
}

func (e Error) Error() string {
	return e.msg
}
