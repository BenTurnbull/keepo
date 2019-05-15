package store

import "fmt"

type State struct {
	code int
	message string
}

var AuthenticationFailedState = &State{10, "authentication failed"}
var ValueAbsentState = &State{11, "value absent"}

func InvalidFormatError(message string) *State {
	return &State{12, fmt.Sprintf("invalid format: %s", message)}
}

func (e *State) Error() string {
	return fmt.Sprintf("%s", e.message)
}