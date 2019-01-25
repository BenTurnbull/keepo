package store

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

type testEntry struct {
	key string
	value string
	password string
}

func TestStoreFunctions(t *testing.T) {

	path := "."
	cleanup(path, t)

	testEntries := []testEntry{
		{"testKey1", "testValue1", "password01"},
		{"testKey2", "testValue2", "password02"}}

	fmt.Println("test setting store key/values")
	for k, v := range testEntries {
		err := SetMapValueV1(path, v.key, v.value, v.password)
		if err != nil {
			t.Errorf("could not set map value %q '%q'", k, err)
		}
	}

	retrievedKeys := GetMapKeysV1(path)
	fmt.Println("test getting store keys")
	for k, v := range testEntries {
		if !strings.EqualFold(v.key, retrievedKeys[k]) {
			t.Errorf("map key did not match '%q' it was '%q'", v.key, retrievedKeys[k])
		}
	}

	fmt.Println("test getting store values")

	for _, v := range testEntries {
		value, err := GetMapValueV1(path, v.key, v.password)
		if err != nil {
			t.Errorf("could not get map value %q '%q'", v.value, err)
		} else if !strings.EqualFold(v.value, string(value)) {
			t.Errorf("map value did not match '%q' it was '%q'", v.value, value)
		}
	}

	cleanup(path, t)
}

func cleanup(path string, t *testing.T) {
	storePath := GetStorePath(path)
	_, err := os.Stat(storePath)
	if err == nil {
		err = os.Remove(storePath)
		if err != nil {
			t.Errorf("could not delete existing store '%q'", err)
		}
	}
}