title: Item Package Interfaces

Ponzu provides a set of interfaces from the `system/item` package which extend 
the functionality of the content in your system and how it interacts with other 
components inside and outside of Ponzu. 

---

## Interfaces

### [item.Pushable](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Pushable)
Pushable, if [HTTP/2 Server Push](https://http2.github.io/http2-spec/#PushResources) 
is supported by the client, can tell a handler which resources it would like to 
have "pushed" preemptively to the client. This saves follow-on roundtrip requests 
for other items which are referenced by the Pushable item. The `Push` method, the 
only method in Pushable, must return a `[]string` containing the `json` field tags 
of the referenced items within the type.

##### Method Set
```go
type Pushable interface {
    // the values contained in []string fields returned by Push must be URL paths
    Push(http.ResponseWriter, *http.Request) ([]string, error)
}
```

##### Implementation
The `Push` method returns a `[]string` containing the `json` tag field names for
which you want to have pushed to a supported client and an error value. The values 
for the field names **must** be URL paths, and cannot be from another origin.

```go
type Post struct {
    item.Item

    HeaderPhoto string `json:"header_photo"`
    Author      string `json:"author"` // reference `/api/content/?type=Author&id=2`
    // ...
}

func (p *Post) Push(res http.ResponseWriter, req *http.Request) ([]string, error) {
    return []string{
        "header_photo",
        "author",
    }, nil
}
```

---

### [item.Hideable](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Hideable)
Hideable tells an API handler that data of this type shouldn’t be exposed outside 
the system. Hideable types cannot be used as references (relations in Content types).
The `Hide` method, the only method in Hideable, takes an `http.ResponseWriter, *http.Request` 
and returns an `error`. A special error in the `items` package, `ErrAllowHiddenItem` 
can be returned as the error from Hide to instruct handlers to show hidden 
content in specific cases.

##### Method Set
```go
type Hideable interface {
    Hide(http.ResponseWriter, *http.Request) error
}
```

##### Implementation
```go
func (p *Post) Hide(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

---

### [item.Omittable](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Omittable)
Omittable tells a content API handler to keep certain fields from being exposed 
through the JSON response. It's single method, `Omit` takes no arguments and 
returns a `[]string` which must be made up of the JSON struct tags for the type 
containing fields to be omitted and an error value.

##### Method Set
```go
type Omittable interface {
    Omit(http.ResponseWriter, *http.Request) ([]string, error)
}
```

##### Implementation
```go
type Post struct {
    item.Item

    HeaderPhoto string `json:"header_photo"`
    Author      string `json:"author"`
    // ...
}

func (p *Post) Omit(res http.ResponseWriter, req *http.Request) ([]string, error) {
    return []string{
        "header_photo",
        "author",
    }, nil
}
```

---

### [item.Hookable](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Hookable)
Hookable provides lifecycle hooks into the http handlers which manage Save, Delete,
Approve, and Reject routines. All methods in its set take an 
`http.ResponseWriter, *http.Request` and return an `error`.

##### Method Set

```go
type Hookable interface {
    BeforeAPICreate(http.ResponseWriter, *http.Request) error
    AfterAPICreate(http.ResponseWriter, *http.Request) error

    BeforeAPIUpdate(http.ResponseWriter, *http.Request) error
    AfterAPIUpdate(http.ResponseWriter, *http.Request) error

    BeforeAPIDelete(http.ResponseWriter, *http.Request) error
    AfterAPIDelete(http.ResponseWriter, *http.Request) error

    BeforeAdminCreate(http.ResponseWriter, *http.Request) error
    AfterAdminCreate(http.ResponseWriter, *http.Request) error

    BeforeAdminUpdate(http.ResponseWriter, *http.Request) error
    AfterAdminUpdate(http.ResponseWriter, *http.Request) error

    BeforeAdminDelete(http.ResponseWriter, *http.Request) error
    AfterAdminDelete(http.ResponseWriter, *http.Request) error

    BeforeSave(http.ResponseWriter, *http.Request) error
    AfterSave(http.ResponseWriter, *http.Request) error

    BeforeDelete(http.ResponseWriter, *http.Request) error
    AfterDelete(http.ResponseWriter, *http.Request) error

    BeforeApprove(http.ResponseWriter, *http.Request) error
    AfterApprove(http.ResponseWriter, *http.Request) error

    BeforeReject(http.ResponseWriter, *http.Request) error
    AfterReject(http.ResponseWriter, *http.Request) error

    // Enable/Disable used exclusively for addons
    BeforeEnable(http.ResponseWriter, *http.Request) error
    AfterEnable(http.ResponseWriter, *http.Request) error

    BeforeDisable(http.ResponseWriter, *http.Request) error
    AfterDisable(http.ResponseWriter, *http.Request) error
}
```

##### Implementations

#### BeforeAPICreate
BeforeAPICreate is called before an item is created via a 3rd-party client. If a 
non-nil `error` value is returned, the item will not be created/saved.

```go
func (p *Post) BeforeAPICreate(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterAPICreate
AfterAPICreate is called after an item has been created via a 3rd-party client.
At this point, the item has been saved to the database. If a non-nil `error` is
returned, it will respond to the client with an empty response, so be sure to 
use the `http.ResponseWriter` from within your hook appropriately.

```go
func (p *Post) AfterAPICreate(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeApprove
BeforeApprove is called before an item is merged as "Public" from its prior 
status as "Pending". If a non-nil `error` value is returned, the item will not be
appproved, and an error message is displayed to the Admin. 

```go
func (p *Post) BeforeApprove(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterApprove
AfterApprove is called after an item has been merged as "Public" from its prior
status as "Pending". If a non-nil `error` is returned, an error message is 
displayed to the Admin, however the item will already be irreversibly merged.

```go
func (p *Post) AfterApprove(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeReject
BeforeReject is called before an item is rejected and deleted by default. To reject
an item, but not delete it, return a non-nil `error` from this hook - doing so 
will allow the hook to do what you want it to do prior to the return, but the item
will remain in the "Pending" section.

```go
func (p *Post) BeforeReject(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterReject
AfterReject is called after an item is rejected and has been deleted.

```go
func (p *Post) AfterReject(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeSave
BeforeSave is called before any CMS Admin or 3rd-party client triggers a save to 
the database. This could be done by clicking the 'Save' button on a Content editor, 
or by a API call to Create or Update the Content item. By returning a non-nil 
`error` value, the item will not be saved.

```go
func (p *Post) BeforeSave(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterSave
AfterSave is called after any CMS Admin or 3rd-party client triggers a save to 
the database. This could be done by clicking the 'Save' button on a Content editor, 
or by a API call to Create or Update the Content item.

```go
func (p *Post) AfterSave(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeDelete
BeforeDelete is called before any CMS Admin or 3rd-party client triggers a delete to 
the database. This could be done by clicking the 'Delete' button on a Content editor, 
or by a API call to Delete the Content item. By returning a non-nil `error` value,
the item will not be deleted.

```go
func (p *Post) BeforeDelete(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterDelete
AfterSave is called after any CMS Admin or 3rd-party client triggers a delete to 
the database. This could be done by clicking the 'Delete' button on a Content editor, 
or by a API call to Delete the Content item.

```go
func (p *Post) AfterDelete(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeAPIDelete
BeforeDelete is only called before a 3rd-party client triggers a delete to the 
database. By returning a non-nil `error` value, the item will not be deleted.

```go
func (p *Post) BeforeAPIDelete(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterAPIDelete
AfterAPIDelete is only called after a 3rd-party client triggers a delete to the 
database.

```go
func (p *Post) AfterAPIDelete(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeAPIUpdate
BeforeAPIUpdate is only called before a 3rd-party client triggers an update to 
the database. By returning a non-nil `error` value, the item will not be updated.

```go
func (p *Post) BeforeAPIUpdate(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterAPIUpdate
AfterAPIUpdate is only called after a 3rd-party client triggers an update to 
the database.

```go
func (p *Post) AfterAPIUpdate(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeAdminCreate
BeforeAdminCreate is only called before a CMS Admin creates a new Content item.
It is not called for subsequent saves to the item once it has been created and 
assigned an ID. By returning a non-nil `error` value, the item will not be created.

```go
func (p *Post) BeforeAdminCreate(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterAdminCreate
AfterAdminCreate is only called after a CMS Admin creates a new Content item.
It is not called for subsequent saves to the item once it has been created and 
assigned an ID.

```go
func (p *Post) AfterAdminCreate(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeAdminUpdate
BeforeAdminUpdate is only called before a CMS Admin updates a Content item. By 
returning a non-nil `error`, the item will not be updated.

```go
func (p *Post) BeforeAdminUpdate(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterAdminUpdate
AfterAdminUpdate is only called after a CMS Admin updates a Content item.

```go
func (p *Post) AfterAdminUpdate(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeAdminDelete
BeforeAdminDelete is only called before a CMS Admin deletes a Content item. By
returning a non-nil `error` value, the item will not be deleted.

```go
func (p *Post) BeforeAdminDelete(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### AfterAdminDelete
AfterAdminDelete is only called after a CMS Admin deletes a Content item.

```go
func (p *Post) AfterAdminDelete(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```

#### BeforeEnable
BeforeEnable is only applicable to Addon items, and is called before the addon
changes status to "Enabled". By returning a non-nil `error` value, the addon
will not become enabled.

```go
func (p *Post) BeforeEnable(http.ResponseWriter, *http.Request) error {
    return nil
}
```

#### AfterEnable
AfterEnable is only applicable to Addon items, and is called after the addon 
changes status to "Enabled".
```go
func (p *Post) AfterEnable(http.ResponseWriter, *http.Request) error {
    return nil
}
```

#### BeforeDisable
BeforeDisable is only applicable to Addon items, and is called before the addon
changes status to "Disabled". By returning a non-nil `error` value, the addon
will not become disabled.
```go
func (p *Post) BeforeDisable(http.ResponseWriter, *http.Request) error {
    return nil
}
```

#### AfterDisable
AfterDisable is only applicable to Addon items, and is called after the addon 
changes status to "Disabled".
```go
func (p *Post) AfterDisable(http.ResponseWriter, *http.Request) error {
    return nil
}
```

Hookable is implemented by Item by default as no-ops which are expected to be overridden. 

!!! note "Note" 
    returning an error from any of these `Hookable` methods will end the request, 
    causing it to halt immediately after the hook. For example, returning an `error` 
    from `BeforeDelete` will result in the content being kept in the database. 
    The same logic applies to all of these interface methods that return an error 
    - **the error defines the behavior**.

---

### [item.Identifiable](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Identifiable)
Identifiable enables a struct to have its ID set/get. Typically this is done to set an ID to -1 indicating it is new for DB inserts, since by default a newly initialized struct would have an ID of 0, the int zero-value, and BoltDB's starting key per bucket is 0, thus overwriting the first record.
Most notable, Identifiable’s `String` method is used to set a meaningful display name for an Item. `String` is called by default in the Admin dashboard to show the Items of certain types, and in the default creation of an Item’s slug.
Identifiable is implemented by Item by default.

##### Method Set
```go
type Identifiable interface {
    ItemID() int
    SetItemID(int)
    UniqueID() uuid.UUID
    String() string
}
```

##### Implementation
`item.Identifiable` has a default implementation in the `system/item` package. 
It is not advised to override these methods, with the exception of `String()`, 
which is commonly used to set the display name of Content items when listed in 
the CMS, and to customize slugs.

```go
func (i Item) ItemID() int {
	return i.ID
}

func (i *Item) SetItemID(id int) {
	i.ID = id
}

func (i Item) UniqueID() uuid.UUID {
	return i.UUID
}

func (i Item) String() string {
	return fmt.Sprintf("Item ID: %s", i.UniqueID())
}
```
---

### [item.Sluggable](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Sluggable)
Sluggable makes a struct locatable by URL with it's own path. As an Item implementing Sluggable, slugs may overlap. If this is an issue, make your content struct (or one which embeds Item) implement Sluggable and it will override the slug created by Item's `SetSlug` method with your own.
It is not recommended to override `SetSlug`, but rather the `String` method on your content struct, which will have a similar, more predictable effect.
Sluggable is implemented by Item by default.

##### Method Set
```go
type Sluggable interface {
    SetSlug(string)
    ItemSlug() string
}
```

##### Implementation
`item.Sluggable` has a default implementation in the `system/item` package. It is
possible to override these methods on your own Content types, but beware, behavior
is undefined. It is tempting to override the `SetSlug()` method to customize your
Content item slug, but try first to override the `String()` method found in the
`item.Identifiable` interface instead. If you don't get the desired results, try
`SetSlug()`.

```go
func (i *Item) SetSlug(slug string) {
	i.Slug = slug
}

func (i *Item) ItemSlug() string {
	return i.Slug
}
```
---


### [item.Sortable](https://godoc.org/github.com/ponzu-cms/ponzu/system/item#Sortable)
Sortable enables items to be sorted by time, as per the sort.Interface interface. Sortable is implemented by Item by default.

##### Method Set
```go
type Sortable interface {
    Time() int64
    Touch() int64
}
```

##### Implementation
`item.Sortable` has a default implementation in the `system/item` package. It is
possible to override these methods on your own Content type, but beware, behavior 
is undefined.

```go
func (i Item) Time() int64 {
	return i.Timestamp
}

func (i Item) Touch() int64 {
	return i.Updated
}
```

