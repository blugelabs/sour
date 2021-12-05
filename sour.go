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
	"context"
	"fmt"
	"sort"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/search"
	segment "github.com/blugelabs/bluge_segment_api"
)

const internalDocNumber uint64 = 0

type Sour struct {
	cfg bluge.Config
	doc *bluge.Document

	// always built during analysis
	fieldIndexes    map[string]int
	fieldNames      []string
	fieldTokenFreqs []analysis.TokenFrequencies
	fieldLens       []int

	// deferred build and cache
	sortedTerms map[string][]string
}

func New(cfg bluge.Config) *Sour {
	return &Sour{
		cfg:         cfg,
		sortedTerms: make(map[string][]string),
	}
}

func NewWithDocument(cfg bluge.Config, doc *bluge.Document) *Sour {
	rv := New(cfg)
	rv.Reset(doc)
	return rv
}

func (s *Sour) Search(ctx context.Context, req bluge.SearchRequest) (search.DocumentMatchIterator, error) {
	collector := req.Collector()
	searcher, err := req.Searcher(s, s.cfg)
	if err != nil {
		return nil, err
	}

	return collector.Collect(ctx, req.Aggregations(), searcher)
}

func (s *Sour) DocumentValueReader(fields []string) (segment.DocumentValueReader, error) {
	return &DocValueReader{
		s:      s,
		fields: fields,
	}, nil
}

func (s *Sour) VisitStoredFields(number uint64, visitor segment.StoredFieldVisitor) error {
	// FIXME no stored fields
	return nil
}

func (s *Sour) CollectionStats(field string) (segment.CollectionStats, error) {
	return &collStats, nil
}

func (s *Sour) DictionaryLookup(field string) (segment.DictionaryLookup, error) {
	if s.doc == nil {
		return fieldDictContainsEmpty, nil
	}
	atf, _, err := s.TokenFreqsAndLen(field)
	if err != nil {
		// only error is field doesn't exist in doc
		return fieldDictContainsEmpty, nil
	}
	return &Dictionary{
		s:     s,
		field: field,
		atf:   atf,
	}, nil
}

func automatonMatch(la segment.Automaton, termStr string) bool {
	state := la.Start()
	for i := range []byte(termStr) {
		state = la.Accept(state, termStr[i])
		if !la.CanMatch(state) {
			return false
		}
	}
	return la.IsMatch(state)
}

// DictionaryIterator provides a way to explore the terms used in the
// specified field.  You can optionally filter these terms
// by the provided Automaton, or start/end terms.
func (s *Sour) DictionaryIterator(field string, automaton segment.Automaton, start,
	end []byte) (segment.DictionaryIterator, error) {
	if s.doc == nil {
		return fieldDictEmpty, nil
	}
	fieldSortedTerms, err := s.SortedTermsForField(field)
	if err != nil {
		// only error is field doesn't exist in doc
		return fieldDictEmpty, nil
	}
	return NewDictionaryIteratorWithTerms(fieldSortedTerms, func(s string) bool {
		return automatonMatch(automaton, s)
	}), nil
}

// PostingsIterator provides a way to find information about all documents
// that use the specified term in the specified field.
func (s *Sour) PostingsIterator(term []byte, field string, includeFreq, includeNorm,
	includeTermVectors bool) (segment.PostingsIterator, error) {
	if s.doc == nil {
		return termFieldReaderEmpty, nil
	}
	atf, l, err := s.TokenFreqsAndLen(field)
	if err != nil {
		// only error is field doesn't exist in doc
		return termFieldReaderEmpty, nil
	}
	tf, ok := atf[string(term)]
	if !ok {
		return termFieldReaderEmpty, nil
	}

	return NewTermFieldReaderFromTokenFreqAndLen(tf, l, includeFreq, includeNorm, includeTermVectors), nil
}

func (s *Sour) Close() error {
	return nil
}

// Add ***********

func (s *Sour) newField(field bluge.Field) {
	af := field.AnalyzedTokenFrequencies()

	// bleve analysis will leave field empty for non-composite fields, fix that here
	for _, tf := range af {
		for _, loc := range tf.Locations {
			if loc.FieldVal == "" {
				loc.FieldVal = field.Name()
			}
		}
	}

	fieldIdx, exists := s.fieldIndexes[field.Name()]
	if !exists {
		s.fieldIndexes[field.Name()] = len(s.fieldTokenFreqs)
		s.fieldNames = append(s.fieldNames, field.Name())
		s.fieldTokenFreqs = append(s.fieldTokenFreqs, af)
		s.fieldLens = append(s.fieldLens, field.Length())
	} else {
		s.fieldTokenFreqs[fieldIdx].MergeAll(field.Name(), af)
		s.fieldLens[fieldIdx] += field.Length()
	}
}

func (s *Sour) analyze() {
	// let bluge do analysis
	s.doc.Analyze()

	for _, field := range *s.doc {
		s.newField(field)
	}
}

func (s *Sour) Reset(doc *bluge.Document) {
	// clear analysis
	s.fieldIndexes = make(map[string]int, len(s.fieldNames))
	s.fieldNames = s.fieldNames[:0]
	s.fieldTokenFreqs = s.fieldTokenFreqs[:0]
	s.fieldLens = s.fieldLens[:0]

	// clear cache
	for k := range s.sortedTerms {
		s.sortedTerms[k] = s.sortedTerms[k][:0]
	}

	// init new doc
	s.doc = doc
	s.analyze()
}

func (s *Sour) fieldIndex(name string) (int, error) {
	if idx, ok := s.fieldIndexes[name]; ok {
		return idx, nil
	}
	return 0, fmt.Errorf("no field named: %s", name)
}

func (s *Sour) Fields() []string {
	return s.fieldNames
}

func (s *Sour) SortedTermsForField(fieldName string) ([]string, error) {
	fieldIdx, err := s.fieldIndex(fieldName)
	if err != nil {
		return nil, err
	}

	terms, ok := s.sortedTerms[fieldName]
	if ok && len(terms) > 0 {
		return terms, nil
	}

	atf := s.fieldTokenFreqs[fieldIdx]
	for k := range atf {
		terms = append(terms, k)
	}
	sort.Strings(terms)
	s.sortedTerms[fieldName] = terms
	return terms, nil
}

func (s *Sour) TokenFreqsAndLen(fieldName string) (analysis.TokenFrequencies, int, error) {
	fieldIdx, err := s.fieldIndex(fieldName)
	if err != nil {
		return nil, 0, err
	}
	return s.fieldTokenFreqs[fieldIdx], s.fieldLens[fieldIdx], nil
}
