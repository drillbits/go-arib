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
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"

	"github.com/drillbits/go-arib/arib/xcs"
	"github.com/drillbits/go-ts/ts"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// ServiceListDescriptor is the service_list_descriptor.
// service_list_descriptor(){
//     descriptor_tag               8 uimsbf
//     descriptor_length            8 uimsbf
//     for (i=0;i<N;i++){
//         service_id              16 uimsbf
//         service_type             8 uimsbf
//     }
// }
type ServiceListDescriptor ts.Descriptor

// IsServiceListDescriptor reports whether the descriptor is the service_list_descriptor.
func IsServiceListDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0x41
}

// ToServiceListDescriptor converts the descriptor to the service_list_descriptor.
func ToServiceListDescriptor(d ts.Descriptor) (ServiceListDescriptor, error) {
	if !IsServiceListDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for service_list_descriptor", d.Tag())
	}
	return ServiceListDescriptor(d), nil
}

func (d ServiceListDescriptor) Services() []ServiceListService {
	var services []ServiceListService
	for pos := 2; pos < (len(d)); {
		size := 3 // service_id .. service_type
		s := ServiceListService(d[pos : pos+size])
		services = append(services, s)
		pos += len(s)
	}
	return services
}

type ServiceListService []byte

func (s ServiceListService) ID() ServiceID {
	return ServiceID(binary.BigEndian.Uint16(s[:2]))
}

func (s ServiceListService) Type() byte {
	return s[2]
}

// SatelliteDeliverySystemDescriptor is the satellite_delivery_system_descriptor.
// 衛星分配システム記述子
// satellite_delivery_system_descriptor(){
//     descriptor_tag               8 uimsbf [0]
//     descriptor_length            8 uimsbf [1]
//     frequency                   32 bslbf  [2-5]
//     orbital_position            16 bslbf  [6-7]
//     west_east_flag               1 bslbf  [8]
//     polarisation                 2 bslbf  [8]
//     modulation                   5 bslbf  [8]
//     symbol_rate                 28 bslbf  [9-12]
//     FEC_inner                    4 bslbf  [12]
// }
type SatelliteDeliverySystemDescriptor ts.Descriptor

// IsSatelliteDeliverySystemDescriptor reports whether the descriptor is the satellite_delivery_system_descriptor.
func IsSatelliteDeliverySystemDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0x43
}

// ToSatelliteDeliverySystemDescriptor converts the descriptor to the satellite_delivery_system_descriptor.
func ToSatelliteDeliverySystemDescriptor(d ts.Descriptor) (SatelliteDeliverySystemDescriptor, error) {
	if !IsSatelliteDeliverySystemDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for satellite_delivery_system_descriptor", d.Tag())
	}
	return SatelliteDeliverySystemDescriptor(d), nil
}

// TODO
// 32bit 4 ビット BCD コード 8 桁で周波 数を表す
// 衛星分配システム記述子では、周波数は 4 桁目以降が小数点以下となる GHz 単位で
// 符号化される。(例 012.73300GHz)
func (d SatelliteDeliverySystemDescriptor) Frequency() []byte {
	return d[2:6]
}

// TODO
// これは 16 ビットのフィールドで、4 ビット BCD コード 4 桁で、
// 4桁目が小数点以下となる度単位で軌道位置を表す。(例 144.0 度)
func (d SatelliteDeliverySystemDescriptor) OrbitalPosition() []byte {
	return d[6:8]
}

func (d SatelliteDeliverySystemDescriptor) WestEastFlag() bool {
	return d[8]&0x80>>7 == 1
}

func (d SatelliteDeliverySystemDescriptor) Polarisation() byte {
	return d[8] & 0x60 >> 5
}

func (d SatelliteDeliverySystemDescriptor) Modulation() byte {
	return d[8] & 0x1F
}

// TODO
// これは 28 ビットフィールドで、4 ビット BCD コード 7 桁で、
// 4 桁目以降が小数点以下となる Msymbol/s 単位でシンボルレート値を表す。
// (例 021.0960)
func (d SatelliteDeliverySystemDescriptor) SymbolRate() []byte {
	// 11111111 11111111 11111111 11110000
	//                            11110000 0xF0
	return d[9:13]
}

func (d SatelliteDeliverySystemDescriptor) FECInner() byte {
	return d[12] & 0x0F
}

