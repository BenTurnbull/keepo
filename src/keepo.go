package main

import (
	"os"
	"fmt"
	"crypto/sha256"
	"./crypto"
	"./util"
	"./io"
	"encoding/base64"
	"sort"
	"math/rand"
	"time"
	"strconv"
	"path/filepath"
)

const version = 1.0

func main() {
	executable, err := os.Executable()
	util.CheckError(err)
	dir := filepath.Dir(executable)
	store := io.Store{Path:dir}

	arguments := os.Args[1:]
	show, clip := parameterSearch(arguments)
	command := commandSearch(arguments)
	processCommand(store, command, show, clip)
}

func parameterSearch(parameters []string) (show bool, clip bool) {
	show, clip = false, false
	for index := 0; index < len(parameters); index++ {
		switch parameters[index] {
		case "-s", "--show":
			show = true
		case "-c", "--copy":
			clip = true
		}
	}
	return show, clip
}

func commandSearch(parameters []string) (command []string) {
	for index := 0; index < len(parameters); index++ {
		switch parameters[index] {
		case "list", "get", "set":
			return parameters[index:]
		}
	}
	return nil
}

func processCommand(store io.Store, command []string, show bool, clip bool) {
	if command == nil {
		printUsage()
		os.Exit(1)
	} else {

		switch command[0] {
		case "list":
			list(store)

		case "get":
			name := getName(command)

			value := get(store, name)
			util.CheckState(value != nil, fmt.Sprintf("Expected name '%s' to have a value", name))

			if clip {
				err := io.CopyToClipboard(value)
				util.CheckError(err)
			}

			if show || !clip {
				fmt.Printf("%s\n", value)
			}

		case "set":
			name := getName(command)
			value := getValue(command)
			set(store, name, value)

		default:
			printUsage()
			os.Exit(1)
		}
	}
}

func printUsage() {
	boldOpen := "\033[1m"
	boldClose := "\033[0m"
	fmt.Println(
		"version: " + strconv.FormatFloat(version, 'f', 1, 64) +
		"\n\n" +
		"usage: " + boldOpen + "keepo [options] <command>" + boldClose +
		"\n\n" +
		"commands:\n" +
		"\t" + boldOpen + "set <name> [value]" + boldClose + "\t\t" + "sets a name and its value (omit for random value)" +
		"\n\n" +
		"\t" + boldOpen + "get <name>" + boldClose + "\t\t\t" + "gets the value for a name" +
		"\n\n" +
		"\toptions:\n" +
		"\t\t" + boldOpen + "-s, --show" + boldClose + "\t\tsend output to stdout\n" +
		"\t\t" + boldOpen + "-c, --copy" + boldClose + "\t\tcopy output to clipboard" +
		"\n")
}

func getName(command []string) string {
	util.CheckState(len(command) > 1, "need a 'name' argument")
	name := command[1]
	return name
}

func getValue(command []string) string {
	var value string
	if len(command) > 2 {
		value = command[2]
	} else {
		value = getRandomValue()
	}
	return value
}

func getRandomValue() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	bytes := make([]byte, 0, 32)
	r.Read(bytes)

	hash := sha256.New()
	hash.Write(bytes)
	return string(base64.RawURLEncoding.EncodeToString(hash.Sum(nil))[:8])
}

func list(store io.Store) {
	dataMap := store.GetDataMap()
	keys := make([]string, 0, len(dataMap))
	for key := range dataMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, v := range keys {
		fmt.Println(v)
	}
}

func get(store io.Store, name string) []byte {
	dataMap := store.GetDataMap()
	dataValue := dataMap[name]
	util.CheckState(len(dataValue) > 0, "name not found")

	hash := sha256.New()
	hash.Write([]byte(io.ReadPassword()))
	key := hash.Sum(nil)

	value, err := crypto.Decrypt(key, dataValue)
	if err != nil {
		switch err.(type) {
		case base64.CorruptInputError: // not ideal, need to use authenticated encryption
			fmt.Fprintln(os.Stderr, "Could not decrypt, check password")
			return nil
		default:
			util.CheckError(err)
		}
	}

	return value
}

func set(store io.Store, name string, value string) {
	hash := sha256.New()
	hash.Write([]byte(io.ReadPassword()))
	key := hash.Sum(nil)

	dataMap := store.GetDataMap()

	encrypted, err := crypto.Encrypt(key, []byte(value))
	util.CheckError(err)

	dataMap[name] = encrypted

	err = store.SetDataMap(dataMap)
	util.CheckError(err)
}