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
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"golang.org/x/text/transform"
)

var _ transform.Transformer = (*xcsDecoder)(nil)

func TestDecode(t *testing.T) {
	for i, tc := range []struct {
		name string
		src  []byte
		dst  []byte
	}{
		{
			name: "HiraganaOnly",
			src:  []byte{0xaa, 0xab, 0xa2, 0xb5, 0xf3, 0xc8, 0xa4, 0xc3, 0xb7, 0xe7},
			dst:  []byte("おかあさんといっしょ"),
		},
		{
			name: "KatakanaOnly",
			src:  []byte{0x1b, 0x7c, 0xd1, 0xba, 0xc9, 0xe9, 0xaf, 0xed, 0xb9},
			dst:  []byte("パズドラクロス"),
		},
		{
			name: "AdditionalSymbols",
			src:  []byte{0xaa, 0xab, 0xa2, 0xb5, 0xf3, 0xc8, 0xa4, 0xc3, 0xb7, 0xe7, 0x1b, 0x24, 0x3b, 0x7a, 0x56},
			dst:  []byte("おかあさんといっしょ【字】"),
		},
	} {
		i, tc := i, tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tr := transform.NewReader(bytes.NewReader(tc.src), XCSEncoding.NewDecoder())
			got, err := ioutil.ReadAll(tr)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tc.dst) {
				t.Errorf("%d: Decode(0x%X) => %s, want %s", i, tc.src, string(got), string(tc.dst))
			}
		})
	}
}

// TODO
func TestXCSDecoderReset(t *testing.T) {}

func TestXCSDecoderInit(t *testing.T) {
	d := &xcsDecoder{
		buf: []byte{0x47},
	}
	d.init()

	if d.buf != nil {
		t.Errorf("d.buf 0x%X, want %v", d.buf, nil)
	}
	for i, tc := range []struct {
		g GraphicSet
	}{
		{KanjiSet},
		{AlphanumericSet},
		{HiraganaSet},
		{KatakanaSet},
	} {
		i, tc := i, tc
		t.Run(fmt.Sprintf("g[%d]", i), func(t *testing.T) {
			t.Parallel()

			got := ptr(d.G[i])
			exp := ptr(tc.g)
			if got != exp {
				t.Errorf("d.g[%d] %v, want %v", i, got, exp)
			}
		})
	}
	got := ptr(d.GL())
	exp := ptr(d.G[0])
	if got != exp {
		t.Errorf("d.gl %v, want %v", got, exp)
	}
	got = ptr(d.GR())
	exp = ptr(d.G[2])
	if got != exp {
		t.Errorf("d.gr %v, want %v", got, exp)
	}
}

// TODO
func TestXCSDecoderTransform(t *testing.T) {}

func TestXCSDecoderGetParamOrNil(t *testing.T) {
	for i, tc := range []struct {
		src []byte
		pos int
		n   int
		p   byte
	}{
		{[]byte{0xAA, 0x1B, 0x28, 0x42}, 1, 1, 0x28},
		{[]byte{0xAA, 0x1B, 0x28, 0x42}, 1, 2, 0x42},
		{[]byte{0xAA, 0x1B, 0x28, 0x42}, 4, 1, 0x00},
		{[]byte{0xAA, 0x1B, 0x28, 0x42}, 1, 3, 0x00},
	} {
		d := newXCSDecoder()
		d.buf = tc.src
		b := d.paramOrNil(tc.pos, tc.n)
		if b != tc.p {
			t.Errorf("%d: buf 0x%02X, getParamOrNil(%d, %d) => 0x%02X, want 0x%02X", i, tc.src, tc.pos, tc.n, b, tc.p)
		}
	}
}

// TODO
func TestXCSDecoderReadControlSet(t *testing.T) {}

