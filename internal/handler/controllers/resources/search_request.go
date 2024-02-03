package resources

type SearchRequestDto struct {
	Query  string `json:"query"`
	Count  int    `json:"count"`
	Offset int    `json:"offset"`
}
