package main

import (
	"os"
	"testing"
	"time"

	"github.com/paulcacheux/xmlite/xml"
)

const TEST_PATH = "./testdata/f48bc264f9ca35fa6d482a6ffb71ba5379093364-primary.xml"

func TestSmoke(t *testing.T) {
	f, err := os.Open(TEST_PATH)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	handler := &PkgHandler{}
	decoder := xml.NewLiteDecoder(f, handler)

	start := time.Now()
	if err := decoder.Parse(); err != nil {
		t.Fatal(err)
	}

	t.Logf("elapsed: %v\n", time.Since(start))
	t.Log(handler.counter)
}

type PkgHandler struct {
	counter int
}

func (ph *PkgHandler) StartTag(name []byte) {
	if string(name) == "package" {
		ph.counter += 1
	}
}

func (ph *PkgHandler) EndTag(name []byte) {
	if string(name) == "package" {
		ph.counter -= 1
	}
}

func (ph *PkgHandler) AttrName(name []byte) {
}

func (ph *PkgHandler) AttrValue(name []byte) {
}

func (ph *PkgHandler) CharData(value []byte) {
}
