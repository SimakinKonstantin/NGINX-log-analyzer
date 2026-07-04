package session

import "fmt"

type Error struct {
	msg string
	err error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %v", e.msg, e.err)
}

func (e Error) Unwrap() error {
	return e.err
}
