package main

import (
	"os"
	"fmt"
	"crypto/sha256"
	"./crypt"
	"./util"
	"encoding/base64"
	"sort"
	"math/rand"
	"path"
	"time"
	"strconv"
)

const version = 1.0

func main() {

	if len(os.Args) == 1 {
		printUsage()
	}

	dir := path.Dir(os.Args[0])
	store := util.Store{Path:dir}
	args := os.Args[2:]
	switch os.Args[1] {
	case "list":
		list(store)
	case "set":
		util.CheckState(len(args) > 0, "need a 'name' argument")

		if len(args) == 1 { // no value supplied so generate random
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			bytes := make([]byte, 0, 32)
			r.Read(bytes)

			hash := sha256.New()
			hash.Write(bytes)
			value := base64.StdEncoding.EncodeToString(hash.Sum(nil))

			args = append(args, value)
		}
		set(store, args)

	case "get":
		util.CheckState(len(args) > 0, "need a 'name' argument")

		value := get(store, args[0])
		if value != nil {
			if len(args) > 1 && (args[1] == "-s" || args[1] == "--show"){
				fmt.Printf("%s\n", value)
			}
			err := util.CopyToClipboard(value)
			if err != nil {
				fmt.Printf("%s\n", value)
			}
		}

	default:
		printUsage()
	}
}

func printUsage() {
	boldOpen := "\033[1m"
	boldClose := "\033[0m"
	fmt.Println(
		"version: " + strconv.FormatFloat(version, 'f', 1, 64) +
		"\n" +
		"\n" +
		"usage: " + boldOpen + "keepo <command> [<args>]" + boldClose +
		"\n" +
		"\n" +
		"commands:" +
		"\n" +
		"\t" + boldOpen + "set <name> [value]" + boldClose + "\t\t" + "sets a name and its value (may be omitted for random value)" +
		"\n" +
		"\n" +
		"\t" + boldOpen + "get <name> [option]" + boldClose + "\t\t" + "gets the value for a name" +
		"\n" +
		"\toptions:" +
		"\n" +
		"\t\t" + boldOpen + "-s, --show" + boldClose + "\t\tsend output to stdout" +
		"\n")
	os.Exit(1)
}

func list(store util.Store) {
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

func get(store util.Store, name string) []byte {

	dataMap := store.GetDataMap()
	dataValue := dataMap[name]
	util.CheckState(len(dataValue) > 0, "name not found")

	hash := sha256.New()
	hash.Write([]byte(util.ReadPassword()))
	key := hash.Sum(nil)

	value, err := crypt.Decrypt(key, dataValue)
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

func set(store util.Store, setterArgs []string) {

	hash := sha256.New()
	hash.Write([]byte(util.ReadPassword()))
	key := hash.Sum(nil)

	dataMap := store.GetDataMap()

	value, err := crypt.Encrypt(key, []byte(setterArgs[1]))
	util.CheckError(err)

	dataMap[setterArgs[0]] = value

	err = store.SetDataMap(dataMap)
	util.CheckError(err)
}