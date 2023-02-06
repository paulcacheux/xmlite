package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/DataDog/nikos/rpm/dnfv2/repo"
	"github.com/DataDog/nikos/rpm/dnfv2/types"
	xmlparser "github.com/paulcacheux/xml-stream-parser"
)

const TEST_PATH = "./testdata/f48bc264f9ca35fa6d482a6ffb71ba5379093364-primary.xml"

func main() {
	f, err := os.Open(TEST_PATH)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	parser := xmlparser.NewXMLParser(bufio.NewReaderSize(f, 65536), "package").SkipElements([]string{"rpm:requires"}).ParseAttributesOnly("location", "checksum", "rpm:entry")

	start := time.Now()
	for pkg := range parser.Stream() {
		if pkg.Err != nil {
			panic(err)
		}

		arch := safeQuery(pkg, "arch").InnerText
		location := safeQuery(pkg, "location").Attrs["href"]
		checksumElement := safeQuery(pkg, "checksum")
		var checksum *types.Checksum
		if checksumElement != nil {
			checksum = &types.Checksum{
				Hash: checksumElement.InnerText,
				Type: checksumElement.Attrs["type"],
			}
		}

		format := safeQuery(pkg, "format")
		provides := format.Childs["rpm:provides"]
		for _, provided := range provides {
			entries := provided.Childs["rpm:entry"]
			for _, provided := range entries {
				name := provided.Attrs["name"]
				version := types.Version{
					Epoch: provided.Attrs["epoch"],
					Ver:   provided.Attrs["ver"],
					Rel:   provided.Attrs["rel"],
				}

				_ = &repo.PkgInfo{
					Name:     name,
					Version:  version,
					Arch:     arch,
					Location: location,
					Checksum: checksum,
				}
			}
		}
	}
	fmt.Printf("elapsed: %v\n", time.Since(start))
}

func safeQuery(elem *xmlparser.XMLElement, childName string) *xmlparser.XMLElement {
	children, ok := elem.Childs[childName]
	if !ok {
		return nil
	}

	if len(children) != 1 {
		return nil
	}

	return &children[0]
}