// ServiceDescriptor is the service_descriptor.
// service_descriptor(){
//     descriptor_tag               8 uimsbf
//     descriptor_length            8 uimsbf
//     service_type                 8 uimsbf
//     service_provider_name_length 8 uimsbf
//     for (i=0;i<N;i++){
//         char                     8 uimsbf
//     }
//     service_name_length          8 uimsbf
//     for (i=0;i<N;i++){
//         char                     8 uimsbf
//     }
// }
type ServiceDescriptor ts.Descriptor

// IsServiceDescriptor reports whether the descriptor is the service_descriptor.
func IsServiceDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0x48
}

// ToServiceDescriptor converts the descriptor to the service_descriptor.
func ToServiceDescriptor(d ts.Descriptor) (ServiceDescriptor, error) {
	if !IsServiceDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for service_descriptor", d.Tag())
	}
	return ServiceDescriptor(d), nil
}

func (d ServiceDescriptor) Type() byte {
	return d[2]
}

func (d ServiceDescriptor) ProviderNameLength() int {
	return int(d[3])
}

func (d ServiceDescriptor) ProviderName() (string, error) {
	return decodeXCS(d[4 : 4+d.ProviderNameLength()])
}

func (d ServiceDescriptor) NameLength() int {
	return int(d[4+d.ProviderNameLength()])
}

func (d ServiceDescriptor) Name() (string, error) {
	pos := 4 + d.ProviderNameLength() + 1
	return decodeXCS(d[pos : pos+d.NameLength()])
}

// ShortEventDescriptor is the short_event_descriptor.
type ShortEventDescriptor ts.Descriptor

// IsShortEventDescriptor reports whether the descriptor is the short_event_descriptor.
func IsShortEventDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0x4D
}

// ToShortEventDescriptor converts the descriptor to the short_event_descriptor.
func ToShortEventDescriptor(d ts.Descriptor) (ShortEventDescriptor, error) {
	if !IsShortEventDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for short_event_descriptor", d.Tag())
	}
	return ShortEventDescriptor(d), nil
}

// ISO639LanguageCode returns the language code.
func (d ShortEventDescriptor) ISO639LanguageCode() (string, error) {
	return decodeISO8859_1(d[2:5])
}

// EventNameLength returns the length of the event name.
func (d ShortEventDescriptor) EventNameLength() int {
	return int(d[5])
}

// EventName returns the name of the event.
func (d ShortEventDescriptor) EventName() (string, error) {
	r := bytes.NewReader(d[6 : 6+d.EventNameLength()])
	tr := transform.NewReader(r, xcs.XCSEncoding.NewDecoder())
	b, err := ioutil.ReadAll(tr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// TextLength returns the length of the text.
func (d ShortEventDescriptor) TextLength() int {
	n := 6 + d.EventNameLength()
	return int(d[n])
}

// Text returns the text of the short_event_descriptor.
func (d ShortEventDescriptor) Text() (string, error) {
	n := 6 + d.EventNameLength() + 1
	return decodeXCS(d[n : n+d.TextLength()])
}

// ComponentDescriptor is the component_descriptor.
type ComponentDescriptor ts.Descriptor

// IsComponentDescriptor reports whether the descriptor is the component_descriptor.
func IsComponentDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0x50
}

// ToComponentDescriptor converts the descriptor to the component_descriptor.
func ToComponentDescriptor(d ts.Descriptor) (ComponentDescriptor, error) {
	if !IsComponentDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for component_descriptor", d.Tag())
	}
	return ComponentDescriptor(d), nil
}

func (d ComponentDescriptor) StreamContent() byte {
	return d[2] & 0x0F
}

func (d ComponentDescriptor) ComponentType() byte {
	return d[3]
}

func (d ComponentDescriptor) ComponentTag() byte {
	return d[4]
}

// ISO639LanguageCode returns the language code.
func (d ComponentDescriptor) ISO639LanguageCode() (string, error) {
	return decodeISO8859_1(d[5:8])
}

// Text returns the text of the component_descriptor.
func (d ComponentDescriptor) Text() (string, error) {
	return decodeXCS(d[8:len(d)])
}

// ContentDescriptor is the content_descriptor.
type ContentDescriptor ts.Descriptor

// IsContentDescriptor reports whether the descriptor is the content_descriptor.
func IsContentDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0x54
}

// ToContentDescriptor converts the descriptor to the content_descriptor.
func ToContentDescriptor(d ts.Descriptor) (ContentDescriptor, error) {
	if !IsContentDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for content_descriptor", d.Tag())
	}
	return ContentDescriptor(d), nil
}

