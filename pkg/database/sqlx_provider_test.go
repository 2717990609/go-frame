package database

import (
	"testing"
)

func TestSQLxProvider_Registration(t *testing.T) {
	if Registry["mysql-sqlx"] == nil {
		t.Fatal("mysql-sqlx driver not registered")
	}
}

func TestSQLxProvider_New(t *testing.T) {
	prov, err := New(Config{
		Driver:   "mysql-sqlx",
		Host:     "localhost",
		Port:     3306,
		Database: "test",
		Username: "root",
		Password: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if prov == nil {
		t.Fatal("provider is nil")
	}
}

func TestBuildSelectSQL_InCondition(t *testing.T) {
	sql, args := BuildSelectSQL("users", Query{
		Where: Conditions{In("status", []int{1, 2, 3})},
	})
	if sql == "" {
		t.Fatal("expected non-empty sql")
	}
	if len(args) != 3 {
		t.Errorf("IN expanded args=%v", args)
	}
}
