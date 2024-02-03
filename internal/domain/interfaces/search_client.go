package interfaces

type SearchClientInterface interface {
	// CreateIndex implementation must take into consideration creating a search schema for the target search
	// engine. Because we do not want to statically define directly the schema for item types, SearchClients must use
	// reflection utils as necessary to determine how to dynamically define schemas and how to search them. The most
	// important to note is that Item exposes searchable attributes which may be used to define schemas. By default,
	// ponzu Item will only expose item id and string fields as searchable. To extend this functionality extend
	// content's Searchable interface to expose types you want to search, and implement the appropriate schema
	// in your target search client
	CreateIndex(entityName string, entityType interface{}) error
	UpdateIndex(entityName string, entityType interface{}) error
	Indexes() (map[string]SearchIndexInterface, error)
	GetIndex(entityName string) (SearchIndexInterface, error)
}
