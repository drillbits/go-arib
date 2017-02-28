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
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/drillbits/go-ts/ts"
)

func IsEIT(p ts.Packet) bool {
	if !p.HasPayload() {
		return false
	}
	tid := p.Payload()[0]
	return tid >= 0x4E && tid <= 0x6F
}

// EIT is a Event Information Table.
type EIT ts.PSI

type EventID uint16

// ServiceID returns the ServiceID.
func (t EIT) ServiceID() ServiceID {
	return ServiceID(binary.BigEndian.Uint16(t[3:5]))
}

// VersionNumber returns the version_number.
func (t EIT) VersionNumber() int {
	return ts.VersionNumber(t)
}

// CurrentNextIndicator returns the current_next_indicator.
func (t EIT) CurrentNextIndicator() byte {
	return ts.CurrentNextIndicator(t)
}

// SectionNumber returns the section_number.
func (t EIT) SectionNumber() byte {
	return ts.SectionNumber(t)
}

// LastSectionNumber returns the last_section_number.
func (t EIT) LastSectionNumber() byte {
	return ts.LastSectionNumber(t)
}

// TransportStreamID returns the TransportStreamID.
func (t EIT) TransportStreamID() ts.TransportStreamID {
	return ts.TransportStreamID(binary.BigEndian.Uint16(t[8:10]))
}

// OriginalNetworkID returns the OriginalNetworkID.
func (t EIT) OriginalNetworkID() OriginalNetworkID {
	return OriginalNetworkID(binary.BigEndian.Uint16(t[10:12]))
}

// SegmentLastSectionNumber returns the segment_last_section_number.
func (t EIT) SegmentLastSectionNumber() int {
	return int(t[12])
}

// LastTableID returns the last_table_id.
func (t EIT) LastTableID() ts.TableID {
	return ts.TableID(t[13])
}

// Event is an information for the event.
type Event []byte

// Events returns the events.
func (t EIT) Events() []Event {
	headsize := 12 // event_id .. descriptors_loop_length
	var events []Event
	pos := 14
	for pos < len(t)-crc32size {
		size := headsize + Event(t[pos:]).DescriptorsLoopLength()
		e := Event(t[pos : pos+size])
		pos += len(e)
		events = append(events, e)
	}
	return events
}

// ID returns the event_id.
func (e Event) ID() EventID {
	return EventID(binary.BigEndian.Uint16(e[:2]))
}

// StartTime returns the start_time.
func (e Event) StartTime() time.Time {
	y, m, d := decodeMJD(e[2], e[3])
	t := time.Date(y+1900, time.Month(m), d, 0, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60))
	return t.Add(bcd(e[4], e[5], e[6]))
}

// Duration returns the duration.
func (e Event) Duration() time.Duration {
	return bcd(e[7], e[8], e[9])
}

// RunningStatus returns the running_status.
func (e Event) RunningStatus() RunningStatus {
	return RunningStatus(e[10] & 0xE0 >> 5)
}

// FreeCAMode returns the free_CA_mode.
func (e Event) FreeCAMode() byte {
	return e[10] & 0x10 >> 4
}

// IsFreeCA reports whether the event is free_CA_mode.
func (e Event) IsFreeCA() bool {
	return e.FreeCAMode() == 1
}

// DescriptorsLoopLength returns the descriptors_loop_length.
func (e Event) DescriptorsLoopLength() int {
	return int(uint16(e[11]&0xFF) | uint16(e[10]&0x0F)<<8)
}

// Descriptors returns the descriptors.
func (e Event) Descriptors() []ts.Descriptor {
	return ts.Descriptors(e[12:]) // event_id .. descriptors_loop_length
}

func decodeMJD(b1, b2 byte) (int, int, int) {
	mjd := float64(uint16(b1&0xFF)<<8 | uint16(b2&0xFF))
	y := math.Trunc((mjd - 15078.2) / 365.25)
	m := math.Trunc((mjd - 14956.1 - math.Trunc(y*365.25)) / 30.6001)
	d := mjd - 14956 - math.Trunc(y*365.25) - math.Trunc(m*30.6001)
	var k float64
	if m == 14 || m == 15 {
		k = 1
	}
	y += k
	m = m - 1 - k*12
	return int(y), int(m), int(d)
}

func bcd(b1, b2, b3 byte) time.Duration {
	hms := fmt.Sprintf("%02Xh%02Xm%02Xs", b1, b2, b3)
	d, err := time.ParseDuration(hms)
	if err != nil {
		return time.Duration(-1)
	}
	return d
}