func (d ContentDescriptor) Nibbles() []Nibble {
	var nibbles []Nibble
	pos := 2
	l := 2
	for pos < len(d) {
		nibbles = append(nibbles, Nibble(d[pos:pos+l]))
		pos += l
	}
	return nibbles
}

// Nibble is the genre of the event.
type Nibble []byte

// ContentNibbleLevel1 returns the genre of the content 1.
func (n Nibble) ContentNibbleLevel1() byte {
	return n[0] & 0xF0 >> 4
}

// ContentNibbleLevel2 returns the genre of the content 2.
func (n Nibble) ContentNibbleLevel2() byte {
	return n[0] & 0x0F
}

// UserNibble1 returns the user defined genre 1.
func (n Nibble) UserNibble1() byte {
	return n[0] & 0xF0 >> 4
}

// UserNibble2 returns the user defined genre 2.
func (n Nibble) UserNibble2() byte {
	return n[0] & 0x0F
}

// TODO: continue...

// EventGroupDescriptor is the event_group_descriptor.
// descriptor_tag            8 [0]
// descriptor_length         8 [1]
// group_type                4 [2]
// event_count               4 [2]
// for event_count
//   service_id             16 [0]
//   event_id               16 [2]
// if group_type == 4 || 5
//   for
//     original_network_id  16 [0]
//     transport_stream_id  16 [2]
//     service_id           16 [4]
//     event_id             16 [6]
// else
//   for
//     private_data_byte     8 []
type EventGroupDescriptor ts.Descriptor

// IsEventGroupDescriptor reports whether the descriptor is the event_group_descriptor.
func IsEventGroupDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0xD6
}

// ToEventGroupDescriptor converts the descriptor to the event_group_descriptor.
func ToEventGroupDescriptor(d ts.Descriptor) (EventGroupDescriptor, error) {
	if !IsEventGroupDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for event_group_descriptor", d.Tag())
	}
	return EventGroupDescriptor(d), nil
}

func (d EventGroupDescriptor) GroupType() int {
	return int(d[2] & 0xF0 >> 4)
}

func (d EventGroupDescriptor) EventCount() int {
	return int(d[2] & 0x0F)
}

func (d EventGroupDescriptor) Events() []*EventGroupEvent {
	var events []*EventGroupEvent
	pos := 3
	rest := len(d) - 3 // descriptor_tag, descriptor_length, group_type, event_count
	l := rest / d.EventCount()
	for pos < len(d) {
		events = append(events, &EventGroupEvent{d[pos : pos+l], d.GroupType()})
		pos += l
	}
	return events
}

type EventGroupEvent struct {
	buf       []byte
	groupType int
}

func (e *EventGroupEvent) ServiceID() []byte {
	return e.buf[0:2]
}

func (e *EventGroupEvent) EventID() uint16 {
	return binary.BigEndian.Uint16(e.buf[2:4])
}

func (e *EventGroupEvent) RelatedEvents() []RelatedEvent {
	if e.groupType != 4 && e.groupType != 5 {
		return nil
	}
	var events []RelatedEvent
	pos := 4
	l := 8
	for pos < len(e.buf) {
		events = append(events, RelatedEvent(e.buf[pos:pos+l]))
		pos += l
	}
	return events
}

func (e *EventGroupEvent) PrivateDataBytes() []byte {
	if e.groupType == 4 && e.groupType == 5 {
		return nil
	}
	return e.buf[4:len(e.buf)]
}

type RelatedEvent []byte

func (e RelatedEvent) OriginalNetworkID() []byte {
	return e[0:2]
}

func (e RelatedEvent) TransportStreamID() []byte {
	return e[2:4]
}

func (e RelatedEvent) ServiceID() []byte {
	return e[4:6]
}

func (e RelatedEvent) EventID() uint16 {
	return binary.BigEndian.Uint16(e[6:8])
}

// DigitalCopyControlDescriptor is the digital_copy_control_descriptor.
// descriptor_tag                                8 [0]
// descriptor_length                             8 [1]
// digital_recording_control_data                2 [2]
// maximum_bitrate_flag                          1 [2]
// component_control_flag                        1 [2]
// user_defined                                  4 [2]
// maximum_bitrate if maximum_bitrate            8 [3?]
// component_control_length if component_control 8 [4?]
// for
//   component_tag                               8 [0]
//   digital_recording_control_data              2 [1]
//   maximum_bitrate_flag                        1 [1]
//   reserved_future_use                         1 [1]
//   user_defined                                4 [1]
//   maximum_bitrate if maximum_bitrate          8 [2?]
type DigitalCopyControlDescriptor ts.Descriptor

