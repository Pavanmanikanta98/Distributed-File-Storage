package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestPathTransform(t *testing.T) {
	key := "momsMagic"

	pathkey := CASPathTransformFunc(key)
	expectedPathName := "e859f/351e1/a12eb/03d6b/424d9/e93d1/2b099/d7d7d"
	expextedFilename := "e859f351e1a12eb03d6b424d9e93d12b099d7d7d"

	if pathkey.Pathname != expectedPathName {
		t.Error("Expected path:", expectedPathName, "got path:", pathkey.Pathname)
	}
	if pathkey.Filename != expextedFilename {
		t.Error("Expected original:", expextedFilename, "got original:", pathkey.Filename)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	id := generateID()
	defer tearDown(t, s)

	for i := 0; i < 50; i++ {

		key := fmt.Sprintf("goo%d", i)
		data := []byte("Test with empty key")

		if _, err := s.writeStream(id, key, bytes.NewBuffer(data)); err != nil {
			t.Fatalf("Failed to write stream with empty key: %v", err)
		}
		if ok := s.Has(id, key); !ok {
			t.Fatalf("Expected to have key %s", key)
		}

		_, r, er := s.Read(id, key)

		if er != nil {
			t.Error(er)
		}

		b, _ := readData(r)

		// fmt.Println(string(b))

		if string(b) != string(data) {
			t.Errorf("want %s have %s ", data, b)
		}

		if err := s.Delete(id, key); err != nil {
			t.Error(err)
		}

		if ok := s.Has(id, key); ok {
			t.Fatalf("Expected to NOT have key %s", key)
		}

		// Clean up the created files
		pathKey := s.PathTransformFunc(key)
		if err := os.RemoveAll(pathKey.Pathname); err != nil {
			t.Errorf("Failed to clean up: %v", err)
		}
	}

}

func newStore() *store {
	return NewStore(StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	})
}

func tearDown(t *testing.T, s *store) {

	if err := s.clear(); err != nil {
		t.Error(err)
	}
}

func readData(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// don't need any more

// func TestStoreDeletekey(t *testing.T) {
// 	opts := StoreOpts{
// 		pathTransformFunc: CASPathTransformFunc,
// 	}
// 	s := NewStore(opts)

// 	key := "Hii THere"

// 	data := []byte("Test with empty key")

// 	if err := s.writeStream(key, bytes.NewBuffer(data)); err != nil {
// 		t.Fatalf("Failed to write stream with empty key: %v", err)
// 	}

// 	if err := s.Delete(key); err != nil {
// 		t.Error(err)

// 	}

//}

// for testing purposes
// go test -v -timeout 30s -run ^TestStoreEmptyKey$
// func TestStoreEmptyKey(t *testing.T) {
// 	opts := StoreOpts{
// 		pathTransformFunc: CASPathTransformFunc,
// 	}
// 	s := NewStore(opts)
// 	key := "Hii THere"

// 	data := []byte("Test with empty key")

// 	if err := s.writeStream(key, bytes.NewBuffer(data)); err != nil {
// 		t.Fatalf("Failed to write stream with empty key: %v", err)
// 	}

// 	if ok := s.Has(key); !ok {
// 		t.Errorf("excepted to have key %s", key)
// 	}

// 	r, er := s.Read(key)

// 	if er != nil {
// 		t.Error(er)
// 	}

// 	b, _ := readData(r)

// 	// fmt.Println(string(b))

// 	if string(b) != string(data) {
// 		t.Errorf("want %s have %s ", data, b)
// 	}

// 	s.Delete(key)

// 	// Clean up
// 	defer func() {
// 		pathKey := s.pathTransformFunc("")
// 		os.RemoveAll(pathKey.Pathname)
// 	}()
// }
