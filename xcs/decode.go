//    Copyright 2017 drillbits
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package xcs

import (
	"errors"
	"fmt"

	"golang.org/x/text/transform"
)

var errInvalidARIBXCS = errors.New("arib: invalid external character set encoding")

func newXCSDecoder() *xcsDecoder {
	d := &xcsDecoder{}
	d.init()
	return d
}

type xcsDecoder struct {
	buf         []byte
	G           [4]GraphicSet // G0, G1, G2, G3
	gl          int
	gr          int
	SS          GraphicSet
	isSmallSize bool
}

func (d *xcsDecoder) GL() GraphicSet {
	return d.G[d.gl]
}

func (d *xcsDecoder) GR() GraphicSet {
	return d.G[d.gr]
}

func (d *xcsDecoder) Reset() {
	d.init()
}

func (d *xcsDecoder) init() {
	d.buf = nil
	d.G = [4]GraphicSet{
		KanjiSet,
		AlphanumericSet,
		HiraganaSet,
		KatakanaSet, // TODO: Macro?
	}
	d.gl = 0
	d.gr = 2
}

func (d *xcsDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	d.buf = src
	size := 0
loop:
	for ; nSrc < len(d.buf); nSrc += size {
		var buf []byte
		switch b := d.buf[nSrc]; {
		case b <= SP:
			buf, size, err = d.readC0(nSrc)

		case b < DEL:
			buf, size, err = d.readGL(nSrc)

		case b <= CC1000:
			buf, size, err = d.readC1(nSrc)

		case b != CC1515:
			buf, size, err = d.readGR(nSrc)

		default:
			err = errInvalidARIBXCS
		}
		if err != nil {
			break loop
		}

		if nDst+len(buf) > len(dst) {
			err = transform.ErrShortDst
			break loop
		}
		nDst += copy(dst[nDst:], buf)
	}
	if atEOF && err == transform.ErrShortSrc {
		err = errInvalidARIBXCS
	}
	return nDst, nSrc, err
}

func (d *xcsDecoder) readC0(pos int) ([]byte, int, error) {
	return d.readControlSet(pos)
}

func (d *xcsDecoder) readC1(pos int) ([]byte, int, error) {
	return d.readControlSet(pos)
}

func (d *xcsDecoder) readControlSet(pos int) ([]byte, int, error) {
	buf, size, err := []byte{}, 1, error(nil)

	switch d.buf[pos] {
	case LS1:
		d.gl = 1
	case LS0:
		d.gl = 0
	case PAPF:
		// skip with parameter
		size++
	case SS2:
		d.SS = d.G[2]
	case ESC:
		buf, size, err = d.readESC(pos)
	case APS:
		// skip with parameter
		size += 2
	case SS3:
		d.SS = d.G[3]
	case SP:
		if d.isSmallSize {
			buf = []byte(" ")
		}
		buf = []byte("　")
	case SSZ, MSZ:
		d.isSmallSize = true
	case NSZ:
		d.isSmallSize = false
	case SZX:
		p1 := d.paramOrNil(pos, 1)
		size++
		d.isSmallSize = p1 == 0x60
	case COL:
		// skip with parameter
		p1 := d.paramOrNil(pos, 1)
		size++
		if p1 == 0x20 {
			// more parameter
			size++
		}
	case FLC, POL, WMM, HLC, RPC:
		// skip with parameter
		size++
	case MACRO:
		// TODO: test
		for i := pos + 1; d.paramOrNil(pos, i) != 0x4F; i++ {
			size++
		}
	case CSI:
		// TODO
		// return d.readCSI(src, n)
	case TIME:
		// TODO: test
		p1 := d.paramOrNil(pos, 1)
		size++
		if p1 == 0x20 {
			// skip with parameter
			size++
		} else {
			for i := pos + 2; ; i++ {
				p := d.paramOrNil(pos, i)
				if p < 0x40 && p > 0x43 {
					size++
				} else {
					break
				}
			}
		}
	default:
		// skip
	}

	return buf, size, err
}

func (d *xcsDecoder) readESC(pos int) ([]byte, int, error) {
	size, err := 1, error(nil)

	p1 := d.paramOrNil(pos, 1)
	size++
	switch {
	case p1 == 0x24 || (p1 >= 0x28 && p1 <= 0x2B):
		var gi int
		var gs byte
		size, gi, gs = d.designateGraphicSet(pos)
		d.G[gi] = graphicSets[gs]
	case p1 == 0x6E:
		d.gl = 2
	case p1 == 0x6F:
		d.gl = 3
	case p1 == 0x7E:
		d.gr = 1
	case p1 == 0x7D:
		d.gr = 2
	case p1 == 0x7C:
		d.gr = 3
	default:
		size = 1
		err = fmt.Errorf("arib: ESC has invalid parameter 0x%02X", p1)
	}

	return nil, size, err
}

// designateGraphicSet reports size of code with parameters, the index of code
// element to be designated and the key of graphic set.
func (d *xcsDecoder) designateGraphicSet(pos int) (size, gi int, gs byte) {
	size = 1 // ESC

	p1 := d.paramOrNil(pos, 1)
	size++
	switch p1 {
	case 0x28, 0x29, 0x2A, 0x2B:
		// 1 byte charset
		gi = int(p1 - 0x28)

		p2 := d.paramOrNil(pos, 2)
		size++
		switch p2 {
		case 0x20:
			// DRCS
			gs = d.paramOrNil(pos, 3)
			size++
		default:
			// G set
			gs = p2
		}
	case 0x24:
		// 2 byte charset
		p2 := d.paramOrNil(pos, 2)
		size++
		switch p2 {
		case 0x28, 0x29, 0x2A, 0x2B:
			gi = int(p2 - 0x28)

			p3 := d.paramOrNil(pos, 3)
			size++
			switch p3 {
			case 0x20:
				// DRCS
				p4 := d.paramOrNil(pos, 4)
				size++
				gs = p4
			default:
				// G set
				if p2 == 0x28 {
					gs = p2
					size--
				} else {
					gs = p3
				}
			}
		default:
			// G set
			gi = 0
			gs = p2
		}
	}

	return
}

func (d *xcsDecoder) readGL(n int) ([]byte, int, error) {
	// TODO: kanji, gaiji etc
	// SingleShift
	var gs GraphicSet
	if d.SS != nil {
		gs = d.SS
		d.SS = nil
	} else {
		gs = d.GL()
	}
	b, size := gs.Get(d.byteOrNil(n), d.byteOrNil(n+1))
	return b, size, nil
}

func (d *xcsDecoder) readGR(n int) ([]byte, int, error) {
	// TODO: kanji, gaiji etc
	b, size := d.GR().Get(d.byteOrNil(n)&0x7F, d.byteOrNil(n+1)&0x7F)
	return b, size, nil
}

func (d *xcsDecoder) byteOrNil(pos int) byte {
	if pos < len(d.buf) {
		return d.buf[pos]
	}
	return 0x00
}

func (d *xcsDecoder) paramOrNil(pos, n int) byte {
	if pos+n < len(d.buf) {
		return d.buf[pos+n]
	}
	return 0x00
}
