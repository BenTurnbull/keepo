package store

import (
	"fmt"
	"keepo/src/crypto"
	"log"
	"os"
	"strings"
	"testing"
)

type testEntry struct {
	key string
	value string
}

func TestAuthenticationFailure(t *testing.T) {

	testPath := "."
	cleanup(testPath, t)

	secret := "password01"
	testEntries := []testEntry{{"testKey1", "testValue1"}}

	for k, v := range testEntries {
		err := SetMapValue(testPath, v.key, v.value, secret)
		if err != nil {
			t.Errorf("could not set map value %q '%q'", k, err)
		}
	}

	wrongSecret := "password02"
	testAuthenticationOnGet(testPath, wrongSecret, testEntries, t)
	testAuthenticationOnSet(testPath, wrongSecret, testEntries, t)
	testAuthenticationOnClear(testPath, wrongSecret, testEntries, t)
}

func testAuthenticationOnClear(testPath, secret string, testEntries []testEntry, t *testing.T) {
	fmt.Println("test clearing store value with incorrect secret")
	err := ClearMapValue(testPath, testEntries[0].key, secret)
	if err == nil {
		t.Errorf("expected authentication error")
	}
	if state, ok := err.(*State); ok && state == AuthenticationFailedState {
		log.Println(state.message)
	} else {
		t.Errorf("expected authentication error, but got '%q'", err)
	}
}

func testAuthenticationOnSet(testPath, secret string, testEntries []testEntry, t *testing.T) {
	fmt.Println("test setting store value with incorrect secret")
	err := SetMapValue(testPath, testEntries[0].key, "newValue", secret)
	if err == nil {
		t.Errorf("expected authentication error")
	}
	if state, ok := err.(*State); ok && state == AuthenticationFailedState {
		log.Println(state.message)
	} else {
		t.Errorf("expected authentication error, but got '%q'", err)
	}
}

func testAuthenticationOnGet(path, secret string, testEntries []testEntry, t *testing.T) {
	fmt.Println("test getting store value with incorrect secret")
	_, err := GetMapValue(path, testEntries[0].key, secret)
	if err == nil {
		t.Errorf("expected authentication error")
	}
	if state, ok := err.(*State); ok && state == AuthenticationFailedState {
		log.Println(state.message)
	} else {
		t.Errorf("expected authentication error, but got '%q'", err)
	}
}

func TestGetAbsentValue(t *testing.T) {

	path := "."
	cleanup(path, t)

	secret := "password01"
	testEntries := []testEntry{{"testKey1", "testValue1"}}

	fmt.Println("test getting value on empty store")
	_, err := GetMapValue(path, testEntries[0].key, secret)
	if err == nil {
		t.Errorf("expected no entry state")
	}
	if state, ok := err.(*State); ok && state == ValueAbsentState {
		log.Println(state.message)
	} else {
		t.Errorf("expected ValueAbsentState, but got '%q'", err)
	}

	fmt.Println("test clearing value on empty store")
	err = ClearMapValue(path, testEntries[0].key, secret)
	if err == nil {
		t.Errorf("expected no entry state")
	}
	if state, ok := err.(*State); ok && state == ValueAbsentState {
		log.Println(state.message)
	} else {
		t.Errorf("expected ValueAbsentState, but got '%q'", err)
	}

	// create a data store
	err = SetMapValue(path, "test", "test", secret)
	if err != nil {
		t.Errorf("could not set map value '%q'", err)
	}

	_, err = GetMapValue(path, testEntries[0].key, secret)
	if err == nil {
		t.Errorf("expected no entry state")
	}

	if state, ok := err.(*State); ok && state == ValueAbsentState {
		log.Println(state.message)
	} else {
		t.Errorf("expected ValueAbsentState, but got '%q'", err)
	}
}

func TestDataStoreRoundTrip(t *testing.T) {

	path := "."
	cleanup(path, t)

	secret := "password01"
	testData := crypto.GenerateNonce()
	testKey2 := "testKey2"
	testValue2 := string(testData[:])
	testEntries := []testEntry{
		{"testKey1", "testValue1"},
		{testKey2, testValue2},
		{"testKey3", ""}}

	fmt.Println("test setting store key/values")
	for k, v := range testEntries {
		err := SetMapValue(path, v.key, v.value, secret)
		if err != nil {
			t.Errorf("could not set map value %q '%q'", k, err)
		}
	}

	retrievedKeys := GetMapKeys(path)
	fmt.Println("test getting store keys")
	for k, v := range testEntries {
		if !strings.EqualFold(v.key, retrievedKeys[k]) {
			t.Errorf("map key did not match '%q' it was '%q'", v.key, retrievedKeys[k])
		}
	}

	fmt.Println("test getting store values")
	for _, v := range testEntries {
		value, err := GetMapValue(path, v.key, secret)
		if err != nil {
			t.Errorf("could not get map value %q '%q'", v.value, err)
		} else if !strings.EqualFold(v.value, string(value)) {
			t.Errorf("map value did not match '%q' it was '%q'", v.value, value)
		}
	}

	fmt.Println("test deletion of store value")
	err := ClearMapValue(path, testKey2, secret)
	if err != nil {
		t.Errorf("could not clear map value %q '%q'", testKey2, err)
	}

	fmt.Println("test deletion of non existent store value")
	err = ClearMapValue(path, testKey2, secret)
	if err != ValueAbsentState {
		t.Errorf("expected value to be absent but was %s", err)
	}

	retrievedKeys = GetMapKeys(path)
	keyCount := len(retrievedKeys)
	if keyCount != 2 {
		t.Errorf("expected 2 keys but found %d", keyCount)
	}

	for _, v := range retrievedKeys {
		if strings.EqualFold(v, testKey2) {
			t.Errorf("expected '%s' to be absent", testKey2)
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