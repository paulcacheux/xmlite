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

	handler := &PkgHandler{
		tb: t,
	}
	decoder := xml.NewLiteDecoder(f, handler)

	start := time.Now()
	if err := decoder.Parse(); err != nil {
		t.Fatal(err)
	}

	t.Logf("elapsed: %v\n", time.Since(start))
}

type PkgHandler struct {
	tb    testing.TB
	stack []string
}

func (ph *PkgHandler) StartTag(name []byte) {
	ph.stack = append(ph.stack, string(name))
}

func (ph *PkgHandler) EndTag(name []byte) {
	last, newStack := ph.stack[len(ph.stack)-1], ph.stack[:len(ph.stack)-1]
	ph.stack = newStack
	if last != string(name) {
		ph.tb.Fatalf("mismatch: `%s` != `%s`", last, string(name))
	}
}

func (ph *PkgHandler) Attr(name, value []byte) {
}

func (ph *PkgHandler) CharData(value []byte) {
}