// IsDigitalCopyControlDescriptor reports whether the descriptor is the digital_copy_control_descriptor.
func IsDigitalCopyControlDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0xC1
}

// ToDigitalCopyControlDescriptor converts the descriptor to the audio_component_descriptor.
func ToDigitalCopyControlDescriptor(d ts.Descriptor) (DigitalCopyControlDescriptor, error) {
	if !IsDigitalCopyControlDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for digital_copy_control_descriptor", d.Tag())
	}
	return DigitalCopyControlDescriptor(d), nil
}

// DigitalRecordingControlData returns the digital recording control data.
func (d DigitalCopyControlDescriptor) DigitalRecordingControlData() byte {
	return d[2] & 0xC0 >> 6
}

// MaximumBitrateFlag returns maximum bitrate flag.
func (d DigitalCopyControlDescriptor) MaximumBitrateFlag() bool {
	return d[2]&0x20>>5 == 1
}

// HasMaximumBitrate reports whether the descriptor has maximum bitrate.
func (d DigitalCopyControlDescriptor) HasMaximumBitrate() bool {
	return d.MaximumBitrateFlag()
}

// ComponentControlFlag returns component control flag.
func (d DigitalCopyControlDescriptor) ComponentControlFlag() bool {
	return d[2]&0x10>>4 == 1
}

// HasComponentControl reports whether the descriptor has component controls.
func (d DigitalCopyControlDescriptor) HasComponentControl() bool {
	return d.ComponentControlFlag()
}

// MaximumBitrate returns the maximum bitrate of the descriptor.
func (d DigitalCopyControlDescriptor) MaximumBitrate() int {
	if !d.HasMaximumBitrate() {
		return 0
	}
	return int(d[3])
}

// ComponentControlLength returns the length of component controls.
func (d DigitalCopyControlDescriptor) ComponentControlLength() int {
	n := 3
	if d.HasComponentControl() {
		n++
	}
	return int(d[n])
}

// Components returns components of the descriptor.
func (d DigitalCopyControlDescriptor) Components() []DigitalCopyControlComponent {
	var components []DigitalCopyControlComponent
	n := 3
	if d.HasMaximumBitrate() {
		n++
	}
	if d.HasComponentControl() {
		n++
	}
	for n < len(d) {
		tmp := DigitalCopyControlComponent(d[n : n+2])
		l := 2
		if tmp.HasMaximumBitrate() {
			l++
		}
		c := DigitalCopyControlComponent(d[n : n+l])
		n += len(c)
		components = append(components, c)
	}
	return components
}

// DigitalCopyControlComponent is the component of the digital_copy_control_descriptor.
type DigitalCopyControlComponent []byte

// Tag returns the tag of the component.
func (c DigitalCopyControlComponent) Tag() byte {
	return c[0]
}

// DigitalRecordingControlData returns the digital recording control data.
func (c DigitalCopyControlComponent) DigitalRecordingControlData() byte {
	return c[1] & 0xC0 >> 6
}

// MaximumBitrateFlag returns maximum bitrate flag.
func (c DigitalCopyControlComponent) MaximumBitrateFlag() bool {
	return c[1]&0x20>>5 == 1
}

// HasMaximumBitrate reports whether the component has maximum bitrate.
func (c DigitalCopyControlComponent) HasMaximumBitrate() bool {
	return c.MaximumBitrateFlag()
}

// MaximumBitrate returns the maximum bitrate of the component.
func (c DigitalCopyControlComponent) MaximumBitrate() int {
	if !c.HasMaximumBitrate() {
		return 0
	}
	return int(c[2])
}

// AudioComponentDescriptor is the audio_component_descriptor.
type AudioComponentDescriptor ts.Descriptor

// IsAudioComponentDescriptor reports whether the descriptor is the audio_component_descriptor.
func IsAudioComponentDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0xC4
}

// ToAudioComponentDescriptor converts the descriptor to the audio_component_descriptor.
func ToAudioComponentDescriptor(d ts.Descriptor) (AudioComponentDescriptor, error) {
	if !IsAudioComponentDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for audio_component_descriptor", d.Tag())
	}
	return AudioComponentDescriptor(d), nil
}

