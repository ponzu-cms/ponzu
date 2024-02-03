package root

// FindByTarget returns a set of content based on the targets / identifiers
// provided in Ponzu target string format: Type:ID
// NOTE: All targets should be of the same type
func (repo *repository) FindByTarget(targets []string) ([]interface{}, error) {
	var contents []interface{}
	for i := range targets {
		entity, err := repo.FindOneByTarget(targets[i])
		if err != nil {
			return nil, err
		}

		contents = append(contents, entity)
	}

	return contents, nil
}
