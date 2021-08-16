package main

import (
	"errors"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/blackhawk42/split/pkg/splitfile"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		buff := make([]byte, *copyBufferSize)
		return &buff
	},
}

func splitFile(filename string, chunkSize int64, checksumAlgo splitfile.ChecksumAlgo, outputDir string) (splitFile *splitfile.Splitfile, err error) {
	// First, open the main file and get it's size
	var f *os.File
	f, err = os.Open(filename)
	if err != nil {
		return
	}
	var fStat os.FileInfo
	fStat, err = f.Stat()
	if err != nil {
		return
	}

	// Build the skeleton for the final splitfile
	splitFile = &splitfile.Splitfile{
		ChecksumAlgo: checksumAlgo,
		MainFile: &splitfile.File{
			Filename: filename,
			Size:     fStat.Size(),
			// Only final checksum remains
		},
		Chunks: make([]*splitfile.File, 0), // 0 chuncks to start with
	}

	// The hash creator function. From here, hashCreator serves as a flag to check
	// if checksum calculation is enabled for the whole thing
	var hashCreator splitfile.HasherFunc
	hashCreator, err = splitfile.GetHasher(checksumAlgo)
	if err != nil {
		return nil, err
	}

	// The hash checksum for the whole, unchunked file, if hashing is desired at all
	var mainFileHasher hash.Hash
	if hashCreator != nil {
		mainFileHasher = hashCreator()
	}

	// The limited reader that will be getting its data from the main file.
	// The limit itself will be set during the main loop.
	reader := &io.LimitedReader{
		R: f,
	}

	// Get a buffer for CopyBuffer, and don't forget to return it when done
	copyBuffer := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(copyBuffer)

	// Main loop, and initialization of needed variables
	var totalWritten int64           // Total written bytes, used for finishing the main loop
	var currentChunk *splitfile.File // The current chunk for the current iteration
	var chunkF *os.File              // File handler for the current chunk
	var chunkHasher hash.Hash        // hasher for the current chunk, if hashing is desired at all
	var finalWriter io.Writer        // The final writer, joining together the hashers and chunk file
	for {
		// First, create skeleton of the chunk struct and give a name to its
		// filename with the appropiate format
		currentChunk = &splitfile.File{
			Filename: splitfile.FormatChunkFilename(filename, len(splitFile.Chunks)),
		}

		// Reset the limited reader
		reader.N = chunkSize
		// Brand new hasher for the current chunk, if desired
		if hashCreator != nil {
			chunkHasher = hashCreator()
		}
		func() { // Closure just to defer chunkF.Close in every iteration
			// Create the file handle for the current chunk
			chunkF, err = os.Create(filepath.Join(outputDir, currentChunk.Filename))
			if err != nil {
				return
			}
			defer chunkF.Close()

			if hashCreator != nil {
				// Create a writer that at the same time calculates the checksum of
				// the main file, the checksum of the current chunk, and writes the
				// actual data to the current chunk.
				finalWriter = io.MultiWriter(mainFileHasher, chunkHasher, chunkF)
			} else {
				// if hash calculation is disabled, just write directly to the chunk
				finalWriter = chunkF
			}

			// Do the copy. The return value conveniently gives us the amount
			// copies and, therefore, the size of the chunk.
			currentChunk.Size, err = io.CopyBuffer(finalWriter, reader, *copyBuffer)
		}()
		// If an error occured that is not EOF, break inmediately
		if err != nil && !errors.Is(err, io.EOF) {
			break
		}

		if hashCreator != nil {
			// Calculate the final checksum of the current chunk,
			// and add it to the collection.
			currentChunk.Checksum = chunkHasher.Sum(nil)
		}

		totalWritten += currentChunk.Size
		splitFile.Chunks = append(splitFile.Chunks, currentChunk)

		if totalWritten == splitFile.MainFile.Size {
			break
		}
	}
	// Any other error that's not EOF is considered a failure; return nothing and
	// the error.
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	err = nil // Reset general error to nil

	// Finally, caulculate the final checksum of the whole file, if checksums are used at all
	if hashCreator != nil {
		splitFile.MainFile.Checksum = mainFileHasher.Sum(nil)
	}

	// All named returned values should be set appropiately by this point
	return
}
