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

package graphicset

import (
	"bytes"
	"testing"
)

var (
	_ GraphicSet = singleByteEmptySet
	_ GraphicSet = doubleByteEmptySet
	_ GraphicSet = hiraganaSet
	_ GraphicSet = katakanaSet
	_ GraphicSet = alphanumericSet
	_ GraphicSet = kanjiSet
	_ GraphicSet = additionalSymbolSet
	_ GraphicSet = jisX0201KatakanaSet
)

func TestAdditionalSymbolSetGet(t *testing.T) {
	for i, tc := range []struct {
		b1   byte
		b2   byte
		buf  []byte
		size int
	}{
		{byte(0x7D), byte(0x2C), []byte("㍻"), 2},
		{byte(0xFF), byte(0xFF), []byte("・"), 2},
	} {
		i, tc := i, tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			got, size := additionalSymbolSet.Get(tc.b1, tc.b2)
			if !bytes.Equal(got, tc.buf) || size != tc.size {
				t.Errorf("%d: additionalSymbolSet.Get(0x%X, 0x%X) => %s, %d, want %s, %d", i, tc.b1, tc.b2, string(got), size, string(tc.buf), tc.size)
				t.Errorf("0x%X", got)
			}
		})
	}
}

func TestKanjiSetGet(t *testing.T) {
	for i, tc := range []struct {
		b1   byte
		b2   byte
		buf  []byte
		size int
	}{
		{byte(0x33), byte(0x5A), []byte("楽"), 2},
	} {
		i, tc := i, tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			got, size := kanjiSet.Get(tc.b1, tc.b2)
			if !bytes.Equal(got, tc.buf) || size != tc.size {
				t.Errorf("%d: kanjiSet.Get(0x%X, 0x%X) => %s, %d, want %s, %d", i, tc.b1, tc.b2, string(got), size, string(tc.buf), tc.size)
			}
		})
	}
}
