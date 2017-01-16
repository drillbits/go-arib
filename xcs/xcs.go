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
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

// Control codes
const (
	NULL   = 0x00
	BEL    = 0x07
	APB    = 0x08
	APF    = 0x09
	APD    = 0x0A
	APU    = 0x0B
	CS     = 0x0C
	APR    = 0x0D
	LS1    = 0x0E
	LS0    = 0x0F
	PAPF   = 0x16
	CAN    = 0x18
	SS2    = 0x19
	ESC    = 0x1B
	APS    = 0x1C
	SS3    = 0x1D
	RS     = 0x1E
	US     = 0x1F
	SP     = 0x20
	DEL    = 0x7F
	BKF    = 0x80
	RDF    = 0x81
	GRF    = 0x82
	YLF    = 0x83
	BLF    = 0x84
	MGF    = 0x85
	CNF    = 0x86
	WHF    = 0x87
	SSZ    = 0x88
	MSZ    = 0x89
	NSZ    = 0x8A
	SZX    = 0x8B
	COL    = 0x90
	FLC    = 0x91
	CDC    = 0x92
	POL    = 0x93
	WMM    = 0x94
	MACRO  = 0x95
	HLC    = 0x97
	RPC    = 0x98
	SPL    = 0x99
	STL    = 0x9A
	CSI    = 0x9B
	TIME   = 0x9D
	CC1000 = 0xA0 // 10/0
	CC1515 = 0xFF // 15/15
)

// XCSEncoding is the ARIB External Character Set Encoding.
var XCSEncoding encoding.Encoding = &Encoding{
	Decoder: newXCSDecoder(),
	Encoder: xcsEncoder{},
}

// Encoding is an implementation of the Encoding interface.
type Encoding struct {
	Decoder transform.Transformer
	Encoder transform.Transformer
}

// NewDecoder returns a Decoder.
func (e *Encoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: e.Decoder}
}

// NewEncoder returns an Encoder.
func (e *Encoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: e.Encoder}
}
