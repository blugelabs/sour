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
	"github.com/blugelabs/bluge/analysis"
	segment "github.com/blugelabs/bluge_segment_api"
)

var fieldDictContainsEmpty = &Dictionary{}

type Dictionary struct {
	s     *Sour
	field string
	atf   analysis.TokenFrequencies
}

func (d *Dictionary) Contains(key []byte) (bool, error) {
	if d.atf == nil {
		return false, nil
	}
	if _, ok := d.atf[string(key)]; ok {
		return true, nil
	}
	return false, nil
}

func (d *Dictionary) Close() error {
	return nil
}

type DictionaryIterator struct {
	terms       []string
	index       int
	includeFunc func(term string) bool

	next DictEntry
}

type DictEntry struct {
	term  string
	count uint64
}

func (d *DictEntry) Term() string {
	return d.term
}

func (d *DictEntry) Count() uint64 {
	return d.count
}

var fieldDictEmpty = NewFieldDictEmpty()

func NewFieldDictEmpty() *DictionaryIterator {
	return &DictionaryIterator{}
}

func NewDictionaryIteratorWithTerms(terms []string, include func(string) bool) *DictionaryIterator {
	return &DictionaryIterator{
		terms:       terms,
		includeFunc: include,
	}
}

func (d *DictionaryIterator) Next() (segment.DictionaryEntry, error) {
	for d.index < len(d.terms) {
		// if we need to skip this item increment and continue
		if d.includeFunc != nil && !d.includeFunc(d.terms[d.index]) {
			d.index++
			continue
		}
		d.next.term = d.terms[d.index]
		d.next.count = 1

		d.index++
		return &d.next, nil
	}
	return nil, nil
}

func (d *DictionaryIterator) Close() error {
	return nil
}
