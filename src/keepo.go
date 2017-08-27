package main

import (
	"os"
	"fmt"
	"io"
	"bufio"
	"encoding/binary"
	"crypto/sha256"
	"./crypt"
	"./util"
	"encoding/base64"
	"sort"
	"math/rand"
	"time"
	"strconv"
)

const version = 1.0
const storeName = "keepo.dat"

func main() {

	if len(os.Args) == 1 {
		printUsage()
	}

	args := os.Args[2:]
	switch os.Args[1] {
	case "list":
		list()
	case "set":
		checkState(len(args) > 0, "need a 'name' argument")

		if len(args) == 1 { // no value supplied so generate random
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			bytes := make([]byte, 0, 32)
			r.Read(bytes)

			hash := sha256.New()
			hash.Write(bytes)
			value := base64.StdEncoding.EncodeToString(hash.Sum(nil))

			args = append(args, value)
		}
		set(args)

	case "get":
		checkState(len(args) > 0, "need a 'name' argument")

		value := get(args[0])
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

func list() {
	dataMap := getDataMap()
	keys := make([]string, 0, len(dataMap))
	for key := range dataMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, v := range keys {
		fmt.Println(v)
	}
}

func get(name string) []byte {

	dataMap := getDataMap()
	dataValue := dataMap[name]
	checkState(len(dataValue) > 0, "name not found")

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
			checkError(err)
		}
	}

	return value
}

func set(setterArgs []string) {

	hash := sha256.New()
	hash.Write([]byte(util.ReadPassword()))
	key := hash.Sum(nil)

	dataMap := getDataMap()

	value, err := crypt.Encrypt(key, []byte(setterArgs[1]))
	checkError(err)

	dataMap[setterArgs[0]] = value

	err = setDataMap(dataMap)
	checkError(err)
}

func getDataMap() (map[string][]byte) {

	dataMap := make(map[string][]byte)

	// open input file
	if fi, err := os.Open(storeName); err == nil {

		// close fi on exit and check for its returned error
		defer func() {
			if err := fi.Close(); err != nil {
				panic(err)
			}
		}()

		// make a read buffer
		r := bufio.NewReader(fi)

		for true {
			// read key
			keyLength := new(uint32)
			err = binary.Read(r, binary.LittleEndian, keyLength)
			if err == io.EOF {
				return dataMap
			}
			checkError(err)

			keyLengthInt := int(*keyLength)
			checkState(keyLengthInt > 0, "key length may not be negative")

			keyNameBytes := make([]byte, keyLengthInt)
			err = binary.Read(r, binary.LittleEndian, keyNameBytes)
			checkError(err)

			keyName := string(keyNameBytes)

			// read value
			valueLength := new(uint32)
			err = binary.Read(r, binary.LittleEndian, valueLength)
			checkError(err)

			valueLengthInt := int(*valueLength)
			checkState(valueLengthInt > 0, "value Length may not be negative")

			valueBytes := make([]byte, valueLengthInt)
			err = binary.Read(r, binary.LittleEndian, valueBytes)
			checkError(err)

			dataMap[keyName] = valueBytes
		}
	}

	return dataMap
}

func setDataMap(dataMap map[string][]byte) error {

	// open output file
	if fo, err := os.Create(storeName); err == nil {

		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		// make a write buffer
		w := bufio.NewWriter(fo)

		for k, v := range dataMap {

			kLen := len(k)
			err = binary.Write(w, binary.LittleEndian, uint32(kLen))
			checkError(err)

			err = binary.Write(w, binary.LittleEndian, []byte(k))
			checkError(err)

			vLen := len(v)
			err = binary.Write(w, binary.LittleEndian, uint32(vLen))
			checkError(err)

			err = binary.Write(w, binary.LittleEndian, []byte(v))
			checkError(err)

			w.Flush()
		}
	}
	return nil
}

func checkState(expression bool, message string) {
	if !expression {
		os.Stderr.Write([]byte("\033[1;31m" + message + "\033[0m\n"))
		os.Exit(1)
	}
}

func checkError(err error) {
	if err != nil {
		os.Stderr.Write([]byte("\033[1;31m" + err.Error() + "\033[0m\n"))
		os.Exit(1)
	}
}