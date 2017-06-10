title: Format Package Interfaces

Ponzu provides a set of interfaces from the `management/format` package which 
determine how content data should be converted and formatted for exporting via
the Admin interface.

---

## Interfaces

### [format.CSVFormattable](https://godoc.org/github.com/ponzu-cms/ponzu/management/format#CSVFormattable)

CSVFormattable controls if an "Export" button is added to the contents view for 
a Content type in the CMS to export the data to CSV. If it is implemented, a
button will be present beneath the "New" button per Content type. 

##### Method Set

```go
type CSVFormattable interface {
    FormatCSV() []string
}
```

##### Implementation

```go
func (p *Post) FormatCSV() []string {
    // []string contains the JSON struct tags generated for your Content type 
    // implementing the interface
    return []string{
        "id",
        "timestamp",
        "slug",
        "title",
        "photos",
        "body",
        "written_by",
    }
}
```

!!! note "FormatCSV() []string"
    Just like other Ponzu content extension interfaces, like `Push()`, you will 
    return the JSON struct tags for the fields you want exported to the CSV file. 
    These will also be the "header" row in the CSV file to give titles to the file
    columns. Keep in mind that all of item.Item's fields are available here as well.

