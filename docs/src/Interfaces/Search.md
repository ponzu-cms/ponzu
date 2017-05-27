title: Search Package Interfaces

Ponzu provides a set of interfaces from the `system/search` package to enable and customize full-text search access to content in your system. **Search is not enabled by default**, and must be enabled per Content type individually.

## Interfaces

### [search.Searchable](https://godoc.org/github.com/ponzu-cms/ponzu/system/search#Searchable)
Searchable determines how content is indexed and whether the system should index the content when it is created and updated or be removed from the index when content is deleted.
    
!!! warning ""
    Search is **disabled** for all Content items by default. Each Content item that should be indexed and searchable must implement the `search.Searchable` interface.

##### Method Set

```go
type Searchable interface {
    SearchMapping() (*mapping.IndexMappingImpl, error)
    IndexContent() bool
}
```

By default, Ponzu sets up the [Bleve's](http://blevesearch.com) "default mapping", which is typically what you want for most content-based systems. This can be overridden by implementing your own `SearchMapping() (*mapping.IndexMappingImpl, error)` method on your Content type. 

This way, all you need to do to get full-text search is to add the `IndexContent() bool` method to each Content type you want search enabled. Return `true` from this method to enable search. 


##### Example
```go
// ...

type Song struct {
    item.Item

    Name string `json:"name"`
    // ...
}

func (s *Song) IndexContent() bool {
    return true
}
```

!!! tip "Indexing Existing Content"
    If you previously had search disabled and had already added content to your system, you will need to re-index old content items in your CMS. Otherwise, they will not show up in search queries.. This requires you to manually open each item and click 'Save'. This could be scripted and Ponzu _might_ ship with a re-indexing function at some point in the fututre.
