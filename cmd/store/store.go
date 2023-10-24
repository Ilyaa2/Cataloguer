package store

type Store interface {
	User() UserRepository
	Message() MessageRepository
}
