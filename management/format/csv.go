// Package format provides interfaces to format content into various kinds of
// data
package format

// CSVFormattable is implemented with the method FormatCSV, which must return the ordered
// slice of JSON struct tag names for the type implmenting it
type CSVFormattable interface {
	FormatCSV() []string
}
