package pools

import (
	"bytes"
	"sync"
	"testing"
)

const testValue = "value"

func TestBufferPool(t *testing.T) {
	pool := NewBufferPool(1024)

	// Test Get
	buf := pool.Get()
	if buf == nil {
		t.Fatal("Get() returned nil buffer")
	}

	// Buffer should be empty and have capacity
	if buf.Len() != 0 {
		t.Error("New buffer should be empty")
	}
	if buf.Cap() < 1024 {
		t.Error("Buffer should have at least initial capacity")
	}

	// Test buffer usage
	buf.WriteString("test data")
	if buf.String() != "test data" {
		t.Error("Buffer should contain written data")
	}

	// Test Put
	pool.Put(buf)

	// Get another buffer - should be the same one, reset
	buf2 := pool.Get()
	if buf2.Len() != 0 {
		t.Error("Reused buffer should be reset")
	}
}

func TestBufferPoolCapacityLimit(t *testing.T) {
	pool := NewBufferPool(100)

	buf := pool.Get()
	// Make buffer grow beyond 4x initial size
	largeData := make([]byte, 500) // 5x initial size
	buf.Write(largeData)

	if buf.Cap() < 500 {
		t.Error("Buffer should have grown to accommodate data")
	}

	// Put back - should be rejected due to size
	pool.Put(buf)

	// Get new buffer - should be fresh (not the oversized one)
	buf2 := pool.Get()
	if buf2.Cap() >= 500 {
		t.Error("Oversized buffer should not have been reused")
	}
}

func TestStringSlicePool(t *testing.T) {
	pool := NewStringSlicePool(10)

	// Test Get
	slice := pool.Get()
	if slice == nil {
		t.Fatal("Get() returned nil slice")
	}
	if len(slice) != 0 {
		t.Error("New slice should have zero length")
	}
	if cap(slice) < 10 {
		t.Error("Slice should have at least initial capacity")
	}

	// Test slice usage
	slice = append(slice, "field1", "field2", "field3")
	if len(slice) != 3 {
		t.Error("Slice should contain appended data")
	}

	// Test Put
	pool.Put(slice)

	// Get another slice - should be reset but keep capacity
	slice2 := pool.Get()
	if len(slice2) != 0 {
		t.Error("Reused slice should have zero length")
	}
	if cap(slice2) < 10 {
		t.Error("Reused slice should maintain capacity")
	}
}

func TestStringSlicePoolCapacityLimit(t *testing.T) {
	pool := NewStringSlicePool(5)

	slice := pool.Get()
	// Make slice grow beyond 4x initial capacity
	for i := 0; i < 25; i++ { // 5x initial capacity
		slice = append(slice, "field")
	}

	if cap(slice) < 25 {
		t.Error("Slice should have grown to accommodate data")
	}

	// Put back - should be rejected due to size
	pool.Put(slice)

	// Get new slice - should be fresh (not the oversized one)
	slice2 := pool.Get()
	if cap(slice2) >= 25 {
		t.Error("Oversized slice should not have been reused")
	}
}

func TestMapPool(t *testing.T) {
	pool := NewMapPool()

	// Test Get
	m := pool.Get()
	if m == nil {
		t.Fatal("Get() returned nil map")
	}
	if len(m) != 0 {
		t.Error("New map should be empty")
	}

	// Test map usage
	m["key1"] = "value1"
	m["key2"] = "value2"
	if len(m) != 2 {
		t.Error("Map should contain added data")
	}

	// Test Put
	pool.Put(m)

	// Get another map - should be cleared
	m2 := pool.Get()
	if len(m2) != 0 {
		t.Error("Reused map should be empty")
	}

	// Should be able to use the map normally
	m2["test"] = testValue
	if m2["test"] != testValue {
		t.Error("Reused map should work normally")
	}
}

func TestMapPoolSizeLimit(t *testing.T) {
	pool := NewMapPool()

	m := pool.Get()
	// Add more than 100 entries
	for i := 0; i < 150; i++ {
		m[string(rune(i))] = testValue
	}

	if len(m) != 150 {
		t.Error("Map should contain all added entries")
	}

	// Put back - should be rejected due to size
	pool.Put(m)

	// Get new map - should be fresh (not the oversized one)
	m2 := pool.Get()
	if len(m2) != 0 {
		t.Error("Map should be fresh and empty")
	}
}

func TestGlobalPools(t *testing.T) {
	// Test helper functions
	buf := GetSmallBuffer()
	if buf == nil {
		t.Error("GetSmallBuffer() returned nil")
	}
	PutSmallBuffer(buf)

	buf = GetLargeBuffer()
	if buf == nil {
		t.Error("GetLargeBuffer() returned nil")
	}
	PutLargeBuffer(buf)

	fields := GetCSVFields()
	if fields == nil {
		t.Error("GetCSVFields() returned nil")
	}
	PutCSVFields(fields)

	record := GetCSVRecord()
	if record == nil {
		t.Error("GetCSVRecord() returned nil")
	}
	PutCSVRecord(record)
}

