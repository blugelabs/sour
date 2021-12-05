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
	"math"

	"github.com/blugelabs/bluge/analysis"
	segment "github.com/blugelabs/bluge_segment_api"
)

type Location struct {
	field string
	start int
	end   int
	pos   int
}

func (l Location) Field() string {
	return l.field
}
func (l Location) Start() int {
	return l.start
}
func (l Location) End() int {
	return l.end
}
func (l Location) Pos() int {
	return l.pos
}
func (l Location) Size() int { return 0 }

type Posting struct {
	term string
	num  uint64
	freq uint64
	norm float64
	locs []segment.Location
}

func (p *Posting) Number() uint64 {
	return p.num
}
func (p *Posting) SetNumber(num uint64) {
	p.num = num
}
func (p *Posting) Frequency() int {
	return int(p.freq)
}
func (p *Posting) Norm() float64 {
	return p.norm
}
func (p *Posting) Locations() []segment.Location {
	return p.locs
}
func (p *Posting) Size() int { return 0 }

type TermFieldReader struct {
	tf                 *analysis.TokenFreq
	len                int
	done               bool
	includeFreq        bool
	includeNorm        bool
	includeTermVectors bool

	preAlloced Posting
}

var termFieldReaderEmpty = NewTermFieldReaderEmpty()

func NewTermFieldReaderEmpty() *TermFieldReader {
	return &TermFieldReader{
		done: true,
	}
}

func NewTermFieldReaderFromTokenFreqAndLen(tf *analysis.TokenFreq, l int, includeFreq, includeNorm,
	includeTermVectors bool) *TermFieldReader {
	return &TermFieldReader{
		tf:                 tf,
		len:                l,
		includeFreq:        includeFreq,
		includeNorm:        includeNorm,
		includeTermVectors: includeTermVectors,
	}
}

func normForLen(l int) float64 {
	return float64(float32(1 / math.Sqrt(float64(l))))
}

func (t *TermFieldReader) Next() (segment.Posting, error) {
	if t.done {
		return nil, nil
	}
	rv := &t.preAlloced
	rv.term = string(t.tf.Term())
	rv.num = internalDocNumber
	if t.includeFreq {
		rv.freq = uint64(t.tf.Frequency())
	}
	if t.includeNorm {
		rv.norm = normForLen(t.len)
	}
	if t.includeTermVectors {
		locs := t.tf.Locations
		if cap(rv.locs) < len(locs) {
			rv.locs = make([]segment.Location, len(locs))
			backing := make([]Location, len(locs))
			for i := range backing {
				rv.locs[i] = &backing[i]
			}
		}
		rv.locs = rv.locs[:len(locs)]
		for i, loc := range locs {
			rv.locs[i] = Location{
				start: loc.Start(),
				end:   loc.End(),
				pos:   loc.Pos(),
				field: loc.Field(),
			}
		}
	}
	t.done = true
	return rv, nil
}

// Advance resets the enumeration at specified document or its immediate
// follower.
func (t *TermFieldReader) Advance(docNum uint64) (segment.Posting, error) {
	if t.done {
		return nil, nil
	}
	if docNum > internalDocNumber {
		// seek is after our internal id
		t.done = true
		return nil, nil
	}
	return t.Next()
}

func (t *TermFieldReader) Count() uint64 {
	if t.tf != nil {
		return 1
	}
	return 0
}

func (t *TermFieldReader) Size() int {
	return 0
}

func (t *TermFieldReader) Empty() bool {
	return t.done
}

func (t *TermFieldReader) Close() error {
	return nil
}
