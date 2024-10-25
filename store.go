package main

import (
	"bytes"
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

func (s *store) Has(key string) bool {

	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)

	return !errors.Is(err, os.ErrNotExist)
}

func (s *store) clear() error {

	return os.RemoveAll(s.Root)
}

func (s *store) Delete(key string) error {

	pathkey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("Deleting %s from disk", pathkey.Filename)
	}()

	firstPAthnameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.FirstPathname())

	return os.RemoveAll(firstPAthnameWithRoot)
}

func (s *store) Write(key string, data io.Reader) error {
	// f, err := s.writeStream(key, data)

	// if err!= nil {
	//     return err
	// }

	// defer f.Close()

	// _, err = io.Copy(f, data)

	// return err
	return s.writeStream(key, data)
}

func (s *store) Read(key string) (io.Reader, error) {

	f, er := s.readStream(key)

	if er != nil {

		return nil, er

	}

	defer f.Close()

	buf := new(bytes.Buffer)

	_, er = io.Copy(buf, f)

	return buf, er
}
func (s *store) readStream(key string) (io.ReadCloser, error) {

	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	return os.Open(fullPathWithRoot)

}

func (s *store) writeStream(key string, r io.Reader) error {

	pathkey := s.PathTransformFunc(key)
	pathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.Pathname)

	if err := os.MkdirAll(pathWithRoot, os.ModePerm); err != nil {
		return err
	}

	// fullPath := pathkey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathkey.FullPath())

	f, err := os.Create(fullPathWithRoot)

	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)

	if err != nil {
		return err
	}
	// log.Printf("working")
	log.Printf("written %d bytes to disk: %s ", n, fullPathWithRoot)

	return nil
}
