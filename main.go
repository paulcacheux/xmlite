package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

	reader := bufio.NewReaderSize(f, 4096*4096)
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
