package chunker

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// xmlText scans an XML stream and collects CharData inside elements whose local
// name matches textLocal. A newline is appended on each paraLocal end element.
func xmlText(r io.Reader, textLocal, paraLocal string) (string, error) {
	dec := xml.NewDecoder(r)
	var sb strings.Builder
	capture := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == textLocal {
				capture = true
			}
		case xml.EndElement:
			if t.Name.Local == textLocal {
				capture = false
				sb.WriteByte(' ')
			} else if paraLocal != "" && t.Name.Local == paraLocal {
				sb.WriteByte('\n')
			}
		case xml.CharData:
			if capture {
				sb.Write(t)
			}
		}
	}
	return strings.TrimSpace(sb.String()), nil
}

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
