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

package sour_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/sour"
)

var anHourAgo = time.Now().Add(-time.Hour)
var anHourFromNow = time.Now().Add(time.Hour)

func Example() {
	s := sour.New(bluge.InMemoryOnlyConfig())

	s.Reset(bluge.NewDocument("id").
		AddField(bluge.NewKeywordField("name", "sour")))

	dmi, err := s.Search(context.Background(),
		bluge.NewTopNSearch(0, bluge.NewTermQuery("sour").SetField("name")).
			WithStandardAggregations())
	if err != nil {
		panic(err)
	}
	if dmi.Aggregations().Count() > 0 {
		fmt.Println("matches name sour")
	} else {
		fmt.Println("does not match name sour")
	}
	// Output: matches name sour
}

func TestSour(t *testing.T) {
	cfg := bluge.InMemoryOnlyConfig()

	docA := bluge.NewDocument("A").
		AddField(bluge.NewKeywordField("name", "marty")).
		AddField(bluge.NewTextField("title", "software developer")).
		AddField(bluge.NewTextField("slogan", "code match")).
		AddField(bluge.NewNumericField("level", 10)).
		AddField(bluge.NewDateTimeField("created", time.Now()))

	sourA := sour.NewWithDocument(cfg, docA)

	tests := []struct {
		name        string
		query       bluge.Query
		expectMatch bool
	}{
		{
			name:        "match by term query - _id",
			query:       bluge.NewTermQuery("A").SetField("_id"),
			expectMatch: true,
		},
		{
			name:        "does not match by term query - _id",
			query:       bluge.NewTermQuery("B").SetField("_id"),
			expectMatch: false,
		},
		{
			name:        "match by term query - name",
			query:       bluge.NewTermQuery("marty").SetField("name"),
			expectMatch: true,
		},
		{
			name:        "does not match by term query - name",
			query:       bluge.NewTermQuery("bob").SetField("name"),
			expectMatch: false,
		},
		{
			name:        "match by match query - title",
			query:       bluge.NewMatchQuery("DeVeLoPeR").SetField("title"),
			expectMatch: true,
		},
		{
			name:        "does not match by match query - title",
			query:       bluge.NewMatchQuery("developers").SetField("title"),
			expectMatch: false,
		},
		{
			name:        "match by numeric range query - level",
			query:       bluge.NewNumericRangeQuery(9, 11).SetField("level"),
			expectMatch: true,
		},
		{
			name:        "does not match by numeric range query - level",
			query:       bluge.NewNumericRangeQuery(11, 13).SetField("level"),
			expectMatch: false,
		},
		{
			name:        "match by date range query - created",
			query:       bluge.NewDateRangeQuery(anHourAgo, anHourFromNow).SetField("created"),
			expectMatch: true,
		},
		{
			name:        "does not match by date range query - created",
			query:       bluge.NewDateRangeQuery(anHourFromNow, anHourFromNow.Add(time.Hour)).SetField("created"),
			expectMatch: false,
		},

		{
			name:        "match by regex query - slogan",
			query:       bluge.NewRegexpQuery(`co.[d-f]`).SetField("slogan"),
			expectMatch: true,
		},

		// FIXME add tests for geo and composite fields
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			req := bluge.NewTopNSearch(1, test.query)
			dmi, err := sourA.Search(context.Background(), req)
			if err != nil {
				t.Fatalf("error executing search: %v", err)
			}

			var sawMatch bool
			match, err := dmi.Next()
			for err == nil && match != nil {
				sawMatch = true
				match, err = dmi.Next()
			}
			if err != nil {
				t.Fatalf("error iterating results: %v", err)
			}
			if sawMatch != test.expectMatch {
				t.Errorf("expectedMatch: %t sawMatch: %t", test.expectMatch, sawMatch)
			}
		})
	}
}
