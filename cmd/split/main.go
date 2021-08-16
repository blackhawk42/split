package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blackhawk42/split/pkg/splitfile"
	"github.com/inhies/go-bytesize"
)

var (
	copyBufferSize     = flag.Int("buffer-size", 32*1024, "size of all buffers used for copying")
	chunkSizeString    = flag.String("size", "100MB", "chunk max byte size, expressed in a human readable `string`")
	NotComputeChecksum = flag.Bool("no-checksum", false, "don't compute checksums")
	NotCreateSplitfile = flag.Bool("no-splitfile", false, "don't create splitfile; implies no-checksum")
	outDir             = flag.String("dir", ".", "output `directory` for all generated files, which will be created if nonexistant; defaults to current directory")
)

func main() {
	flag.Parse()

	checksumAlgorithm := splitfile.ChecksumAlgoCRC32

	if *NotCreateSplitfile {
		*NotComputeChecksum = true
	}

	if *NotComputeChecksum {
		checksumAlgorithm = splitfile.ChecksumAlgoNone
	}

	chunkSize, err := bytesize.Parse(*chunkSizeString)
	if err != nil {
		printError("while parsing chunk size", err, 1)
	}
	if chunkSize <= 0 {
		fmt.Fprintf(os.Stderr, "error: chunk size must be > 0, was %d bytes\n", chunkSize)
		os.Exit(1)
	}

	err = os.MkdirAll(*outDir, os.ModePerm)
	if err != nil {
		printError("while creating output directory", err, 1)
	}

	for _, filename := range flag.Args() {
		splitfileStruct, err := splitFile(filename, int64(chunkSize), checksumAlgorithm, *outDir)
		if err != nil {
			printError("while splitting file", err, 1)
		}

		if !*NotCreateSplitfile {
			func() {
				splitF, err := os.Create(filepath.Join(*outDir, splitfile.FormatSplitfileFilename(filename)))
				if err != nil {
					printError("while creating splitfile", err, 1)
				}
				defer splitF.Close()

				JSONEncoder := json.NewEncoder(splitF)
				JSONEncoder.SetIndent("", " ")
				err = JSONEncoder.Encode(splitfileStruct)
				if err != nil {
					printError("while encoding JSON", err, 1)
				}
			}()
		}
	}
}

func printError(origin string, err error, exitStatus int) {
	fmt.Fprintf(os.Stderr, "error: %s: %v\n", origin, err)
	os.Exit(1)
}
