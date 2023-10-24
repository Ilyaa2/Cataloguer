package sqlstore

import (
	"Cataloguer/cmd/TestConfig"
	"testing"
)

func TestSqlStore(t *testing.T) (*Sqlstore, *TestConfig.Config) {
	t.Helper()
	config := TestConfig.NewConfig()
	myStore, err := New(config.StoreUrl)
	if err != nil {
		t.Fatal(err)
	}
	return myStore, config
}
