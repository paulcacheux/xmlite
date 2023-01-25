package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const TEST_PATH = "./testdata/f48bc264f9ca35fa6d482a6ffb71ba5379093364-primary.xml"

type State int

const (
	CharData State = iota
	Tag
	TagArgs
)

func main() {
	f, err := os.Open(TEST_PATH)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(f)

	var state State = CharData
	var current strings.Builder

	tags := make(map[string]bool)

	for {
		b, err := reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		if b == '<' {
			state = Tag
			current.Reset()
		} else if b == '>' {
			state = CharData
		} else if state == Tag {
			if b == '/' || b == '>' || b == ' ' {
				if b == '>' {
					state = CharData
				} else {
					state = TagArgs
				}
				tags[current.String()] = true
				current.Reset()
			} else {
				current.WriteByte(b)
			}
		}
	}

	fmt.Println(tags)
}
