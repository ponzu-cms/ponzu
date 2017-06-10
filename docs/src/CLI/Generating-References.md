title: How to Generate References using Ponzu CLI

In Ponzu, users make connections between Content types using references. In order 
to use the CLI to generate these references, a slightly different syntax is required. 
In all cases, the Content type you wish to reference does not need to exist prior
to the "parent" type referencing it at generate-time, but in the following examples,
the referenced "child" type will be shown before the parent type for clarity.

## Syntax

### @

The **@** symbol is used to declare that the following name is a reference. The 
CLI will take care to parse the name and treat it as a Content type to which the 
current type refers.

### []

The `[]`, which if used, is always in front of the **@** symbol. It signifies 
that the reference type is a slice or a collection of references. When `[]`
is used, the CLI will automatically generate a `reference.SelectRepeater()` view 
for you.

### ,arg1,arg2,argN

Immediately following the reference name (after the @ symbol), users may optionally
pass arguments to specify how the reference is displayed in the parent type's
editor. References are included in the parent types editor as a dropdown menu, with
each possible reference as an option. These arguments define what goes inside the
`<option></option>` text node, as would be seen by an Admin.

The arguments must be valid JSON struct tag names from the reference type's fields. 
Notice in the example below, the `title` and `price` are formatted exactly as they 
were in the generate command for the `product` type.

---
###

##### Example

```bash
$ ponzu gen content product title:string price:int description:string:textarea
$ ponzu gen content catalog year:int products:"[]@product",title,price
```

The commands above output the following. For demonstration, we will omit the full
code generated for the `Product`, since the reference is in the `Catalog` type.

```go
// content/product.go
package content
...

type Product struct {
	item.Item

	Title       string `json:"title"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

...
```

```go
package content

import (
	"fmt"

	"github.com/bosssauce/reference"

	"github.com/ponzu-cms/ponzu/management/editor"
	"github.com/ponzu-cms/ponzu/system/item"
)

type Catalog struct {
	item.Item

	Year     int      `json:"year"`
    // all references are stored as []string or string types
	Products []string `json:"products"` 
}

func (c *Catalog) MarshalEditor() ([]byte, error) {
	view, err := editor.Form(c,
		editor.Field{
			View: editor.Input("Year", c, map[string]string{
				"label":       "Year",
				"type":        "text",
				"placeholder": "Enter the Year here",
			}),
		},
		editor.Field{
            // reference.SelectRepeater since []@product was used
			View: reference.SelectRepeater("Products", c, map[string]string{
				"label": "Products",
			},
				"Product", // generated from @product
				`{{ .title }} {{ .price }} `, // generated from ,title,price args
			),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to render Catalog editor view: %s", err.Error())
	}

	return view, nil
}

func init() {
	item.Types["Catalog"] = func() interface{} { return new(Catalog) }
}
```

**Note:**
If the reference should be only a single item, rather than a slice (or collection)
of items, omit the `[]`, changing the command to:

```bash
$ ponzu gen content catalog year:int product:@product,title,price
```
