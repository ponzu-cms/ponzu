title: Content Overview

Nearly everything you work on in Ponzu is inside content files on the content types you create. These types must all reside in the `content` package and are the fundamental core of your CMS. In order for Content types to be rendered and managed by the CMS, they must implement the `editor.Editable` interface, and add their own `interface{}` container to the global `item.Types` map. 

Sound like a lot? Don't worry, all of this can be done for you by using the code-generating command line tools that come with Ponzu.

It is rare to hand-write a new Content type, and should be generated instead!

### Generating Content types

To generate content types and boilerplate code, use the Ponzu CLI `generate` command as such:
```bash
$ ponzu generate content post title:string body:string:richtext author:string
``` 

The command above will create a file at `content/post.go` and will generate the following code:
```go
package content

import (
	"fmt"

	"github.com/ponzu-cms/ponzu/management/editor"
	"github.com/ponzu-cms/ponzu/system/item"
)

type Post struct {
	item.Item

	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

// MarshalEditor writes a buffer of html to edit a Post within the CMS
// and implements editor.Editable
func (p *Post) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(p,
		// Take note that the first argument to these Input-like functions
		// is the string version of each Post field, and must follow
		// this pattern for auto-decoding and auto-encoding reasons:
		editor.Field{
			View: editor.Input("Title", p, map[string]string{
				"label":       "Title",
				"type":        "text",
				"placeholder": "Enter the Title here",
			}),
		},
		editor.Field{
			View: editor.Richtext("Body", p, map[string]string{
				"label":       "Body",
				"placeholder": "Enter the Body here",
			}),
		},
		editor.Field{
			View: editor.Input("Author", p, map[string]string{
				"label":       "Author",
				"type":        "text",
				"placeholder": "Enter the Author here",
			}),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Post editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Post"] = func() interface{} { return new(Post) }
}
```

The code above is the baseline amount required to manage content for the `Post` type from within the CMS. See [Extending Content](/Content/Extending-Content) for information about how to add more functionality to your Content types. 

All content managed by the CMS and exposed via the API is considered an "item", and thus should embed the `item.Item` type. There are many benefits to this, such as becoming automatically sortable by time, and being given default methods that are useful inside and out of the CMS. All content types that are created by the `generate` command via Ponzu CLI will embed Item. 

### Related packages

The `item` package has a number of useful interfaces, which make it simple to add functionality to all content types and other types that embed Item. 

The `editor` package has the Editable interface, which allows types to create an editor for their fields within the CMS. Additionally, there is a helper function `editor.Form` which simplifies defining the editor's input layout and input types using `editor.Input` and various other functions to make HTML input elements like Select, Checkbox, Richtext, Textarea and more.

The `api` package has interfaces including `api.Createable` and `api.Mergeable` which make it trivial to accept and approve or reject content submitted from 3rd parties (POST from HTML forms, mobile clients, etc).

