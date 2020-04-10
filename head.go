package videohead

import (
	"encoding/binary"
	"errors"
	"image"
	"io"
	"os"
	"unsafe"
)

type Head struct {
	Duration int64
	Size     image.Point
}

type MP4 struct {
	r    io.ReadSeeker
	MOOV moov
}

type fixed uint32

func (f fixed) I() fixed {
	return (f & 0xFFFF0000) >> 16
}

type atom struct {
	Size uint32
	Type [4]byte
}

type moov struct {
	MVHD *mvhd
	TRAK []*trak
}

type trak struct {
	TKHD *tkhd
}

type mvhd struct {
	Pad       [12]byte
	TimeScale uint32
	Duration  uint32
}

type tkhd struct {
	Pad      [12]byte
	TrackID  uint32
	Reserved [4]byte
	Duration uint32
	Skip     [52]byte
	Width    fixed
	Height   fixed
}

func (p *MP4) Parse() (*Head, error) {
	for {
		a, err := p.readAtom()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if a.Type == [4]byte{'m', 'o', 'o', 'v'} {
			if err := p.parseMoov(int(a.Size - 8)); err != nil {
				return nil, err
			}
		} else {
			if _, err := p.skipAtom(a); err != nil {
				return nil, err
			}
		}
	}

	if p.MOOV.MVHD != nil && len(p.MOOV.TRAK) >= 1 && p.MOOV.TRAK[0].TKHD != nil {
		return &Head{
			Duration: int64(p.MOOV.MVHD.Duration) * 1000000000 / int64(p.MOOV.MVHD.TimeScale),
			Size:     image.Point{int(p.MOOV.TRAK[0].TKHD.Width.I()), int(p.MOOV.TRAK[0].TKHD.Height.I())},
		}, nil
	}

	return nil, errors.New("format error")
}

func (p *MP4) parseMoov(stop int) error {
	for cur := 0; cur < stop; {
		a, err := p.readAtom()

		if err != nil {
			return err
		}

		if a.Type == [4]byte{'m', 'v', 'h', 'd'} {
			h := mvhd{}
			if err := binary.Read(p.r, binary.BigEndian, &h); err != nil {
				return err
			}
			if _, err := p.r.Seek(int64(a.Size)-8-int64(unsafe.Sizeof(h)), os.SEEK_CUR); err != nil {
				return err
			}
			p.MOOV.MVHD = &h
		} else if a.Type == [4]byte{'t', 'r', 'a', 'k'} {
			c, err := p.parseTrak(int(a.Size) - 8)

			if err != nil {
				return err
			}

			p.MOOV.TRAK = append(p.MOOV.TRAK, c)
		} else {
			if _, err := p.skipAtom(a); err != nil {
				return err
			}
		}

		cur += int(a.Size)
	}

	return nil
}

func (p *MP4) parseTrak(stop int) (*trak, error) {
	trak := trak{}

	for cur := 0; cur < stop; {
		a, err := p.readAtom()

		if err != nil {
			return nil, err
		}

		switch a.Type {
		case [4]byte{'t', 'k', 'h', 'd'}:
			e := tkhd{}
			if err := binary.Read(p.r, binary.BigEndian, &e); err != nil {
				return nil, err
			}
			if _, err := p.r.Seek(int64(a.Size)-8-int64(unsafe.Sizeof(e)), os.SEEK_CUR); err != nil {
				return nil, err
			}
			trak.TKHD = &e
		default:
			if _, err := p.skipAtom(a); err != nil {
				return nil, err
			}
		}

		cur += int(a.Size)
	}

	return &trak, nil
}

func (p *MP4) readAtom() (atom, error) {
	a := atom{}
	err := binary.Read(p.r, binary.BigEndian, &a)
	if err != nil {
		return a, err
	}
	return a, nil
}

func (p *MP4) skipAtom(a atom) (int64, error) {
	return p.r.Seek(int64(a.Size-8), os.SEEK_CUR)
}

func Decode(rd io.ReadSeeker) (*Head, error) {
	p := &MP4{rd, moov{}}
	return p.Parse()
}
