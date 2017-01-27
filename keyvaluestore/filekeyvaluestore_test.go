package keyvaluestore

import (
	"testing"
)

func TestFKVSMethods(t *testing.T) {
	stateStore, err := CreateNewFileKeyValueStore("/tmp/keyvaluestore")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	stateStore.SetValue("testvalue", []byte("data"))
	value, err := stateStore.GetValue("testvalue")
	if err != nil {
		t.Fatalf("stateStore.SetValue return error[%s]", err)
	}
	if string(value) != "data" {
		t.Fatalf("stateStore.GetValue didn't return the right value")
	}

}
