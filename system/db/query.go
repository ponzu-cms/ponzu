package db

// Set inserts or updates values in the database.
// The `key` argument is a string made up of namespace:id (string:int)
func Set(key string) error {

	return nil
}

// Get retrives one item from the database. Non-existent values will return an empty []byte
// The `key` argument is a string made up of namespace:id (string:int)
func Get(key string) []byte {

	return nil
}

// GetAll retrives all items from the database within the provided namespace
func GetAll(namespace string) [][]byte {

	return nil
}
