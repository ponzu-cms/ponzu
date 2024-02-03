package search

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/fanky5g/ponzu/internal/domain/interfaces"
	"reflect"
	"strings"
)

type Index struct {
	Name              string
	idx               bleve.Index
	contentRepository interfaces.ContentRepositoryInterface
}

func (index *Index) Key(id string) string {
	if !strings.HasPrefix(id, index.Name) {
		id = fmt.Sprintf("%s:%s", index.Name, id)
	}

	return id
}

func (index *Index) Update(id string, data interface{}) error {
	entity, ok := data.(interfaces.Searchable)
	if !ok {
		return errors.New("entity does not implement searchable interface")
	}

	if !entity.IndexContent() {
		return nil
	}

	content := make(map[string]string)
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	content[TypeField] = index.Name
	for fieldName := range entity.GetSearchableAttributes() {
		field := v.FieldByName(fieldName)
		switch field.Kind() {
		case reflect.String:
			if field.IsValid() && !field.IsZero() {
				// TODO: use json tag(annotation) for field name
				content[fieldName] = field.String()
			}
		default:
			return fmt.Errorf("%s type %s is not supported in search", fieldName, field.Kind())
		}
	}

	return index.idx.Index(index.Key(id), content)
}

func (index *Index) Delete(id string) error {
	return index.idx.Delete(index.Key(id))
}

func (index *Index) Search(query string, count, offset int) ([]interface{}, error) {
	q := bleve.NewQueryStringQuery(fmt.Sprintf("+%s:%s +%s", TypeField, index.Name, query))
	searchRequest := bleve.NewSearchRequestOptions(q, count, offset, false)
	res, err := index.idx.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, hit := range res.Hits {
		results = append(results, hit.ID)
	}

	// since we only index searchable fields. We need to fetch the original entities
	return index.contentRepository.FindByTarget(results)
}

func NewSearchIndex(name string, index bleve.Index, contentRepository interfaces.ContentRepositoryInterface) (interfaces.SearchIndexInterface, error) {
	return &Index{
		Name:              name,
		idx:               index,
		contentRepository: contentRepository,
	}, nil
}
