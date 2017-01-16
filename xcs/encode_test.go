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
	"io/ioutil"
	"testing"

	"golang.org/x/text/transform"
)

var _ transform.Transformer = (*xcsEncoder)(nil)

func TestEncode(t *testing.T) {
	s := "おかあさんといっしょ【字】"
	exp := "arib: xcs encoding not supported"

	tr := transform.NewReader(bytes.NewReader([]byte(s)), XCSEncoding.NewEncoder())
	_, err := ioutil.ReadAll(tr)
	if err == nil || err.Error() != exp {
		t.Errorf("Encode(%s) occurs error %s, want %s", s, err, exp)
	}
}
