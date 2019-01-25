package store

import (
	"bufio"
	"encoding/binary"
	"io"
	"keepo/src/util"
	"os"
)

/**
 * Store layout:
 *
 * store-header (meta-data):
 * header count n  - uint32 (the number of entries)
 *
 * header-key-length-1 - uint32
 * header-key-value-1  - header-key-length-1 bytes
 * header-length-1 - uint32
 * header-value-1  - header-length-1 bytes
 * ...
 * header-key-length-n - uint32
 * header-key-value-n  - header-key-length-1 bytes
 * header-length-n - uint32
 * header-value-n  - header-length-n bytes
 *
 * store-content (data):
 *
 * data-key-length-1 - uint32
 * data-key-value-1  - data-key-length-1 bytes
 * data-length-1 - uint32
 * data-value-1  - data-length-1 bytes
 * ...
 * data-key-length-n - uint32
 * data-key-value-n  - data-key-length-1 bytes
 * data-length-n - uint32
 * data-value-n  - data-length-n bytes
 *
 */

func get(path string) (key []byte, data map[string][]byte, err error) {

	// open input file
	if fi, err := os.Open(path); err == nil {

		// close fi on exit and check for its returned error
		defer func() {
			if err := fi.Close(); err != nil {
				panic(err)
			}
		}()

		// make a read buffer
		r := bufio.NewReader(fi)

		// read the key
		keyLength := new(uint32)
		err = binary.Read(r, binary.LittleEndian, keyLength)
		if err == io.EOF {
			return key, data, err
		}
		util.CheckError(err, "could not read key length")

		key = make([]byte, *keyLength)
		err = binary.Read(r, binary.LittleEndian, key)
		util.CheckError(err, "could not read key")

		// read the data
		data, err = getDataMap(r)
		util.CheckError(err, "could not read data")
	}

	return key, data, err
}

func set(path string, key []byte, dataMap map[string][]byte) error {

	// open output file
	if fo, err := os.Create(path); err == nil {

		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		// make a write buffer
		w := bufio.NewWriter(fo)

		// write the key
		kLen := len(key)
		err = binary.Write(w, binary.LittleEndian, uint32(kLen))
		util.CheckError(err, "could not write key length")

		err = binary.Write(w, binary.LittleEndian, []byte(key))
		util.CheckError(err, "could not write key")

		// write the data
		err := setDataMap(w, dataMap)
		util.CheckError(err, "could not write data")

	} else {
		return err
	}

	return nil
}

func getDataMap(r *bufio.Reader) (dataMap map[string][]byte, err error) {

	dataMap = make(map[string][]byte)

	// read the entry count
	entryCount := new(uint32)
	err = binary.Read(r, binary.LittleEndian, entryCount)
	if err == io.EOF {
		return dataMap, err
	}

	for index := uint32(0); index < *entryCount; index++ {

		// read key
		keyLength := new(uint32)
		err = binary.Read(r, binary.LittleEndian, keyLength)
		util.CheckError(err, "could not read key length")
		keyLengthInt := uint32(*keyLength)

		keyNameBytes := make([]byte, keyLengthInt)
		err = binary.Read(r, binary.LittleEndian, keyNameBytes)
		util.CheckError(err, "could not read key")

		keyName := string(keyNameBytes)

		// read value
		valueLength := new(uint32)
		err = binary.Read(r, binary.LittleEndian, valueLength)
		util.CheckError(err, "could not read value length")
		valueLengthInt := uint32(*valueLength)

		valueBytes := make([]byte, valueLengthInt)
		err = binary.Read(r, binary.LittleEndian, valueBytes)
		util.CheckError(err, "could not read value")

		dataMap[keyName] = valueBytes
	}

	return dataMap, err
}

func setDataMap(writer *bufio.Writer, dataMap map[string][]byte) (err error) {

	entryCount := len(dataMap)
	if entryCount > 0 {
		err = binary.Write(writer, binary.LittleEndian, uint32(entryCount))
	}

	for k, v := range dataMap {

		kLen := len(k)
		err = binary.Write(writer, binary.LittleEndian, uint32(kLen))
		util.CheckError(err, "could not write key length")

		err = binary.Write(writer, binary.LittleEndian, []byte(k))
		util.CheckError(err, "could not write key")

		vLen := len(v)
		err = binary.Write(writer, binary.LittleEndian, uint32(vLen))
		util.CheckError(err, "could not write value length")

		err = binary.Write(writer, binary.LittleEndian, []byte(v))
		util.CheckError(err, "could not write value")

		err = writer.Flush()
		util.CheckError(err, "could not flush writes")
	}

	return err
}