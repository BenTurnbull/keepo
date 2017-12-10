package io

import (
	"bufio"
	"encoding/binary"
	"os"
	"io"
	"path/filepath"
	"../util"
)

const storeName = "keepo.dat"

type Store struct {
	Path string
}

func getStorePath(store Store) string {
	return store.Path + string(filepath.Separator) + storeName
}

func (store Store) GetDataMap() (map[string][]byte) {

	dataMap := make(map[string][]byte)

	// open input file
	storePath := getStorePath(store)
	if fi, err := os.Open(storePath); err == nil {

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
			util.CheckError(err)

			keyLengthInt := int(*keyLength)
			util.CheckState(keyLengthInt > 0, "key length may not be negative")

			keyNameBytes := make([]byte, keyLengthInt)
			err = binary.Read(r, binary.LittleEndian, keyNameBytes)
			util.CheckError(err)

			keyName := string(keyNameBytes)

			// read value
			valueLength := new(uint32)
			err = binary.Read(r, binary.LittleEndian, valueLength)
			util.CheckError(err)

			valueLengthInt := int(*valueLength)
			util.CheckState(valueLengthInt > 0, "value Length may not be negative")

			valueBytes := make([]byte, valueLengthInt)
			err = binary.Read(r, binary.LittleEndian, valueBytes)
			util.CheckError(err)

			dataMap[keyName] = valueBytes
		}
	}

	return dataMap
}

func (store Store) SetDataMap(dataMap map[string][]byte) error {

	// open output file
	storePath := getStorePath(store)
	if fo, err := os.Create(storePath); err == nil {

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
			util.CheckError(err)

			err = binary.Write(w, binary.LittleEndian, []byte(k))
			util.CheckError(err)

			vLen := len(v)
			err = binary.Write(w, binary.LittleEndian, uint32(vLen))
			util.CheckError(err)

			err = binary.Write(w, binary.LittleEndian, []byte(v))
			util.CheckError(err)

			w.Flush()
		}

	} else {
		return err
	}

	return nil
}



