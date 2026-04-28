package database

import (
	"testing"
)

func TestBuildSelectSQL(t *testing.T) {
	sql, args := BuildSelectSQL("users", Query{
		Where: Conditions{Eq("email", "a@b.com")},
		Limit: intPtr(10),
	})
	if sql == "" {
		t.Fatal("expected non-empty sql")
	}
	if len(args) != 1 || args[0] != "a@b.com" {
		t.Errorf("args=%v", args)
	}
}

func TestBuildSelectSQL_Order(t *testing.T) {
	sql, args := BuildSelectSQL("users", Query{
		Where: Conditions{Eq("id", 1)},
		Order: []Order{{Column: "created_at", Direction: "desc"}},
	})
	if sql == "" {
		t.Fatal("expected non-empty sql")
	}
	if len(args) != 1 {
		t.Errorf("args=%v", args)
	}
}

func intPtr(n int) *int {
	return &n
}
