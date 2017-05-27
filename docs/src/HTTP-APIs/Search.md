title: Full-text Search HTTP API

Ponzu provides a read-only HTTP API to search the contents of your system's database. 
Full-text search is made possible by the use of [Bleve](http://blevesearch.com), 
which handles the indexing and querying. 

---

### Endpoints

#### Search Content

<kbd>GET</kbd> `/api/search?type=<Type>&q=<Query String>`

!!! warning "Search must be enabled individually for each Content type"
    - Search is not on by default to protect your data in case it shouldn't be indexed and published via the API.
    - `SearchMapping()` is implemented with default mapping (ideal for 99% of use cases). 
    - To enable search, add a `IndexContent() bool` method to your content type and return `true` (default implementation returns false).

- `<Type>` must implement [db.Searchable](/Interfaces/Search/#searchsearchable)

- Search is currently limited to single `<Type>` per request

- `<Query String>` documentation here: [Bleve Docs - Query String](http://www.blevesearch.com/docs/Query-String-Query/)

- Search results are formatted exactly the same as standard Content API calls, so you don't need to change your client data model  

- Search handler will respect other interface implementations on your content, including: 
    - [`item.Hideable`](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Hideable)
    - [`item.Omittable`](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Omittable) 
    - [`item.Pushable`](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Pushable) _(Note: only the first search result will be pushed)_

##### Sample Response
```javascript
{
  "data": [
    {
        "uuid": "024a5797-e064-4ee0-abe3-415cb6d3ed18",
        "id": 6,
        "slug": "item-id-024a5797-e064-4ee0-abe3-415cb6d3ed18", // customizable
        "timestamp": 1493926453826, // milliseconds since Unix epoch
        "updated": 1493926453826,
        // your content data...,
    }
  ]
}
```