func TestXCSDecoderReadESC(t *testing.T) {
	d := newXCSDecoder()
	gl, gr := d.GL(), d.GR()
	g0, g1, g2, g3 := d.G[0], d.G[1], d.G[2], d.G[3]

	for i, tc := range []struct {
		buf  []byte
		pos  int
		ret  []byte
		size int
		err  error
		gl   GraphicSet
		gr   GraphicSet
		g    []GraphicSet
	}{
		{[]byte{0x1B, 0x6E}, 0, nil, 2, nil, g2, gr, []GraphicSet{g0, g1, g2, g3}},
		{[]byte{0x1B, 0x6F}, 0, nil, 2, nil, g3, gr, []GraphicSet{g0, g1, g2, g3}},
		{[]byte{0x1B, 0x7E}, 0, nil, 2, nil, gl, g1, []GraphicSet{g0, g1, g2, g3}},
		{[]byte{0x1B, 0x7D}, 0, nil, 2, nil, gl, g2, []GraphicSet{g0, g1, g2, g3}},
		{[]byte{0x1B, 0x7C}, 0, nil, 2, nil, gl, g3, []GraphicSet{g0, g1, g2, g3}},
		{[]byte{0x1B, 0x28, 0x38}, 0, nil, 3, nil, graphicSets[0x38], gr, []GraphicSet{graphicSets[0x38], g1, g2, g3}},
		{[]byte{0x1B, 0x00}, 0, nil, 1, errors.New("arib: ESC has invalid parameter 0x00"), gl, gr, []GraphicSet{g0, g1, g2, g3}},
	} {
		i, tc := i, tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			d := newXCSDecoder()
			d.buf = tc.buf
			buf, size, err := d.readESC(tc.pos)
			// return values without error
			if !bytes.Equal(buf, tc.ret) || size != tc.size {
				t.Errorf("%d: buf 0x%X, readESC(%d) => 0x%X, %d, _, want 0x%X, %d, _", i, tc.buf, tc.pos, buf, size, tc.ret, tc.size)
			}
			// error
			if tc.err == nil {
				if err != tc.err {
					t.Errorf("%d: buf 0x%X, readESC(%d) => _, _, %s, want _, _, %s", i, tc.buf, tc.pos, err, tc.err)
				}
			} else {
				if err.Error() != tc.err.Error() {
					t.Errorf("%d: buf 0x%X, readESC(%d) => _, _, %s, want _, _, %s", i, tc.buf, tc.pos, err, tc.err)
				}
			}
			// code elements
			got := ptr(d.GL())
			exp := ptr(tc.gl)
			if got != exp {
				t.Errorf("%d: buf 0x%X, readESC(%d) GL %d, want %d", i, tc.buf, tc.pos, got, exp)
			}
			got = ptr(d.GR())
			exp = ptr(tc.gr)
			if got != exp {
				t.Errorf("%d: buf 0x%X, readESC(%d) GR %d, want %d", i, tc.buf, tc.pos, got, exp)
			}
			for j := 0; j < 4; j++ {
				got = ptr(d.G[j])
				exp = ptr(tc.g[j])
				if got != exp {
					t.Errorf("%d: buf 0x%X, readESC(%d) G%d %d, want %d", i, tc.buf, tc.pos, j, got, exp)
				}
			}
		})
	}
}

