package cache

type Cache interface {
	Session() SessionRepository
}
