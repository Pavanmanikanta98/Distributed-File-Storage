package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPathTransform(t *testing.T) {
	key := "momsMagic"
	pathkey := CASPathTransformFunc(key)
	expectedPathName := "e859f/351e1/a12eb/03d6b/424d9/e93d1/2b099/d7d7d"
	expectedOriginalKey := "e859f351e1a12eb03d6b424d9e93d12b099d7d7d"

	if pathkey.Pathname != expectedPathName {
		t.Error("Expected path:", expectedPathName, "got path:", pathkey.Pathname)
	}
	if pathkey.Filename != expectedOriginalKey {
		t.Error("Expected original:", expectedOriginalKey, "got original:", pathkey.Filename)
	}
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		pathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	data := bytes.NewReader([]byte("Hii THere"))
	if err := s.writeStream("samplefile", data); err != nil {
		t.Fatalf("Failed to write stream: %v", err)
	}

	// Clean up the created files
	defer func() {
		pathKey := s.pathTransformFunc("samplefile")
		os.RemoveAll(pathKey.Pathname) // Remove the directory and its contents
	}()
}

func TestStoreDeletekey(t *testing.T) {
	opts := StoreOpts{
		pathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)

	key := "Hii THere"

	data := []byte("Test with empty key")

	if err := s.writeStream(key, bytes.NewBuffer(data)); err != nil {
		t.Fatalf("Failed to write stream with empty key: %v", err)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)

	}

}

func readData(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// for testing purposes
// go test -v -timeout 30s -run ^TestStoreEmptyKey$
func TestStoreEmptyKey(t *testing.T) {
	opts := StoreOpts{
		pathTransformFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "Hii THere"

	data := []byte("Test with empty key")

	if err := s.writeStream(key, bytes.NewBuffer(data)); err != nil {
		t.Fatalf("Failed to write stream with empty key: %v", err)
	}

	if ok := s.Has(key); !ok {
		t.Errorf("excepted to have key %s", key)
	}

	r, er := s.Read(key)

	if er != nil {
		t.Error(er)
	}

	b, _ := readData(r)

	// fmt.Println(string(b))

	if string(b) != string(data) {
		t.Errorf("want %s have %s ", data, b)
	}

	s.Delete(key)

	// Clean up
	defer func() {
		pathKey := s.pathTransformFunc("")
		os.RemoveAll(pathKey.Pathname)
	}()
}
