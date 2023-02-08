package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
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
	for i := 0; i < math.MaxInt; i++ {
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

		lt.space()

		for !lt.isNextSlashOrRightOrErr() {
			attrName, err := lt.name()
			if err != nil {
				return "", err
			}

			if err := lt.eat('='); err != nil {
				return "", err
			}

			attrValue, err := lt.quote()
			if err != nil {
				return "", err
			}
			// fmt.Println(attrName, attrValue)
			_, _ = attrName, attrValue

			lt.space()
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

func (lt *LiteDecoder) quote() (string, error) {
	delim, err := lt.getc()
	if err != nil {
		return "", err
	}

	if delim != '\'' && delim != '"' {
		return "", fmt.Errorf("expected quote delimiter")
	}

	var buff strings.Builder
	for {
		curr, err := lt.getc()
		if err != nil {
			return "", err
		}

		if curr != delim {
			buff.WriteByte(curr)
		} else {
			return buff.String(), nil
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
