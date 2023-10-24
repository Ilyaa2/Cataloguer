package cache

type SessionRepository interface {
	GetValue(key string) (string, error)
	SetValue(key string, value string, ex int) error
	DeleteValue(key string)
}
