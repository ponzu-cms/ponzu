package interfaces

// CSVFormattable is implemented with the method FormatCSV, which must return the ordered
// slice of JSON struct tag names for the type implementing it
type CSVFormattable interface {
	FormatCSV() []string
}
