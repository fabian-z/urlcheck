package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"strings"

	"golang.org/x/text/encoding"

	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// NewDecoder returns a *mime.WordDecoder wit support for additional charsets
func NewDecoder() *mime.WordDecoder {
	decoder := new(mime.WordDecoder)
	// attach custom charset decoder
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		var dec *encoding.Decoder
		// get proper charset decoder, fast path
		switch charset {
		case "gb2312":
			// gbk is an extension of gb2312
			// but not handled by ianaindex
			dec = simplifiedchinese.GBK.NewDecoder()
		default:
			// Try to resolve via ianaindex
			// matching is case-insensitive
			enc, err := ianaindex.MIME.Encoding(charset)
			if enc != nil && err == nil {
				dec = enc.NewDecoder()
			} else {
				return nil, fmt.Errorf("unhandled charset %q", charset)
			}

		}
		// read input data
		content, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}
		// decode
		data, err := dec.Bytes(content)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(data), nil
	}
	return decoder
}

// WordDecode automatically decodes word if necessary and returns decoded data
func WordDecode(word string) (string, error) {
	// return word unchanged if not RFC 2047 encoded
	if !strings.HasPrefix(word, "=?") || !strings.HasSuffix(word, "?=") || strings.Count(word, "?") != 4 {
		return word, nil
	}
	// decode and return word
	decoder := NewDecoder()
	return decoder.Decode(word)
}

// StringDecode splits text to list of words, decodes
// every word and assembles it back into a decoded string
func StringDecode(text string) (string, error) {
	words := strings.Split(text, " ")
	var err error
	for idx := range words {
		words[idx], err = WordDecode(words[idx])
		if err != nil {
			return "", err
		}
	}
	return strings.Join(words, ""), nil
}
