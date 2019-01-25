package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"keepo/src/data/input"
	"keepo/src/data/output"
	"keepo/src/data/store"
	"keepo/src/util"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const version = 1.0

func main() {
	executable, err := os.Executable()
	util.CheckError(err, "could not get executable path")
	dir := filepath.Dir(executable)

	arguments := os.Args[1:]
	show, clip := parameterSearch(arguments)
	command := commandSearch(arguments)
	processCommand(dir, command, show, clip)
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
		case "convert", "list", "get", "set":
			return parameters[index:]
		}
	}
	return nil
}

func processCommand(path string, command []string, show bool, clip bool) {
	if command == nil {
		printUsage()
		os.Exit(1)
	} else {

		switch command[0] {
		/*case "convert":
			convert(store, getName(command))*/

		case "list":
			list(path)

		case "get":
			name := getName(command)

			value := get(path, name)
			util.CheckState(value != nil, fmt.Sprintf("Expected name '%s' to have a value", name))

			if clip {
				err := output.CopyToClipboard(value)
				util.CheckError(err, "could not copy to clipboard")
			}

			if show || !clip {
				fmt.Printf("%s\n", value)
			}

		case "set":
			name := getName(command)
			value := getValue(command)
			set(path, name, value)

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

func list(path string) {
	for _, v := range store.GetMapKeysV1(path) {
		fmt.Println(v)
	}
}

func get(path, name string) []byte {
	value, err := store.GetMapValueV1(path, name, input.ReadPassword())
	if err != nil {
		switch err.(type) {
		case base64.CorruptInputError: // not ideal, need to use authenticated encryption
			_, _ = fmt.Fprintln(os.Stderr, "Could not decrypt, check password")
			return nil
		default:
			util.CheckError(err, "could not decrypt value")
		}
	}

	return value
}

func set(path, name string, value string) {
	err := store.SetMapValueV1(path, name, value, input.ReadPassword())
	util.CheckError(err, "could not save store")
}

/*func convert(store store.StorePath, existingKey string) {
	dataMap := store.GetDataMapV1()
	convertedDataMap := make(map[string][]byte)
	key := crypto.GetHash([]byte(existingKey))

	for k, v := range dataMap {
		value, err := crypto.DecryptCFB(key[:], v)
		util.CheckError(err, "could not decrypt store value")
		convertedDataMap[k] = value
	}

	fixedKey := crypto.GenerateKey()
	log.Printf("fixed key %s", hex.EncodeToString(fixedKey[:]))

	var nonce [crypto.NonceSize]byte
	for k, v := range dataMap {
		nonce = crypto.GenerateNonce()
		out := make([]byte, len(nonce))
		copy(out, nonce[:])

		out = secretbox.Seal(out, v, &nonce, &fixedKey)
		convertedDataMap[k] = out
	}

	nonce = crypto.GenerateNonce()
	out := secretbox.Seal(nonce[:], fixedKey[:], &nonce, &key)
	err := store.SetV2(out, convertedDataMap)
	util.CheckError(err, "could not save converted store")
}*/