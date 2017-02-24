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
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, NITher express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package arib

import (
	"encoding/binary"

	"github.com/drillbits/go-ts/ts"
)

// NetworkID is a network_id which serves as a label to identify the delivery
// system, about which the NIT informs, from any other delivery system.
// The standardization organization shall specify allocation of the value of
// this field.
type NetworkID uint16

// OriginalNetworkID is a original_network_id which gives the label identifying
// the network_id of the originating delivery system.
type OriginalNetworkID uint16

// NIT is a Network Information Table.
type NIT ts.PSI

// NetworkID returns the NetworkID.
func (t NIT) NetworkID() NetworkID {
	return NetworkID(binary.BigEndian.Uint16(t[3:5]))
}

// VersionNumber returns the version_number.
func (t NIT) VersionNumber() int {
	return ts.VersionNumber(t)
}

// CurrentNextIndicator returns the current_next_indicator.
func (t NIT) CurrentNextIndicator() byte {
	return ts.CurrentNextIndicator(t)
}

// SectionNumber returns the section_number.
func (t NIT) SectionNumber() byte {
	return ts.SectionNumber(t)
}

// LastSectionNumber returns the last_section_number.
func (t NIT) LastSectionNumber() byte {
	return ts.LastSectionNumber(t)
}

// NetworkDescriptorsLength returns the network_descriptors_length.
func (t NIT) NetworkDescriptorsLength() int {
	return int(uint16(t[9]&0xFF) | uint16(t[8]&0x0F)<<8)
}

// Descriptors returns the descriptors.
func (t NIT) Descriptors() []ts.Descriptor {
	return ts.Descriptors(t[10 : 10+t.NetworkDescriptorsLength()])
}

// TransportStreamLoopLength returns the transport_stream_loop_length.
func (t NIT) TransportStreamLoopLength() int {
	pos := 10 + t.NetworkDescriptorsLength()
	return int(uint16(t[pos+1]&0xFF) | uint16(t[pos]&0x0F)<<8)
}

// NetworkTransportStream is an information relating to the physical
// organization of the TS carried via a given network.
type NetworkTransportStream []byte

// NetworkTransportStreams returns the list of NetworkTransportStream.
func (t NIT) NetworkTransportStreams() []NetworkTransportStream {
	var xs []NetworkTransportStream
	pos := 10 + t.NetworkDescriptorsLength() + 2
	crc32Len := 4
	for pos < (len(t) - crc32Len) {
		fixtedLen := 6 // transport_stream_id .. transport_descriptors_length
		descLoopLen := int(uint16(t[pos+4]&0x0F)<<8 | uint16(t[pos+5]&0xFF))
		nts := NetworkTransportStream(t[pos : pos+fixtedLen+descLoopLen])
		xs = append(xs, nts)
		pos += len(nts)
	}
	return xs
}

// TransportStreamID returns the transport_stream_id.
func (nts NetworkTransportStream) TransportStreamID() ts.TransportStreamID {
	return ts.TransportStreamID(binary.BigEndian.Uint16(nts[:2]))
}

// OriginalNetworkID returns the original_network_id.
func (nts NetworkTransportStream) OriginalNetworkID() OriginalNetworkID {
	return OriginalNetworkID(binary.BigEndian.Uint16(nts[2:4]))
}

// TransportDescriptorsLength returns the transport_descriptors_length.
func (nts NetworkTransportStream) TransportDescriptorsLength() int {
	return int(uint16(nts[4]&0x0F)<<8 | uint16(nts[5]&0xFF))
}

// Descriptors returns descriptors.
func (nts NetworkTransportStream) Descriptors() []ts.Descriptor {
	return ts.Descriptors(nts[6:]) // transport_stream_id .. transport_descriptors_length
}
