package output

import (
	"os"
	"os/exec"
	"strconv"
)

type WriteError int

func (k WriteError) Error() string {
	return "only wrote portion of length '" + strconv.Itoa(int(k)) + "' to clipboard input"
}

func CopyToClipboard(input []byte) error {

	command := exec.Command("xsel", "--input", "--clipboard")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	inputPipe, err := command.StdinPipe()
	if err != nil {
		return err
	}
	defer func() {
		if err := inputPipe.Close(); err != nil {
			panic(err)
		}
	}()

	err = command.Start()
	if err != nil {
		return err
	}

	written, err := inputPipe.Write(input)
	if written != len(input) {
		return WriteError(written)
	}
	if err != nil {
		return err
	}

	return nil
}