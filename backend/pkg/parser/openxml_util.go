package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// openXMLEntry opens a named entry in a ZIP-based Open XML archive.
func openXMLEntry(zr *zip.Reader, name string) (io.ReadCloser, error) {
	for _, f := range zr.File {
		if f.Name == name {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("entry %q not found", name)
}

// openXMLReader reads r to completion and returns a zip.Reader over its bytes.
func openXMLReader(r io.ReadCloser) (*zip.Reader, error) {
	defer r.Close()
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
}

// xmlStreamText calls yield for each paragraph found in an XML stream.
// textLocal is the element whose CharData is collected (e.g. "t");
// paraLocal is the element whose end tag flushes the current paragraph (e.g. "p").
func xmlStreamText(r io.Reader, textLocal, paraLocal string, yield func(string) bool) error {
	dec := xml.NewDecoder(r)
	var cur []byte
	capture := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == textLocal {
				capture = true
			}
		case xml.EndElement:
			if t.Name.Local == textLocal {
				capture = false
				cur = append(cur, ' ')
			} else if t.Name.Local == paraLocal {
				text := string(bytes.TrimSpace(cur))
				cur = cur[:0]
				if text != "" && !yield(text) {
					return nil
				}
			}
		case xml.CharData:
			if capture {
				cur = append(cur, t...)
			}
		}
	}
	return nil
}
