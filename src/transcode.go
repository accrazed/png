package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

type Transcoder struct {
	Width, Height uint32
	BitDepth      BitDepth
	ColorType     ColorType
	Interlace     InterlaceMethod // whether Adam7 Interlacer is used

	// seenset keeps track of which index in data the ChunkTypes begin
	// no support for multiple chunks of the same type yet
	SeenSet    map[ChunkType][]uint32
	DataChunks []*Image
	DataState  DataState

	Filterer   *AdaptiveFilter
	compressor Compresser
}
type Image struct {
	DataState DataState
	data      []byte
	Scanlines []Scanline
}

type Scanline struct {
	pos       uint
	numPixels uint
	filter    FilterType
}

func (t *Transcoder) String() string {
	return fmt.Sprintf("img W/H %vx%v BD/CT %v/%v SeenSet %v",
		t.Width, t.Height, t.BitDepth, t.ColorType, t.SeenSet)
}

func NewTranscoder(file io.Reader) (*Transcoder, error) {
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}

	if !bytes.Equal(b[:8], PNG_SIGNATURE) {
		return nil, fmt.Errorf("invalid png header")
	}

	//? todo: consider the very chaotic idea of making chunk processing
	//? concurrent for the memes hehehe
	t := &Transcoder{
		SeenSet:   make(map[ChunkType][]uint32),
		DataState: DataStateCompressed,
	}

	// Process raw data
	var pos uint32 = 8
	for {
		if _, ok := t.SeenSet["IEND"]; ok {
			break
		}

		// file length
		loc := pos
		len := binary.BigEndian.Uint32(b[pos : pos+4])

		pos += 4
		typ := ChunkType(string(b[pos : pos+4]))

		pos += 4
		chunk := b[pos : pos+len]

		pos += len
		crc := b[pos : pos+4]

		pos += 4

		err := t.initChunk(loc, typ, chunk, crc)
		if err != nil {
			return nil, err
		}
	}

	return t, err
}

func (t *Transcoder) initChunk(loc uint32, chunkType ChunkType, rawData, crc []byte) error {
	if crc32.ChecksumIEEE(append([]byte(chunkType), rawData...)) != binary.BigEndian.Uint32(crc) {
		return fmt.Errorf("crc32 failed for chunk %s (byte %v)", string(chunkType), loc)
	}

	switch chunkType {
	case ImageHeaderType:
		t.Width = binary.BigEndian.Uint32(rawData[:4])
		t.Height = binary.BigEndian.Uint32(rawData[4:8])

		t.BitDepth = BitDepth(rawData[8])
		t.ColorType = ColorType(rawData[9])
		if err := verifyBitDepthAndColorType(t.BitDepth, t.ColorType); err != nil {
			return err
		}

		// Compression method
		switch rawData[10] {
		case 0:
			t.compressor = NewFlater()
		default:
			return fmt.Errorf("unsupported compressor type: %v", rawData[10])
		}

		// Filter Method
		switch rawData[11] {
		case 0:
			t.Filterer = &AdaptiveFilter{Width: t.Width}
		default:
			return fmt.Errorf("unsupported filter method: %v", rawData[11])
		}

		// Interlace method
		t.Interlace = InterlaceMethod(rawData[12])

	case ImageDataType:
		if _, ok := t.SeenSet[ImageHeaderType]; !ok {
			return fmt.Errorf("IDAT header declared before IHDR")
		}
		t.DataChunks = append(t.DataChunks, t.initImageData(rawData))
	case ImageEndType:
		// only seenset is updated
	default:
		fmt.Printf("WARNING: unimplemented type %s\n", string(chunkType))
	}

	if _, ok := t.SeenSet[chunkType]; !ok {
		t.SeenSet[chunkType] = make([]uint32, 0)
	}
	t.SeenSet[chunkType] = append(t.SeenSet[chunkType], loc)

	return nil
}

func verifyBitDepthAndColorType(bd BitDepth, ct ColorType) error {
	if ct == CTUnknown || bd == BDUnknown {
		return fmt.Errorf("invalid ColorType/BitDepth: %v/%v", ct, bd)
	}
	switch ct {
	case CT2, CT4, CT6:
		if bd < BD8 {
			return fmt.Errorf("invalid ColorType/BitDepth: %v/%v", ct, bd)
		}
	case CT3:
		if bd > BD8 {
			return fmt.Errorf("invalid ColorType/BitDepth: %v/%v", ct, bd)
		}
	}

	return nil
}

func (t *Transcoder) initImageData(rawData []byte) *Image {

	return &Image{
		DataState: DataStateUnfiltered,
		data:      rawData,
		Scanlines: nil,
	}
}
