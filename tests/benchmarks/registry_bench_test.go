package benchmarks_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/ivannovak/glide/v3/pkg/registry"
)

// testItem is a simple struct for benchmarking
type testItem struct {
	Name  string
	Value int
}

// BenchmarkRegistryNew benchmarks creating a new registry
func BenchmarkRegistryNew(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = registry.New[testItem]()
	}
}

// BenchmarkRegistryRegister benchmarks registering items
func BenchmarkRegistryRegister(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := registry.New[testItem]()
		_ = r.Register("item", testItem{Name: "test", Value: 1})
	}
}

// BenchmarkRegistryRegisterWithAliases benchmarks registering items with aliases
func BenchmarkRegistryRegisterWithAliases(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := registry.New[testItem]()
		_ = r.Register("item", testItem{Name: "test", Value: 1}, "alias1", "alias2", "alias3")
	}
}

// BenchmarkRegistryGet benchmarks retrieving items by name
func BenchmarkRegistryGet(b *testing.B) {
	r := registry.New[testItem]()
	_ = r.Register("item", testItem{Name: "test", Value: 1})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Get("item")
	}
}

// BenchmarkRegistryGetByAlias benchmarks retrieving items by alias
func BenchmarkRegistryGetByAlias(b *testing.B) {
	r := registry.New[testItem]()
	_ = r.Register("item", testItem{Name: "test", Value: 1}, "alias1", "alias2", "alias3")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Get("alias2")
	}
}

// BenchmarkRegistryGetMiss benchmarks retrieving non-existent items
func BenchmarkRegistryGetMiss(b *testing.B) {
	r := registry.New[testItem]()
	_ = r.Register("item", testItem{Name: "test", Value: 1})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Get("nonexistent")
	}
}

// BenchmarkRegistryHas benchmarks checking item existence
func BenchmarkRegistryHas(b *testing.B) {
	r := registry.New[testItem]()
	_ = r.Register("item", testItem{Name: "test", Value: 1})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = r.Has("item")
	}
}

// BenchmarkRegistryList benchmarks listing all items
func BenchmarkRegistryList(b *testing.B) {
	r := registry.New[testItem]()
	for j := 0; j < 100; j++ {
		_ = r.Register("item"+strconv.Itoa(j), testItem{Name: "test", Value: j})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = r.List()
	}
}

// BenchmarkRegistryListNames benchmarks listing all item names
func BenchmarkRegistryListNames(b *testing.B) {
	r := registry.New[testItem]()
	for j := 0; j < 100; j++ {
		_ = r.Register("item"+strconv.Itoa(j), testItem{Name: "test", Value: j})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = r.ListNames()
	}
}

// BenchmarkRegistryRemove benchmarks removing items
func BenchmarkRegistryRemove(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := registry.New[testItem]()
		_ = r.Register("item", testItem{Name: "test", Value: 1})
		_ = r.Remove("item")
	}
}

// BenchmarkRegistryCount benchmarks counting items
func BenchmarkRegistryCount(b *testing.B) {
	r := registry.New[testItem]()
	for j := 0; j < 100; j++ {
		_ = r.Register("item"+strconv.Itoa(j), testItem{Name: "test", Value: j})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = r.Count()
	}
}

// BenchmarkRegistrySmall benchmarks a small registry (10 items)
func BenchmarkRegistrySmall(b *testing.B) {
	r := registry.New[testItem]()
	for j := 0; j < 10; j++ {
		_ = r.Register("item"+strconv.Itoa(j), testItem{Name: "test", Value: j})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Get("item5")
	}
}

// BenchmarkRegistryMedium benchmarks a medium registry (100 items)
func BenchmarkRegistryMedium(b *testing.B) {
	r := registry.New[testItem]()
	for j := 0; j < 100; j++ {
		_ = r.Register("item"+strconv.Itoa(j), testItem{Name: "test", Value: j})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Get("item50")
	}
}

// BenchmarkRegistryLarge benchmarks a large registry (1000 items)
func BenchmarkRegistryLarge(b *testing.B) {
	r := registry.New[testItem]()
	for j := 0; j < 1000; j++ {
		_ = r.Register("item"+strconv.Itoa(j), testItem{Name: "test", Value: j})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Get("item500")
	}
}

// BenchmarkRegistryConcurrentReads benchmarks concurrent read access
func BenchmarkRegistryConcurrentReads(b *testing.B) {
	r := registry.New[testItem]()
	for j := 0; j < 100; j++ {
		_ = r.Register("item"+strconv.Itoa(j), testItem{Name: "test", Value: j})
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = r.Get("item" + strconv.Itoa(i%100))
			i++
		}
	})
}

// BenchmarkRegistryConcurrentMixed benchmarks concurrent read/write access
func BenchmarkRegistryConcurrentMixed(b *testing.B) {
	r := registry.New[testItem]()
	for j := 0; j < 100; j++ {
		_ = r.Register("item"+strconv.Itoa(j), testItem{Name: "test", Value: j})
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				// 10% writes
				name := fmt.Sprintf("newitem%d", i)
				_ = r.Register(name, testItem{Name: name, Value: i})
			} else {
				// 90% reads
				_, _ = r.Get("item" + strconv.Itoa(i%100))
			}
			i++
		}
	})
}

// BenchmarkRegistryAllocation measures allocations for registry operations
func BenchmarkRegistryAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r := registry.New[testItem]()
		_ = r.Register("item1", testItem{Name: "test1", Value: 1}, "alias1")
		_ = r.Register("item2", testItem{Name: "test2", Value: 2}, "alias2")
		_, _ = r.Get("item1")
		_, _ = r.Get("alias2")
		_ = r.ListNames()
	}
}
