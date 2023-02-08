package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
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
	decoder := NewLiteDecoder(f, handler)

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

type Handler interface {
	Name(name []byte)
	AttrName(name []byte)
	AttrValue(value []byte)
	CharData(value []byte)
}

type LiteDecoder struct {
	reader    *bufio.Reader
	handler   Handler
	peekStore int
	buff      bytes.Buffer
}

func NewLiteDecoder(reader io.Reader, handler Handler) *LiteDecoder {
	return &LiteDecoder{
		reader:    bufio.NewReader(reader),
		peekStore: -1,
		handler:   handler,
	}
}

func (lt *LiteDecoder) getc() (byte, error) {
	if lt.peekStore >= 0 {
		val := lt.peekStore
		lt.peekStore = -1
		return byte(val), nil
	}

	return lt.reader.ReadByte()
}

func (lt *LiteDecoder) peekc() (byte, error) {
	if lt.peekStore < 0 {
		b, err := lt.reader.ReadByte()
		if err != nil {
			return b, err
		}
		lt.peekStore = int(b)
	}
	return byte(lt.peekStore), nil
}

func (lt *LiteDecoder) clearPeek() {
	lt.peekStore = -1
}

func (lt *LiteDecoder) Parse() error {
	for {
		if err := lt.NextToken(); err != nil {
			return err
		}
	}
}

func (lt *LiteDecoder) NextToken() error {
	for {
		curr, err := lt.peekc()
		if err != nil {
			return err
		}

		if curr != '<' {
			cd, err := lt.charData()
			if err != nil {
				return err
			}
			lt.handler.CharData(cd)
			return nil
		}

		// handle tag
		lt.clearPeek()

		// possibly </ or <?
		curr1, err := lt.peekc()
		if err != nil {
			return err
		}

		if curr1 == '/' {
			lt.clearPeek()
		} else if curr1 == '?' {
			lt.skipUntil('>')
			continue
		}

		// handle name
		name, err := lt.name()
		if err != nil {
			return err
		}
		lt.handler.Name(name)

		lt.space()

		// handle attributes
		for !lt.isNextSlashOrRightOrErr() {
			attrName, err := lt.name()
			if err != nil {
				return err
			}
			if err := lt.eat('='); err != nil {
				return err
			}
			lt.handler.AttrName(attrName)

			attrValue, err := lt.quote()
			if err != nil {
				return err
			}
			lt.handler.AttrValue(attrValue)

			lt.space()
		}

		// end tag
		lt.skipUntil('>')
		return nil
	}
}

func (lt *LiteDecoder) charData() ([]byte, error) {
	lt.buff.Reset()
	for {
		curr, err := lt.peekc()
		if err != nil {
			return nil, err
		}

		if curr == '<' {
			return lt.buff.Bytes(), nil
		}

		lt.clearPeek()
		lt.buff.WriteByte(curr)
	}
}

func (lt *LiteDecoder) skipUntil(target byte) error {
	for {
		curr, err := lt.getc()
		if err != nil {
			return err
		}

		if curr == target {
			return nil
		}
	}
}

func (lt *LiteDecoder) isNextSlashOrRightOrErr() bool {
	next, err := lt.peekc()
	return err != nil || next == '/' || next == '>'
}

func (lt *LiteDecoder) eat(expected byte) error {
	next, err := lt.getc()
	if err != nil {
		return err
	}
	if next != expected {
		return fmt.Errorf("expected `%c` and found `%c`", expected, next)
	}
	return nil
}

func (lt *LiteDecoder) name() ([]byte, error) {
	lt.buff.Reset()
	for {
		curr, err := lt.peekc()
		if err != nil {
			return nil, err
		}

		if isNameChar(curr) {
			lt.clearPeek()
			lt.buff.WriteByte(curr)
		} else {
			return lt.buff.Bytes(), nil
		}
	}
}

func (lt *LiteDecoder) quote() ([]byte, error) {
	delim, err := lt.getc()
	if err != nil {
		return nil, err
	}

	if delim != '\'' && delim != '"' {
		return nil, fmt.Errorf("expected quote delimiter")
	}

	lt.buff.Reset()
	for {
		curr, err := lt.getc()
		if err != nil {
			return nil, err
		}

		if curr != delim {
			lt.buff.WriteByte(curr)
		} else {
			return lt.buff.Bytes(), nil
		}
	}
}

func (lt *LiteDecoder) space() error {
	for {
		curr, err := lt.peekc()
		if err != nil {
			return err
		}

		if curr == ' ' || curr == '\t' || curr == '\n' || curr == '\r' {
			lt.clearPeek()
			continue
		}

		return nil
	}
}

func isNameChar(c byte) bool {
	return c == ':' || c == '-' || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}
