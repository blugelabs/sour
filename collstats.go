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

import segment "github.com/blugelabs/bluge_segment_api"

var collStats CollectionStats

type CollectionStats struct{}

func (c *CollectionStats) TotalDocumentCount() uint64 {
	return 1
}

// DocumentCount returns the number of documents with at least one term for this field
func (c *CollectionStats) DocumentCount() uint64 {
	return 1
}

// SumTotalTermFrequency returns to total number of tokens across all documents
func (c *CollectionStats) SumTotalTermFrequency() uint64 {
	return 1 // FIXME this is wrong and will affect term similarity scoring
}

func (c *CollectionStats) Merge(segment.CollectionStats) {

}
