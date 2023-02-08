package main

import (
	"fmt"
	"os"
	"time"

	"github.com/paulcacheux/xmlite/xml"
)

const TEST_PATH = "./testdata/f48bc264f9ca35fa6d482a6ffb71ba5379093364-primary.xml"

func main() {
	f, err := os.Open(TEST_PATH)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	handler := &DebugHandler{
		names: make(map[string]bool),
		attrs: make(map[string]bool),
	}
	decoder := xml.NewLiteDecoder(f, handler)

	start := time.Now()
	if err := decoder.Parse(); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("elapsed: %v\n", time.Since(start))
	fmt.Println(handler.names)
	fmt.Println(handler.attrs)
}

type DebugHandler struct {
	names map[string]bool
	attrs map[string]bool
}

func (dh *DebugHandler) Name(name []byte) {
	dh.names[string(name)] = true
}

func (dh *DebugHandler) AttrName(name []byte) {
	dh.attrs[string(name)] = true
}

func (dh *DebugHandler) AttrValue(name []byte) {
}

func (dh *DebugHandler) CharData(value []byte) {
}
