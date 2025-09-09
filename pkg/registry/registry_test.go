package registry

import (
	"fmt"
	"sync"
	"testing"
)

func TestRegistry_Register(t *testing.T) {
	r := New[string]()

	// Test successful registration
	err := r.Register("test", "value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test duplicate registration
	err = r.Register("test", "value2")
	if err == nil {
		t.Error("expected error for duplicate registration")
	}

	// Test registration with aliases
	err = r.Register("item", "value", "alias1", "alias2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test alias conflict
	err = r.Register("item2", "value3", "alias1")
	if err == nil {
		t.Error("expected error for alias conflict")
	}

	// Test name conflict with alias
	err = r.Register("alias2", "value4")
	if err == nil {
		t.Error("expected error for name conflicting with existing alias")
	}
}

func TestRegistry_Get(t *testing.T) {
	r := New[string]()
	r.Register("item1", "value1", "alias1")
	r.Register("item2", "value2")

	// Test get by name
	val, ok := r.Get("item1")
	if !ok || val != "value1" {
		t.Errorf("expected value1, got %v, %v", val, ok)
	}

	// Test get by alias
	val, ok = r.Get("alias1")
	if !ok || val != "value1" {
		t.Errorf("expected value1 via alias, got %v, %v", val, ok)
	}

	// Test get non-existent
	val, ok = r.Get("nonexistent")
	if ok {
		t.Error("expected false for non-existent item")
	}
}

func TestRegistry_MustGet(t *testing.T) {
	r := New[string]()
	r.Register("item", "value")

	// Test successful must get
	val := r.MustGet("item")
	if val != "value" {
		t.Errorf("expected value, got %v", val)
	}

	// Test panic on non-existent
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-existent item")
		}
	}()
	r.MustGet("nonexistent")
}

func TestRegistry_Has(t *testing.T) {
	r := New[string]()
	r.Register("item", "value", "alias")

	if !r.Has("item") {
		t.Error("expected true for existing item")
	}

	if !r.Has("alias") {
		t.Error("expected true for existing alias")
	}

	if r.Has("nonexistent") {
		t.Error("expected false for non-existent item")
	}
}

func TestRegistry_List(t *testing.T) {
	r := New[string]()
	r.Register("item1", "value1")
	r.Register("item2", "value2")

	items := r.List()
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	// Check values exist (order not guaranteed)
	found := make(map[string]bool)
	for _, item := range items {
		found[item] = true
	}
	if !found["value1"] || !found["value2"] {
		t.Error("missing expected values")
	}
}

func TestRegistry_Remove(t *testing.T) {
	r := New[string]()
	r.Register("item", "value", "alias1", "alias2")

	// Remove by name
	if !r.Remove("item") {
		t.Error("expected true for successful removal")
	}

	// Check item and aliases are gone
	if r.Has("item") || r.Has("alias1") || r.Has("alias2") {
		t.Error("item or aliases still exist after removal")
	}

	// Test removing non-existent
	if r.Remove("nonexistent") {
		t.Error("expected false for non-existent item")
	}
}

func TestRegistry_RemoveByAlias(t *testing.T) {
	r := New[string]()
	r.Register("item", "value", "alias")

	// Remove by alias
	if !r.Remove("alias") {
		t.Error("expected true for successful removal via alias")
	}

	// Check item and alias are gone
	if r.Has("item") || r.Has("alias") {
		t.Error("item or alias still exists after removal")
	}
}

func TestRegistry_ResolveAlias(t *testing.T) {
	r := New[string]()
	r.Register("item", "value", "alias")

	canonical, ok := r.ResolveAlias("alias")
	if !ok || canonical != "item" {
		t.Errorf("expected item, got %v, %v", canonical, ok)
	}

	canonical, ok = r.ResolveAlias("item")
	if ok {
		t.Error("expected false for non-alias")
	}
}

func TestRegistry_GetAliases(t *testing.T) {
	r := New[string]()
	r.Register("item", "value", "alias1", "alias2")

	aliases := r.GetAliases("item")
	if len(aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(aliases))
	}

	// Check aliases exist (order not guaranteed)
	found := make(map[string]bool)
	for _, alias := range aliases {
		found[alias] = true
	}
	if !found["alias1"] || !found["alias2"] {
		t.Error("missing expected aliases")
	}

	// Test non-existent item
	aliases = r.GetAliases("nonexistent")
	if aliases != nil {
		t.Error("expected nil for non-existent item")
	}
}

func TestRegistry_IsAlias(t *testing.T) {
	r := New[string]()
	r.Register("item", "value", "alias")

	if !r.IsAlias("alias") {
		t.Error("expected true for alias")
	}

	if r.IsAlias("item") {
		t.Error("expected false for non-alias")
	}

	if r.IsAlias("nonexistent") {
		t.Error("expected false for non-existent")
	}
}

func TestRegistry_Clear(t *testing.T) {
	r := New[string]()
	r.Register("item1", "value1")
	r.Register("item2", "value2", "alias")

	r.Clear()

	if r.Count() != 0 {
		t.Errorf("expected 0 items after clear, got %d", r.Count())
	}

	if r.Has("item1") || r.Has("item2") || r.Has("alias") {
		t.Error("items or aliases still exist after clear")
	}
}

func TestRegistry_ForEach(t *testing.T) {
	r := New[string]()
	r.Register("item1", "value1")
	r.Register("item2", "value2")

	visited := make(map[string]string)
	r.ForEach(func(name string, item string) {
		visited[name] = item
	})

	if len(visited) != 2 {
		t.Errorf("expected 2 items visited, got %d", len(visited))
	}

	if visited["item1"] != "value1" || visited["item2"] != "value2" {
		t.Error("incorrect values in ForEach")
	}
}

func TestRegistry_Filter(t *testing.T) {
	r := New[int]()
	r.Register("item1", 10)
	r.Register("item2", 20)
	r.Register("item3", 30)

	filtered := r.Filter(func(name string, item int) bool {
		return item > 15
	})

	if len(filtered) != 2 {
		t.Errorf("expected 2 filtered items, got %d", len(filtered))
	}

	// Check values (order not guaranteed)
	found := make(map[int]bool)
	for _, item := range filtered {
		found[item] = true
	}
	if !found[20] || !found[30] {
		t.Error("incorrect filtered values")
	}
}

func TestRegistry_ThreadSafety(t *testing.T) {
	r := New[int]()
	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				name := fmt.Sprintf("item_%d_%d", id, j)
				r.Register(name, id*1000+j)

				if val, ok := r.Get(name); ok {
					if val != id*1000+j {
						t.Errorf("incorrect value: expected %d, got %d", id*1000+j, val)
					}
				}

				if j%2 == 0 {
					r.Remove(name)
				}
			}
		}(i)
	}

	wg.Wait()
}
