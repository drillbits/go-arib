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

package arib

import (
	"testing"

	"github.com/drillbits/go-ts/ts"
)

func TestEIT(t *testing.T) {
	// table_id: 0x4E
	//
	// section_syntax_indicator: 1
	// reserved_future_use: 0
	// reserved: 11
	// section_length: 001111111111
	// = 00000011 11111111 = 0x03 0xFF = 1023
	//
	// service_id: 0000010000001000 = 0x04 0x08 = 1032
	//
	// reserved: 11
	// version_number: 01111 = 15
	// current_next_indicator: 0
	// = 11011110 = 0xDE
	//
	// section_number: 11111111 = 0xFF = 255
	// last_section_number: 00001111 = 0x0F = 15
	// transport_stream_id: 0111111111100001 = 0x7F 0xE1 = 32737
	// original_network_id: 0111111111100001 = 0x7F 0xE1 = 32737
	// segment_last_section_number: 00001111 = 0x0F = 15
	// last_table_id: 00000001 = 0x01
	b := []byte{0x4E, 0xB3, 0xFF, 0x04, 0x08, 0xDE, 0xFF, 0x0F, 0x7F, 0xE1, 0x7F, 0xE1, 0x0F, 0x01}
	eit := EIT(b)
	t.Errorf("table_id:                    0x%02X\n", ts.PSI(eit).TableID())
	t.Errorf("section_syntax_indicator:    %b\n", ts.PSI(eit).SectionSyntaxIndicator())
	t.Errorf("section_length:              %d\n", ts.PSI(eit).SectionLength())
	t.Errorf("service_id:                  %d\n", eit.ServiceID())
	t.Errorf("version_number:              %d\n", eit.VersionNumber())
	t.Errorf("current_next_indicator:      %b\n", eit.CurrentNextIndicator())
	t.Errorf("section_number:              %d\n", eit.SectionNumber())
	t.Errorf("last_section_number:         %d\n", eit.LastSectionNumber())
	t.Errorf("transport_stream_id:         %d\n", eit.TransportStreamID())
	t.Errorf("original_network_id:         %d\n", eit.OriginalNetworkID())
	t.Errorf("segment_last_section_number: %d\n", eit.SegmentLastSectionNumber())
	t.Errorf("last_table_id:               0x%02X\n", eit.LastTableID())
}
