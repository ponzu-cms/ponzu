package interfaces

type Cacheable interface {
	Cache() Cache
	InvalidateCache() error
}
