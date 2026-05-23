package png

import (
	"bytes"
	"compress/zlib"
	"io"
)

type Compresser interface {
	Compress(*Image) (*Image, error) // TODO: figure out signature
	Uncompress(*Image) (*Image, error)
}

type Flater struct{}

func NewFlater() Compresser {
	return &Flater{}
}

func (*Flater) Compress(id *Image) (*Image, error) {
	b := bytes.Buffer{}
	w := zlib.NewWriter(&b)
	_, err := w.Write(id.data)
	if err != nil {
		return nil, err
	}
	id.data = b.Bytes()

	return id, nil
}

func (*Flater) Uncompress(id *Image) (*Image, error) {
	r, err := zlib.NewReader(bytes.NewReader(id.data))
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	io.Copy(&b, r)

	id.data = b.Bytes()

	return id, nil
}

var _ Compresser = (*Flater)(nil)
