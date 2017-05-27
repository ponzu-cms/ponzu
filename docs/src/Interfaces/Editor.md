title: Editor Package Interfaces

Ponzu provides a set of interfaces from the `management/editor` package which 
extend the system's functionality and determine how content editors are rendered 
within the CMS.

---

## Interfaces

### [editor.Editable](https://godoc.org/github.com/ponzu-cms/ponzu/management/editor#Editable)

Editable determines what `[]bytes` are rendered inside the editor page. Use 
Edtiable on anything inside your CMS that you want to provide configuration, editable 
fields, or any HTML/markup to display to an end-user.

!!! note "Implementing `editor.Editable`"
    Most of the time, Ponzu developers generate the majority of this code using 
    the Ponzu CLI [`generate` command](/CLI/Usage).

##### Method Set

```go
type Editable interface {
    MarshalEditor() ([]byte, error)
}
```

##### Implementation

```go
func (p *Post) MarshalEditor() ([]byte, error) {
    // The editor.Form func sets up a structured UI with default styles and form
    // elements based on the fields provided. Most often, Ponzu developers will
    // have the `$ ponzu generate` command generate the MarshalEditor func and 
    // its internal form fields
    view, err := editor.Form(p,
		editor.Field{
			View: editor.Input("Name", p, map[string]string{
				"label":       "Name",
				"type":        "text",
				"placeholder": "Enter the Name here",
			}),
		},
    )
}
```

!!! note "MarshalEditor() & View Rendering"
    Although it is common to use the `editor.Form` and `editor.Fields` to structure your content editor inside `MarshalEditor()`, the method signature defines that its return value needs only to be `[]byte, error`. Keep in mind that you can return a `[]byte` of any raw HTML or other markup to be rendered in the editor view.

---

### [editor.Mergeable](https://godoc.org/github.com/ponzu-cms/ponzu/management/editor#Mergeable)

Mergable enables a CMS end-user to merge the "Pending" content from an outside source into the "Public" section, and thus making it visible via the public content API. It also allows the end-user to reject content. "Approve" and "Reject" buttons will be visible on the edit page for content submitted.

##### Method Set
```go
type Mergeable interface {
    Approve(http.ResponseWriter, *http.Request) error
}
```

##### Example
```go
func (p *Post) Approve(res http.ResponseWriter, req *http.Request) error {
    return nil
}
```
