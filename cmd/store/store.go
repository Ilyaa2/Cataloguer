package store

type Store interface {
	User() UserRepository
	//todo Message() MessageRepository
}
