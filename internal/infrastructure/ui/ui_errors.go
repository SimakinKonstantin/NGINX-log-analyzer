package ui

import "fmt"

type ArgError struct {
	msg string
}

func (e ArgError) Error() string {
	return e.msg
}

type ArgWrapError struct {
	msg string
	err error
}

func (e ArgWrapError) Error() string {
	return fmt.Sprintf("%s: %v", e.msg, e.err)
}

func (e ArgWrapError) Unwrap() error {
	return e.err
}

type FileWriteError struct {
	msg string
	err error
}

func (e FileWriteError) Error() string {
	return fmt.Sprintf("%s: %v", e.msg, e.err)
}

func (e FileWriteError) Unwrap() error {
	return e.err
}
