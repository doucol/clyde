package cache

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := New[string, int]()

	// Test setting and getting a value
	cache.Set("key1", 42)
	if val, ok := cache.Get("key1"); !ok || val != 42 {
		t.Errorf("Get() = %v, %v; want 42, true", val, ok)
	}

	// Test getting non-existent key
	if val, ok := cache.Get("nonexistent"); ok {
		t.Errorf("Get() = %v, %v; want 0, false", val, ok)
	}
}

func TestCache_SetTTL(t *testing.T) {
	cache := New[string, int]()

	// Test setting with TTL
	cache.SetTTL("key1", 42, 100*time.Millisecond)

	// Value should be available immediately
	if val, ok := cache.Get("key1"); !ok || val != 42 {
		t.Errorf("Get() = %v, %v; want 42, true", val, ok)
	}

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Value should be expired
	if val, ok := cache.Get("key1"); ok {
		t.Errorf("Get() = %v, %v; want 0, false", val, ok)
	}
}

func TestCache_Remove(t *testing.T) {
	cache := New[string, int]()

	// Test removing a value
	cache.Set("key1", 42)
	cache.Remove("key1")

	if val, ok := cache.Get("key1"); ok {
		t.Errorf("Get() = %v, %v; want 0, false", val, ok)
	}
}

func TestCache_Pop(t *testing.T) {
	cache := New[string, int]()

	// Test popping a value
	cache.Set("key1", 42)

	// Pop should return the value and remove it
	if val, ok := cache.Pop("key1"); !ok || val != 42 {
		t.Errorf("Pop() = %v, %v; want 42, true", val, ok)
	}

	// Value should be removed
	if val, ok := cache.Get("key1"); ok {
		t.Errorf("Get() = %v, %v; want 0, false", val, ok)
	}

	// Test popping non-existent key
	if val, ok := cache.Pop("nonexistent"); ok {
		t.Errorf("Pop() = %v, %v; want 0, false", val, ok)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := New[string, int]()

	// Test concurrent access
	done := make(chan bool)

	// Start multiple goroutines that read and write to the cache
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "key" + string(rune(id+'0'))
			cache.Set(key, id)
			time.Sleep(10 * time.Millisecond)
			if val, ok := cache.Get(key); !ok || val != id {
				t.Errorf("Concurrent access failed for key %s", key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
