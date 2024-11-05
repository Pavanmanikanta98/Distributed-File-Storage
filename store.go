package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFoldername = "glnetwork"

func CASPathTransformFunc(key string) PathKey {
	//[20]byte => []byte -> [:](convert into slice)
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
	// return strings.Join(paths, "/")

}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	Pathname string
	Filename string
}

func (p PathKey) FirstPathname() string {
	paths := strings.Split(p.Pathname, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p PathKey) FullPath() string {

	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

type StoreOpts struct {
	//Root is the folder name of the root directory, containing the folders/files of the system.
	Root              string
	PathTransformFunc PathTransformFunc
}

var DefaultPathTransformFunc = func(key string) PathKey {

	return PathKey{
		Pathname: key,
		Filename: key,
	}

}

type store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *store {

	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFoldername
	}
	return &store{
		StoreOpts: opts,
	}

}

func (s *store) Has(id string, key string) bool {

	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)

	return !errors.Is(err, os.ErrNotExist)
}

func (s *store) clear() error {

	return os.RemoveAll(s.Root)
}

func (s *store) Delete(id string, key string) error {

	pathkey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("Deleting %s from disk", pathkey.Filename)
	}()

	firstPAthnameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathkey.FirstPathname())

	return os.RemoveAll(firstPAthnameWithRoot)
}

func (s *store) Write(id string, key string, data io.Reader) (int64, error) {
	// f, err := s.writeStream(key, data)

	// if err!= nil {
	//     return err
	// }

	// defer f.Close()

	// _, err = io.Copy(f, data)

	// return err
	return s.writeStream(id, key, data)
}

func (s *store) WriteDecrypt(enckey []byte, id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)

	if err != nil {
		return 0, err
	}

	n, err := copyDecrypt(enckey, r, f)

	return int64(n), err
}

func (s *store) openFileForWriting(id string, key string) (*os.File, error) {
	pathkey := s.PathTransformFunc(key)
	pathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathkey.Pathname)

	if err := os.MkdirAll(pathWithRoot, os.ModePerm); err != nil {
		return nil, err
	}
	// fullPath := pathkey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathkey.FullPath())
	return os.Create(fullPathWithRoot)
}
func (s *store) writeStream(id string, key string, r io.Reader) (int64, error) {

	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)

}

// FIXME: Done
func (s *store) Read(id string, key string) (int64, io.Reader, error) {

	return s.readStream(id, key)

}
func (s *store) readStream(id string, key string) (int64, io.ReadCloser, error) {

	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	file, err := os.Open(fullPathWithRoot)

	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()

	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil
}
