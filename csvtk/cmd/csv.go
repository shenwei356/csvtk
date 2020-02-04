// Copyright © 2016-2019 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/shenwei356/xopen"
)

// CSVRecordsChunk is chunk of CSV records
type CSVRecordsChunk struct {
	ID   uint64
	Data [][]string
	Err  error
}

// CSVReader is
type CSVReader struct {
	Reader *csv.Reader

	bufferSize int
	chunkSize  int
	Ch         chan CSVRecordsChunk
	MetaLine   []byte // meta line of separator declaration used by MS Excel

	IgnoreEmptyRow   bool
	IgnoreIllegalRow bool

	NumEmptyRows   []int
	NumIllegalRows []int

	fh *xopen.Reader
}

// NewCSVReader is
func NewCSVReader(file string, bufferSize int, chunkSize int) (*CSVReader, error) {
	if bufferSize < 1 {
		return nil, fmt.Errorf("value of bufferSize should be greater than 0")
	}
	if chunkSize < 1 {
		return nil, fmt.Errorf("value of chunkSize should be greater than 0")
	}

	fh, err := xopen.Ropen(file)
	if err != nil {
		return nil, err
	}

	var metaLine []byte

	// var line []byte
	// line, _, err = fh.ReadLine()
	// if err != nil {
	// 	return nil, err
	// }
	//
	// if len(line) >= 5 && bytes.Equal(line[0:4], []byte("sep=")) {
	// 	metaLine = line
	// } else {
	// 	// put it back.
	// 	// but how?
	// }

	reader := csv.NewReader(fh)

	ch := make(chan CSVRecordsChunk, bufferSize)

	csvReader := &CSVReader{
		Reader:         reader,
		bufferSize:     bufferSize,
		chunkSize:      chunkSize,
		Ch:             ch,
		fh:             fh,
		MetaLine:       metaLine,
		NumEmptyRows:   make([]int, 0, 100),
		NumIllegalRows: make([]int, 0, 100),
	}
	return csvReader, nil
}

// Run begins to read
func (csvReader *CSVReader) Run() {
	go func() {
		defer func() {
			csvReader.fh.Close()
		}()

		chunkData := make([][]string, csvReader.chunkSize)
		var id uint64
		var i int
		var notBlank bool
		var data string
		var lineNum int
		for {
			record, err := csvReader.Reader.Read()
			if err == io.EOF {
				id++
				csvReader.Ch <- CSVRecordsChunk{id, chunkData[0:i], nil}
				break
			}
			lineNum++
			if err != nil {
				if csvReader.IgnoreIllegalRow {
					csvReader.NumIllegalRows = append(csvReader.NumIllegalRows, lineNum)
					continue
				} else {
					csvReader.Ch <- CSVRecordsChunk{id, chunkData[0:i], err}
					break
				}
			}
			if record == nil {
				continue
			}
			if csvReader.IgnoreEmptyRow {
				notBlank = false
				for _, data = range record {
					if data != "" {
						notBlank = true
						break
					}
				}
				if !notBlank {
					csvReader.NumEmptyRows = append(csvReader.NumEmptyRows, lineNum)
					continue
				}
			}
			chunkData[i] = record
			i++
			if i == csvReader.chunkSize {
				id++
				csvReader.Ch <- CSVRecordsChunk{id, chunkData, nil}

				chunkData = make([][]string, csvReader.chunkSize)
				i = 0
			}
		}
		close(csvReader.Ch)
	}()
}
