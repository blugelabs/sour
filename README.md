# sour

[![PkgGoDev](https://pkg.go.dev/badge/github.com/blugelabs/sour)](https://pkg.go.dev/github.com/blugelabs/sour)
[![Tests](https://github.com/blugelabs/sour/workflows/Tests/badge.svg?branch=master&event=push)](https://github.com/blugelabs/sour/actions?query=workflow%3ATests+event%3Apush+branch%3Amaster)
[![Lint](https://github.com/blugelabs/sour/workflows/Lint/badge.svg?branch=master&event=push)](https://github.com/blugelabs/sour/actions?query=workflow%3ALint+event%3Apush+branch%3Amaster)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Will this `bluge.Document` match this `bluge.Query`?  This library allows you to efficiently answer this question.

```
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
```

## Details

- This implementation is NOT thread-safe.  If you need to run multiple queries concurrently, you must use separate Sour instances.
- Loading stored fields is not supported

## Approach

- Single document, data sizes are small.
- Therefore, avoid heavy document analysis and complex data structures.
- After regular document analysis is complete, use this structure in place.
- Do not build more complicated structures like vellums or roaring bitmaps.
- If additional structure is needed, prefer arrays which have good cache locality, and can be reused.
- Avoid copying data, prefer sub-slicing, and brute-force processing over arrays.
- Cache reusable parts of the query, as we expect the same query to be run over multiple documents.

## License

Apache License Version 2.0