package puppet

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
)

type file struct {
	ContentType string
	Base        string
	Data        []byte
}

func toFiles(src io.Reader) (files []*file, err error) {
	read := bufio.NewReader(src)
	tp := textproto.NewReader(read)

	hdr, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	_, params, err := mime.ParseMediaType(hdr.Get("Content-Type"))

	if err != nil {
		return nil, err
	}

	boundary := []byte("--" + params["boundary"])
	var lines []byte
	for {
		line, _, err := read.ReadLine()
		if err == io.EOF {
			return files, nil
		}
		if err != nil {
			return nil, err
		}
		if len(line) == 0 {
			lines = append(lines, '\n')
			continue
		}

		if !bytes.Equal(line, boundary) {
			lines = append(lines, line...)
			lines = append(lines, '\n')
			continue
		}
		if len(lines) == 0 {
			continue
		}

		par := bufio.NewReader(bytes.NewBuffer(lines))

		tp := textproto.NewReader(par)

		hdr, err := tp.ReadMIMEHeader()
		if err != nil {
			return nil, err
		}
		contentLocation := hdr.Get("Content-Location")

		data, err := ioutil.ReadAll(par)
		if err != nil {
			return nil, err
		}

		switch hdr.Get("Content-Transfer-Encoding") {
		case "base64":
			buf := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
			n, err := base64.StdEncoding.Decode(buf, data)
			if err != nil {
				return nil, err
			}
			data = buf[:n]
		case "quoted-printable":
			read := quotedprintable.NewReader(bytes.NewBuffer(data))
			buf, err := ioutil.ReadAll(read)
			if err != nil {
				return nil, err
			}
			data = buf
		}

		contentType := hdr.Get("Content-Type")
		file := &file{contentType, contentLocation, data}

		files = append(files, file)
		lines = lines[:0]
	}
	return files, nil
}
