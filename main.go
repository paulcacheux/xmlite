package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/DataDog/nikos/rpm/dnfv2/repo"
	"github.com/DataDog/nikos/rpm/dnfv2/types"
	xmlparser "github.com/paulcacheux/xml-stream-parser"
)

const TEST_PATH = "./testdata/f48bc264f9ca35fa6d482a6ffb71ba5379093364-primary.xml"
const PPROF_PATH = "./cpu.pprof"

func main() {
	f, err := os.Open(TEST_PATH)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	prof, err := os.Create(PPROF_PATH)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer prof.Close()
	if err := pprof.StartCPUProfile(prof); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	parser := xmlparser.NewXMLParser(bufio.NewReaderSize(f, 65536), "package").SkipElements([]string{"rpm:requires"}).ParseAttributesOnly("location", "checksum", "rpm:entry")

	start := time.Now()
	for pkg := range parser.Stream() {
		if pkg.Err != nil {
			panic(err)
		}

		arch := safeQueryChild(pkg, "arch").InnerText
		locationElement := safeQueryChild(pkg, "location")
		location := safeQueryAttr(locationElement, "href")
		checksumElement := safeQueryChild(pkg, "checksum")
		var checksum *types.Checksum
		if checksumElement != nil {
			checksum = &types.Checksum{
				Hash: checksumElement.InnerText,
				Type: safeQueryAttr(checksumElement, "type"),
			}
		}

		format := safeQueryChild(pkg, "format")
		for _, provided := range format.Childs {
			if provided.Name != "rpm:provides" {
				continue
			}

			for _, entry := range provided.Element.Childs {
				if entry.Name != "rpm:entry" {
					continue
				}

				name := safeQueryAttr(&entry.Element, "name")
				if strings.Contains(name, "(") {
					continue
				}

				version := types.Version{
					Epoch: safeQueryAttr(&entry.Element, "epoch"),
					Ver:   safeQueryAttr(&entry.Element, "ver"),
					Rel:   safeQueryAttr(&entry.Element, "rel"),
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

func safeQueryChild(elem *xmlparser.XMLElement, childName string) *xmlparser.XMLElement {
	for _, child := range elem.Childs {
		if child.Name == childName {
			return &child.Element
		}
	}
	return nil
}

func safeQueryAttr(elem *xmlparser.XMLElement, attrName string) string {
	for _, attr := range elem.Attrs {
		if attr.Name == attrName {
			return attr.Value
		}
	}
	return ""
}
