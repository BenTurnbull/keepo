package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"keepo/src/data/input"
	"keepo/src/data/output"
	"keepo/src/data/store"
	"keepo/src/util"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const version = 1.2

func main() {
	arguments := os.Args[1:]
	show, clip, pass := parameterSearch(arguments)
	arguments = commandSearch(arguments)
	processCommand(arguments, show, clip, pass)
}

func printUsage() {
	boldOpen := "\033[1m"
	boldClose := "\033[0m"
	fmt.Println(
		"version: " + boldOpen +  strconv.FormatFloat(version, 'f', 1, 64) + boldClose +
			"\n\n" +
			"Keepo is a utility for managing key-value stores" +
			"\n\n" +
			"usage: " + boldOpen + "keepo [options] <command>" + boldClose +
			"\n\n" +
			"commands:\n\n" +
			"\t" + boldOpen + "list \t[store]" + boldClose + "\t\t\t" + "list status and keys for store (omit for all stores)" +
			"\n\n" +
			"\t" + boldOpen + "set \t[store:]<key> [value]" + boldClose + "\t" + "sets a key and its value (omit for random value)" +
			"\n\n" +
			"\t" + boldOpen + "get \t[store:]<key>" + boldClose + "\t\t" + "gets the value for a key" +
			"\n\n" +
			"\t" + boldOpen + "clear \t[store:]<key>" + boldClose + "\t\t" + "clears the key/value" +
			"\n\n" +
			"\toptions:\n" +
			"\t\t" + boldOpen + "-s, --show" + boldClose + "\t\tsend output to stdout\n" +
			"\t\t" + boldOpen + "-c, --copy" + boldClose + "\t\tcopy output to clipboard\n" +
			"\t\t" + boldOpen + "-p, --pass" + boldClose + "\t\tnext argument will be passphrase" +
			"\n")
}

func parameterSearch(parameters []string) (show bool, clip bool, pass string) {
	show, clip, pass = false, false, ""
	for index := 0; index < len(parameters); index++ {
		switch parameters[index] {
		case "-s", "--show":
			show = true
		case "-c", "--copy":
			clip = true
		case "-p", "--pass":
			pass = parameters[index + 1]
		}
	}
	return show, clip, pass
}

func commandSearch(parameters []string) (command []string) {
	for index := 0; index < len(parameters); index++ {
		switch parameters[index] {
		case "list", "get", "set", "clear":
			return parameters[index:]
		}
	}
	return nil
}

func processCommand(arguments []string, show bool, clip bool, pass string) {
	if arguments == nil {
		printUsage()
		os.Exit(0)
	} else {

		commandWord := arguments[0]
		arguments = arguments[1:]
		switch commandWord {

		case "list":

			if len(arguments) > 0 {
				listStore(arguments[0])
			}

			listAll()

		case "get":
			util.CheckState(len(arguments) > 0, "need a 'key' argument")
			storeName, KeyName := getStoreAndKeyName(arguments[0])

			value := get(storeName, KeyName, pass)
			util.CheckState(value != nil, fmt.Sprintf("expected key '%s' to have a value", KeyName))

			if clip {
				err := output.CopyToClipboard(value)
				util.CheckError(err, "could not copy to clipboard")
			}

			if show || !clip {
				fmt.Printf("%s\n", value)
			}

		case "set":
			util.CheckState(len(arguments) > 0, "need a 'key' argument")
			storeName, KeyName := getStoreAndKeyName(arguments[0])

			value := getValue(arguments[1:])
			set(storeName, KeyName, value, pass)

		case "clear":
			util.CheckState(len(arguments) > 0, "need a 'key' argument")
			storeName, KeyName := getStoreAndKeyName(arguments[0])

			clear(storeName, KeyName, pass)

		default:
			fmt.Printf("unknown command '%s'\n", commandWord)
			printUsage()
			os.Exit(1)
		}
	}
}


func listAll() {
	executable, err := os.Executable()
	util.CheckError(err, "could not get executable path")
	path := filepath.Dir(executable)
	files, err := ioutil.ReadDir(path)
	util.CheckError(err, "could not read current directory")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), store.Extension) {
			listStore(f.Name())
		}
	}
}

func listStore(storeName string) {
	if fi, err := os.Stat(store.GetStorePath(storeName)); err == nil {
		printStatus(fmt.Sprintf("'%s' (%d bytes)", storeName, fi.Size()))
		for _, v := range store.GetMapKeys(storeName) {
			fmt.Println(v)
		}
	} else {
		printStatus(fmt.Sprintf("'%s' is absent", storeName))
	}
}

func printStatus(status string) {
	boldOpen := "\033[1m"
	boldClose := "\033[0m"
	fmt.Println("\nstore: " + boldOpen +  status + boldClose + "\n")
}

func get(storeName, key, pass string) []byte {
	if len(pass) == 0 {
		pass = input.ReadPassword()
	}

	value, err := store.GetMapValue(storeName, key, pass)
	checks("could not get value", err)
	return value
}

func set(storeName, key, value, pass string) {
	if len(pass) == 0 {
		pass = input.ReadPassword()
	}
	err := store.SetMapValue(storeName, key, value, pass)
	checks("could not set value", err)
}

func clear(storeName, key, pass string) {
	if len(pass) == 0 {
		pass = input.ReadPassword()
	}
	err := store.ClearMapValue(storeName, key, pass)
	checks("could not clear value", err)
}

func checks(message string, err error) {
	if state, ok := err.(*store.State); ok && state == store.AuthenticationFailedState {
		fmt.Println("\033[1;31mAuthentication Failed\033[0m")
		os.Exit(1)
	}
	util.CheckError(err, message)
}

func getStoreAndKeyName(argument string) (string, string) {
	values := strings.Split(argument, ":")
	if len(values) > 1 {
		return values[0], values[1]
	}

	return store.DefaultStoreName, argument
}

func getValue(arguments []string) string {
	if len(arguments) > 0 {
		return arguments[0]
	} else {
		return getRandomValue()
	}
}

func getRandomValue() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, 0, 32)
	r.Read(bytes)

	hash := sha256.New()
	hash.Write(bytes)
	return string(base64.RawURLEncoding.EncodeToString(hash.Sum(nil))[:8])
}