func TestXCSDecoderDesignateGraphicSet(t *testing.T) {
	for i, tc := range []struct {
		buf  []byte
		pos  int
		size int
		gi   int
		gs   byte
	}{
		{[]byte{0xAA, 0x1B, 0x28, 0x42}, 1, 3, 0, 0x42},
		{[]byte{0xAA, 0x1B, 0x29, 0x30}, 1, 3, 1, 0x30},
		{[]byte{0xAA, 0x1B, 0x2A, 0x31}, 1, 3, 2, 0x31},
		{[]byte{0xAA, 0x1B, 0x2B, 0x42}, 1, 3, 3, 0x42},
		{[]byte{0xAA, 0x1B, 0x28, 0x20, 0x41}, 1, 4, 0, 0x41},
		{[]byte{0xAA, 0x1B, 0x29, 0x20, 0x42}, 1, 4, 1, 0x42},
		{[]byte{0xAA, 0x1B, 0x2A, 0x20, 0x43}, 1, 4, 2, 0x43},
		{[]byte{0xAA, 0x1B, 0x2B, 0x20, 0x70}, 1, 4, 3, 0x70},
		{[]byte{0xAA, 0x1B, 0x24, 0x28, 0x42}, 1, 3, 0, 0x28},
		{[]byte{0xAA, 0x1B, 0x24, 0x29, 0x39}, 1, 4, 1, 0x39},
		{[]byte{0xAA, 0x1B, 0x24, 0x2A, 0x3A}, 1, 4, 2, 0x3A},
		{[]byte{0xAA, 0x1B, 0x24, 0x2B, 0x3B}, 1, 4, 3, 0x3B},
		{[]byte{0xAA, 0x1B, 0x24, 0x28, 0x20, 0x40}, 1, 5, 0, 0x40},
		{[]byte{0xAA, 0x1B, 0x24, 0x29, 0x20, 0x40}, 1, 5, 1, 0x40},
		{[]byte{0xAA, 0x1B, 0x24, 0x2A, 0x20, 0x40}, 1, 5, 2, 0x40},
		{[]byte{0xAA, 0x1B, 0x24, 0x2B, 0x20, 0x40}, 1, 5, 3, 0x40},
	} {
		i, tc := i, tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			d := newXCSDecoder()
			d.buf = tc.buf
			size, gi, gs := d.designateGraphicSet(tc.pos)
			if size != tc.size || gi != tc.gi || gs != tc.gs {
				t.Errorf("%d: buf 0x%X, designateGraphicSet(%d) => %d, %d, 0x%X, want %d, %d, 0x%X", i, tc.buf, tc.pos, size, gi, gs, tc.size, tc.gi, tc.gs)
			}
		})
	}
}

// TODO
func TestXCSDecoderReadGL(t *testing.T) {}
func TestXCSDecoderReadGR(t *testing.T) {}

func TestXCSDecoderRevertSingleShift(t *testing.T) {
	for i, tc := range []struct {
		gl            int
		gss           int
		singleShifted bool
		expGL         int
		expSS         int
	}{
		{3, 0, false, 0, 4},
		{0, 4, false, 0, 4},
		{3, 0, true, 3, 0},
	} {
		i, tc := i, tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			d := newXCSDecoder()
			d.gl, d.gss, d.singleShifted = tc.gl, tc.gss, tc.singleShifted
			d.revertSingleShift()
			if d.gl != tc.expGL {
				t.Errorf("%d: revertSingleShift() GL %d, want %d", i, d.gl, tc.expGL)
			}
			if d.gss != tc.expSS {
				t.Errorf("%d: revertSingleShift() SS %d, want %d", i, d.gss, tc.expSS)
			}
		})
	}
}

func TestXCSDecoderParamOrNil(t *testing.T) {
	for i, tc := range []struct {
		buf []byte
		pos int
		n   int
		exp byte
	}{
		{[]byte{0xAA, 0xAB, 0xAC}, 0, 1, 0xAB},
		{[]byte{0xAA, 0xAB, 0xAC}, 0, 2, 0xAC},
		{[]byte{0xAA, 0xAB, 0xAC}, 0, 3, 0x00},
		{[]byte{0xAA, 0xAB, 0xAC}, 1, 1, 0xAC},
		{[]byte{0xAA, 0xAB, 0xAC}, 1, 2, 0x00},
		{[]byte{0xAA, 0xAB, 0xAC}, 1, 3, 0x00},
		{[]byte{0xAA, 0xAB, 0xAC}, 3, 1, 0x00},
		{[]byte{0xAA, 0xAB, 0xAC}, 4, 1, 0x00},
	} {
		i, tc := i, tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			d := newXCSDecoder()
			d.buf = tc.buf
			got := d.paramOrNil(tc.pos, tc.n)
			if got != tc.exp {
				t.Errorf("%d: paramOrNil(%d, %d) => 0x%X, want 0x%X", i, tc.pos, tc.n, got, tc.exp)
			}
		})
	}
}
