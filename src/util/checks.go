package util

import (
	"log"
	"os"
)

func CheckState(expression bool, message string) {
	if !expression {
		_, err := os.Stderr.Write([]byte("\033[1;31m" + message + "\033[0m\n"))
		if err != nil {
			log.Println("could not write state message: " + err.Error())
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func CheckError(err error, message string) {
	if err != nil {
		_, err := os.Stderr.Write([]byte("\033[1;31m" + message + " (" + err.Error() + ")\033[0m\n"))
		if err != nil {
			log.Println("could not write error message: " + err.Error())
			os.Exit(2)
		}
		os.Exit(1)
	}
}
