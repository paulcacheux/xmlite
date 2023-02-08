package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const TEST_PATH = "./testdata/f48bc264f9ca35fa6d482a6ffb71ba5379093364-primary.xml"

func main() {
	f, err := os.Open(TEST_PATH)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := NewLiteDecoder(f)
	names := make(map[string]bool)

	start := time.Now()
	for {
		name, err := decoder.Next()
		if err != nil {
			fmt.Println(err)
			break
		}
		if name == "" {
			panic("WTF")
		}
		names[name] = true
	}

	fmt.Printf("elapsed: %v\n", time.Since(start))
	fmt.Println(names)
}

type LiteDecoder struct {
	reader    *bufio.Reader
	peekStore int
}

func NewLiteDecoder(reader io.Reader) *LiteDecoder {
	return &LiteDecoder{
		reader:    bufio.NewReader(reader),
		peekStore: -1,
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

func (lt *LiteDecoder) Next() (string, error) {
	for {
		curr, err := lt.getc()
		if err != nil {
			return "", err
		}

		if curr != '<' {
			continue
		}

		curr1, err := lt.peekc()
		if err != nil {
			return "", err
		}

		if curr1 == '/' {
			lt.clearPeek()
		} else if curr1 == '?' {
			lt.skipUntil('>')
			continue
		}

		name, err := lt.name()
		if err != nil {
			return "", err
		}

		lt.skipUntil('>')
		return name, nil
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

func (lt *LiteDecoder) name() (string, error) {
	var buff strings.Builder
	for {
		curr, err := lt.peekc()
		if err != nil {
			return "", err
		}

		if isNameChar(curr) {
			lt.clearPeek()
			buff.WriteByte(curr)
		} else {
			return buff.String(), nil
		}
	}
}

func isNameChar(c byte) bool {
	return c == ':' || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}