func TestStatsBufferPool(t *testing.T) {
	pool := NewStatsBufferPool(1024)

	// Check initial stats
	stats := pool.GetStats()
	if stats.Hits != 0 || stats.Puts != 0 {
		t.Error("Initial stats should be zero")
	}

	// Get a buffer
	buf := pool.Get()
	if buf == nil {
		t.Error("Get() returned nil")
	}

	// Check stats after get
	stats = pool.GetStats()
	if stats.Hits != 1 {
		t.Error("Hits should be 1 after Get()")
	}

	// Put buffer back
	pool.Put(buf)

	// Check stats after put
	stats = pool.GetStats()
	if stats.Puts != 1 {
		t.Error("Puts should be 1 after Put()")
	}

	// Reset stats
	pool.ResetStats()
	stats = pool.GetStats()
	if stats.Hits != 0 || stats.Puts != 0 {
		t.Error("Stats should be zero after reset")
	}
}

func TestPooledCSVParser(t *testing.T) {
	parser := NewPooledCSVParser()

	headers := []string{"id", "name", "type"}
	fields := []string{"123", "Test Station", "0"}

	// Test ParseRecord
	record := parser.ParseRecord(fields, headers)
	if record == nil {
		t.Fatal("ParseRecord returned nil")
	}

	if record["id"] != "123" {
		t.Error("Record should contain correct id")
	}
	if record["name"] != "Test Station" {
		t.Error("Record should contain correct name")
	}
	if record["type"] != "0" {
		t.Error("Record should contain correct type")
	}

	// Test ReturnRecord
	parser.ReturnRecord(record)

	// Test buffer operations
	buf := parser.GetBuffer()
	if buf == nil {
		t.Error("GetBuffer returned nil")
	}
	parser.ReturnBuffer(buf)

	// Test field operations
	fieldSlice := parser.GetFields()
	if fieldSlice == nil {
		t.Error("GetFields returned nil")
	}
	parser.ReturnFields(fieldSlice)
}

func TestPooledCSVParserMismatchedFields(t *testing.T) {
	parser := NewPooledCSVParser()

	headers := []string{"id", "name", "type"}
	fields := []string{"123", "Test Station"} // Missing type field

	record := parser.ParseRecord(fields, headers)
	if record != nil {
		t.Error("ParseRecord should return nil for mismatched field count")
	}
}

func TestConcurrentPoolAccess(t *testing.T) {
	pool := NewBufferPool(1024)

	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				buf := pool.Get()
				buf.WriteString("test data")
				pool.Put(buf)
			}
		}()
	}

	wg.Wait()
	// Test passes if no race conditions occur
}

func TestConcurrentStatsPoolAccess(t *testing.T) {
	pool := NewStatsBufferPool(1024)

	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				buf := pool.Get()
				buf.WriteString("test data")
				pool.Put(buf)

				// Occasionally check/reset stats
				if j%10 == 0 {
					_ = pool.GetStats()
				}
				if j%50 == 0 {
					pool.ResetStats()
				}
			}
		}()
	}

	wg.Wait()

	// Verify final stats make sense
	stats := pool.GetStats()
	if stats.Hits < 0 || stats.Puts < 0 {
		t.Error("Stats should not be negative")
	}
}

// Benchmark tests
func BenchmarkBufferPoolGet(b *testing.B) {
	pool := NewBufferPool(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		pool.Put(buf)
	}
}

func BenchmarkBufferPoolGetPut(b *testing.B) {
	pool := NewBufferPool(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get()
		buf.WriteString("test data for benchmarking")
		pool.Put(buf)
	}
}

func BenchmarkDirectBufferAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		buf.WriteString("test data for benchmarking")
		// No pool.Put() equivalent
	}
}

func BenchmarkStringSlicePool(b *testing.B) {
	pool := NewStringSlicePool(32)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slice := pool.Get()
		slice = append(slice, "field1", "field2", "field3", "field4", "field5")
		pool.Put(slice)
	}
}

func BenchmarkMapPool(b *testing.B) {
	pool := NewMapPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := pool.Get()
		m["id"] = "123"
		m["name"] = "Test Station"
		m["type"] = "0"
		pool.Put(m)
	}
}

func BenchmarkPooledCSVParser(b *testing.B) {
	parser := NewPooledCSVParser()
	headers := []string{"id", "name", "type", "lat", "lon"}
	fields := []string{"123", "Test Station", "0", "37.7749", "-122.4194"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		record := parser.ParseRecord(fields, headers)
		parser.ReturnRecord(record)
	}
}
