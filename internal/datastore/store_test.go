package datastore

import (
	"testing"
)

func TestSetAndGet(t *testing.T) {
	store := NewStore(nil) // Pass nil since we're not testing AOF here

	store.Set("foo", "bar", 0)

	val, ok := store.Get("foo")
	if !ok || val != "bar" {
		t.Errorf("Expected 'bar', got '%s'", val)
	}
}
