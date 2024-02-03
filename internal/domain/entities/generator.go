package entities

type (
	TypeDefinition struct {
		Name          string
		Initial       string
		Fields        []Field
		HasReferences bool
	}

	Field struct {
		Name     string
		Initial  string
		TypeName string
		JSONName string
		ViewType string
		View     string

		IsReference       bool
		ReferenceName     string
		ReferenceJSONTags []string
	}
)
