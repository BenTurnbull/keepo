package util

import "os"

func CheckState(expression bool, message string) {
	if !expression {
		os.Stderr.Write([]byte("\033[1;31m" + message + "\033[0m\n"))
		os.Exit(1)
	}
}

func CheckError(err error) {
	if err != nil {
		os.Stderr.Write([]byte("\033[1;31m" + err.Error() + "\033[0m\n"))
		os.Exit(1)
	}
}
