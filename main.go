package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/paulcacheux/xmlite/xml"
)

const TEST_PATH = "./testdata/f48bc264f9ca35fa6d482a6ffb71ba5379093364-primary.xml"
const CPU_PPROF_PATH = "./cpu.pprof"
const MEM_PPROF_PATH = "./mem.pprof"

func main() {
	f, err := os.Open(TEST_PATH)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// CPU profiling
	prof, err := os.Create(CPU_PPROF_PATH)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer prof.Close()
	if err := pprof.StartCPUProfile(prof); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	freader := &FilterReader{inner: f}
	reader := bufio.NewReaderSize(freader, 1024)

	decoder := xml.NewDecoder(reader)

	start := time.Now()

	for {
		_, err := decoder.RawToken()
		if err != nil {
			fmt.Println(err)
			break
		}
	}

	fmt.Printf("elapsed: %v\n", time.Since(start))

	// Heap profiling
	mem, err := os.Create(MEM_PPROF_PATH)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer mem.Close()
	runtime.GC()
	if err := pprof.WriteHeapProfile(mem); err != nil {
		panic(fmt.Errorf("could not write memory profile: %w", err))
	}
}

type FilterReader struct {
	inner io.Reader
}

var re = regexp.MustCompile(`<rpm:entry\s+name="[^"]*?\([^"]*?".*?\/>`)

func (fr *FilterReader) Read(p []byte) (int, error) {
	n, err := fr.inner.Read(p)
	if err != nil || !bytes.Contains(p, []byte("<rpm:entry")) {
		return n, err
	}

	fixedBuf := re.ReplaceAllLiteral(p, nil)
	resCount := len(fixedBuf)
	copy(p[:resCount], fixedBuf)
	return resCount, nil
}
