//  Copyright (c) 2021 The Bluge Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//              http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sour

import (
	"fmt"

	segment "github.com/blugelabs/bluge_segment_api"
)

type DocValueReader struct {
	s      *Sour
	fields []string
}

func (d *DocValueReader) VisitDocumentValues(number uint64, visitor segment.DocumentValueVisitor) error {
	if d.s.doc == nil {
		return nil
	}
	if number != internalDocNumber {
		return fmt.Errorf("unknown doc number: '%d", number)
	}

	for _, dvrField := range d.fields {
		atf, _, err := d.s.TokenFreqsAndLen(dvrField)
		if err == nil {
			for _, v := range atf {
				visitor(dvrField, v.Term())
			}
		}
	}
	return nil
}
