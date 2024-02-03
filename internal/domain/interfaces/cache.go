package interfaces

type Cache interface {
	GetByKey(key string) interface{}
	Warm(value []byte) error
}