func (d AudioComponentDescriptor) StreamContent() byte {
	return d[2] & 0x0F
}

func (d AudioComponentDescriptor) ComponentType() byte {
	return d[3]
}

func (d AudioComponentDescriptor) ComponentTag() byte {
	return d[4]
}

func (d AudioComponentDescriptor) StreamType() byte {
	return d[5]
}

func (d AudioComponentDescriptor) SimulcastGroupTag() byte {
	return d[6]
}

func (d AudioComponentDescriptor) ESMultiLingualFlag() bool {
	return d[7]&0x80>>7 == 1
}

func (d AudioComponentDescriptor) MainComponentFlag() bool {
	return d[7]&0x40>>6 == 1
}

func (d AudioComponentDescriptor) QualityIndicator() bool {
	return d[7]&0x30>>4 == 1
}

func (d AudioComponentDescriptor) SamplingRate() byte {
	return d[7] & 0xE
}

// ISO639LanguageCode returns the language code.
func (d AudioComponentDescriptor) ISO639LanguageCode() (string, error) {
	return decodeISO8859_1(d[8:11])
}

// ISO639LanguageCode2 returns the second language code if the audio_component_descriptor is multi lingual.
func (d AudioComponentDescriptor) ISO639LanguageCode2() (string, error) {
	if !d.ESMultiLingualFlag() {
		return "", nil
	}
	return decodeISO8859_1(d[11:14])
}

// Text returns the text of the audio_component_descriptor.
func (d AudioComponentDescriptor) Text() (string, error) {
	n := 11
	if d.ESMultiLingualFlag() {
		n += 3 // size of ISO_639_language_code_2
	}
	return decodeXCS(d[n:len(d)])
}

// DataContentDescriptor is the data_content_descriptor.
type DataContentDescriptor ts.Descriptor

// IsDataContentDescriptor reports whether the descriptor is the data_content_descriptor.
func IsDataContentDescriptor(d ts.Descriptor) bool {
	return d.Tag() == 0xC7
}

// ToDataContentDescriptor converts the descriptor to the data_content_descriptor.
func ToDataContentDescriptor(d ts.Descriptor) (DataContentDescriptor, error) {
	if !IsDataContentDescriptor(d) {
		return nil, fmt.Errorf("0x%02X is not a tag for data_content_descriptor", d.Tag())
	}
	return DataContentDescriptor(d), nil
}

// TODO
func (d DataContentDescriptor) DataComponentID() []byte {
	return d[2:4]
}

// TODO
func (d DataContentDescriptor) EntryComponent() byte {
	return d[4]
}

// TODO
func (d DataContentDescriptor) SelectorLength() int {
	return int(d[5])
}

// TODO
func (d DataContentDescriptor) SelectorBytes() []byte {
	return d[6 : 6+d.SelectorLength()]
}

// TODO
func (d DataContentDescriptor) NumOfComponentRef() int {
	return int(d[6+d.SelectorLength()])
}

func (d DataContentDescriptor) offsetComponentRefs() int {
	return 6 + d.SelectorLength() + 1
}

// TODO
func (d DataContentDescriptor) ComponentRefs() []byte {
	n := d.offsetComponentRefs()
	return d[n : n+d.NumOfComponentRef()]
}

// ISO639LanguageCode returns the language code.
func (d DataContentDescriptor) ISO639LanguageCode() (string, error) {
	n := d.offsetComponentRefs() + d.NumOfComponentRef()
	return decodeISO8859_1(d[n : n+3])
}

// TextLength returns the length of the text.
func (d DataContentDescriptor) TextLength() int {
	n := d.offsetComponentRefs() + d.NumOfComponentRef() + 3
	return int(d[n])
}

// Text returns the text of the data_content_descriptor.
func (d DataContentDescriptor) Text() (string, error) {
	n := d.offsetComponentRefs() + d.NumOfComponentRef() + 3 + 1
	return decodeXCS(d[n : n+d.TextLength()])
}

func decode(b []byte, t transform.Transformer) (string, error) {
	r := bytes.NewReader(b)
	tr := transform.NewReader(r, t)
	b, err := ioutil.ReadAll(tr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeXCS(b []byte) (string, error) {
	return decode(b, xcs.XCSEncoding.NewDecoder())
}

func decodeISO8859_1(b []byte) (string, error) {
	return decode(b, charmap.ISO8859_1.NewDecoder())
}
