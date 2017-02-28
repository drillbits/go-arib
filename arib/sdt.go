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

	"github.com/drillbits/go-ts/ts"
)

const crc32size = 4

// SDT is a Service Description Table.
type SDT ts.PSI

type ServiceID ts.ProgramNumber

type RunningStatus byte

// TransportStreamID returns the TransportStreamID.
func (t SDT) TransportStreamID() ts.TransportStreamID {
	return ts.TransportStreamID(binary.BigEndian.Uint16(t[3:5]))
}

// VersionNumber returns the version_number.
func (t SDT) VersionNumber() int {
	return ts.VersionNumber(t)
}

// CurrentNextIndicator returns the current_next_indicator.
func (t SDT) CurrentNextIndicator() byte {
	return ts.CurrentNextIndicator(t)
}

// SectionNumber returns the section_number.
func (t SDT) SectionNumber() byte {
	return ts.SectionNumber(t)
}

// LastSectionNumber returns the last_section_number.
func (t SDT) LastSectionNumber() byte {
	return ts.LastSectionNumber(t)
}

// OriginalNetworkID returns the OriginalNetworkID.
func (t SDT) OriginalNetworkID() OriginalNetworkID {
	return OriginalNetworkID(binary.BigEndian.Uint16(t[8:10]))
}

// Service is an information for the service that are contained within a TS.
type Service []byte

// Services returns the list of Service.
func (t SDT) Services() []Service {
	headsize := 5 // service_id .. descriptors_loop_length
	var services []Service
	pos := 11
	for pos < len(t)-crc32size {
		size := headsize + Service(t[pos:]).DescriptorsLoopLength()
		s := Service(t[pos : pos+size])
		pos += len(s)
		services = append(services, s)
	}
	return services
}

// ID returns the service_id.
func (s Service) ID() ServiceID {
	return ServiceID(binary.BigEndian.Uint16(s[:2]))
}

// EITUserDefinedFlags returns the EIT_user_defined_flags.
func (s Service) EITUserDefinedFlags() byte {
	return s[2] & 0x1C >> 2
}

// EITScheduleFlag returns the EIT_schedule_flag.
func (s Service) EITScheduleFlag() byte {
	return s[2] & 0x02 >> 1
}

// EITPresentFollowingFlag returns the EIT_present_following_flag.
func (s Service) EITPresentFollowingFlag() byte {
	return s[2] & 0x01
}

// RunningStatus returns the running_status.
func (s Service) RunningStatus() RunningStatus {
	return RunningStatus(s[3] & 0xE0 >> 5)
}

// FreeCAMode returns the free_CA_mode.
func (s Service) FreeCAMode() byte {
	return s[3] & 0x10 >> 4
}

// IsFreeCA reports whether the service is free_CA_mode.
func (s Service) IsFreeCA() bool {
	return s.FreeCAMode() == 1
}

// DescriptorsLoopLength returns the descriptors_loop_length.
func (s Service) DescriptorsLoopLength() int {
	return int(uint16(s[4]&0xFF) | uint16(s[3]&0x0F)<<8)
}

// Descriptors returns the descriptors.
func (s Service) Descriptors() []ts.Descriptor {
	return ts.Descriptors(s[5:]) // service_id .. descriptors_loop_length
}
