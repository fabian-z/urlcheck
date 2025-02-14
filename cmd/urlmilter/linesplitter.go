// Code in this file from https://github.com/tuck1s/go-smtpproxy, Apache License v2.0
// Package smtpproxy is based heavily on https://github.com/emersion/go-smtp, with increased transparency of response codes and no sasl dependency.
package main

import "io"

// Linesplitter is an io.Writer
// See https://www.ietf.org/rfc/rfc2045.txt, section 6.8 for notes on maximum line length of 76 characters

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type linesplitter struct {
	len   int
	count int
	sep   []byte
	w     io.Writer
}

// NewLineSplitter splits input every len bytes with a sep byte sequence, outputting to writer w
func NewLineSplitter(len int, sep []byte, w io.Writer) io.Writer {
	return &linesplitter{len: len, count: 0, sep: sep, w: w}
}

// Split a line in to ls.len chunks with separator
func (ls *linesplitter) Write(in []byte) (n int, err error) {
	writtenThisCall := 0
	readPos := 0
	// Leading chunk size is limited by: how much input there is; defined split length; and
	// any residual from last time
	chunkSize := min(len(in), ls.len-ls.count)
	// Pass on chunk(s)
	for {
		ls.w.Write(in[readPos:(readPos + chunkSize)])
		readPos += chunkSize // Skip forward ready for next chunk
		ls.count += chunkSize
		writtenThisCall += chunkSize

		// if we have completed a chunk, emit a separator
		if ls.count >= ls.len {
			ls.w.Write(ls.sep)
			writtenThisCall += len(ls.sep)
			ls.count = 0
		}
		inToGo := len(in) - readPos
		if inToGo <= 0 {
			break // reached end of input data
		}
		// Determine size of the NEXT chunk
		chunkSize = min(inToGo, ls.len)
	}
	return writtenThisCall, nil
}
