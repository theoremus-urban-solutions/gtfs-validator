// Package pools provides memory pooling functionality to reduce garbage collection
// overhead during GTFS validation, especially for CSV parsing operations
package pools

import (
	"bytes"
	"sync"
)

// BufferPool provides a pool of byte buffers for CSV parsing and other operations
type BufferPool struct {
	pool sync.Pool
	size int
}

// NewBufferPool creates a new buffer pool with buffers of the specified initial size
func NewBufferPool(initialSize int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, initialSize))
			},
		},
		size: initialSize,
	}
}

// Get retrieves a buffer from the pool
func (p *BufferPool) Get() *bytes.Buffer {
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset() // Clear the buffer before use
	return buf
}

// Put returns a buffer to the pool for reuse
func (p *BufferPool) Put(buf *bytes.Buffer) {
	// Only put back buffers that aren't too large to prevent memory bloat
	if buf.Cap() <= p.size*4 { // Allow up to 4x the initial size
		p.pool.Put(buf)
	}
}

// StringSlicePool provides a pool of string slices for CSV row parsing
type StringSlicePool struct {
	pool sync.Pool
	size int
}

// NewStringSlicePool creates a new string slice pool with the specified initial capacity
func NewStringSlicePool(initialCapacity int) *StringSlicePool {
	return &StringSlicePool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]string, 0, initialCapacity)
			},
		},
		size: initialCapacity,
	}
}

// Get retrieves a string slice from the pool
func (p *StringSlicePool) Get() []string {
	slice := p.pool.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// Put returns a string slice to the pool for reuse
func (p *StringSlicePool) Put(slice []string) {
	// Only put back slices that aren't too large
	if cap(slice) <= p.size*4 {
		p.pool.Put(slice)
	}
}

// MapPool provides a pool of string maps for CSV record parsing
type MapPool struct {
	pool sync.Pool
}

// NewMapPool creates a new map pool
func NewMapPool() *MapPool {
	return &MapPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make(map[string]string)
			},
		},
	}
}

// Get retrieves a map from the pool
func (p *MapPool) Get() map[string]string {
	m := p.pool.Get().(map[string]string)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// Put returns a map to the pool for reuse
func (p *MapPool) Put(m map[string]string) {
	// Only put back maps that aren't too large
	if len(m) <= 100 { // Reasonable limit for GTFS CSV rows
		p.pool.Put(m)
	}
}

// GlobalPools provides pre-configured pools for common use cases
var GlobalPools = struct {
	// SmallBuffer pool for small operations (1KB initial size)
	SmallBuffer *BufferPool
	// LargeBuffer pool for large operations (8KB initial size)
	LargeBuffer *BufferPool
	// CSVFields pool for CSV field parsing (typical GTFS row has ~10-20 fields)
	CSVFields *StringSlicePool
	// CSVRecord pool for CSV record maps
	CSVRecord *MapPool
}{
	SmallBuffer: NewBufferPool(1024),    // 1KB
	LargeBuffer: NewBufferPool(8192),    // 8KB
	CSVFields:   NewStringSlicePool(32), // 32 fields capacity
	CSVRecord:   NewMapPool(),
}

// Stats provides statistics about pool usage for monitoring
type Stats struct {
	// Hits is the number of times an object was retrieved from the pool
	Hits int64
	// Misses is the number of times a new object had to be created
	Misses int64
	// Puts is the number of times an object was returned to the pool
	Puts int64
}

// StatsBufferPool is a BufferPool with statistics tracking
type StatsBufferPool struct {
	*BufferPool
	stats Stats
	mutex sync.RWMutex
}

// NewStatsBufferPool creates a new buffer pool with statistics tracking
func NewStatsBufferPool(initialSize int) *StatsBufferPool {
	return &StatsBufferPool{
		BufferPool: NewBufferPool(initialSize),
	}
}

// Get retrieves a buffer and updates statistics
func (p *StatsBufferPool) Get() *bytes.Buffer {
	p.mutex.Lock()
	p.stats.Hits++
	p.mutex.Unlock()
	return p.BufferPool.Get()
}

// Put returns a buffer and updates statistics
func (p *StatsBufferPool) Put(buf *bytes.Buffer) {
	p.mutex.Lock()
	p.stats.Puts++
	p.mutex.Unlock()
	p.BufferPool.Put(buf)
}

// GetStats returns a copy of the current statistics
func (p *StatsBufferPool) GetStats() Stats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.stats
}

// ResetStats resets the statistics counters
func (p *StatsBufferPool) ResetStats() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.stats = Stats{}
}

// PooledCSVParser provides a CSV parser that uses memory pools
type PooledCSVParser struct {
	bufferPool *BufferPool
	fieldPool  *StringSlicePool
	recordPool *MapPool
}

// NewPooledCSVParser creates a new pooled CSV parser
func NewPooledCSVParser() *PooledCSVParser {
	return &PooledCSVParser{
		bufferPool: GlobalPools.SmallBuffer,
		fieldPool:  GlobalPools.CSVFields,
		recordPool: GlobalPools.CSVRecord,
	}
}

// ParseRecord parses a CSV record using pooled memory
// This is a helper function that demonstrates how to use the pools
func (p *PooledCSVParser) ParseRecord(fields []string, headers []string) map[string]string {
	if len(fields) != len(headers) {
		return nil
	}

	record := p.recordPool.Get()

	for i, header := range headers {
		if i < len(fields) {
			record[header] = fields[i]
		}
	}

	return record
}

// ReturnRecord returns a record map to the pool
func (p *PooledCSVParser) ReturnRecord(record map[string]string) {
	p.recordPool.Put(record)
}

// GetBuffer gets a buffer from the pool
func (p *PooledCSVParser) GetBuffer() *bytes.Buffer {
	return p.bufferPool.Get()
}

// ReturnBuffer returns a buffer to the pool
func (p *PooledCSVParser) ReturnBuffer(buf *bytes.Buffer) {
	p.bufferPool.Put(buf)
}

// GetFields gets a string slice from the pool
func (p *PooledCSVParser) GetFields() []string {
	return p.fieldPool.Get()
}

// ReturnFields returns a string slice to the pool
func (p *PooledCSVParser) ReturnFields(fields []string) {
	p.fieldPool.Put(fields)
}

// Helper functions for easy global pool access

// GetSmallBuffer gets a small buffer from the global pool
func GetSmallBuffer() *bytes.Buffer {
	return GlobalPools.SmallBuffer.Get()
}

// PutSmallBuffer returns a small buffer to the global pool
func PutSmallBuffer(buf *bytes.Buffer) {
	GlobalPools.SmallBuffer.Put(buf)
}

// GetLargeBuffer gets a large buffer from the global pool
func GetLargeBuffer() *bytes.Buffer {
	return GlobalPools.LargeBuffer.Get()
}

// PutLargeBuffer returns a large buffer to the global pool
func PutLargeBuffer(buf *bytes.Buffer) {
	GlobalPools.LargeBuffer.Put(buf)
}

// GetCSVFields gets a string slice from the global pool
func GetCSVFields() []string {
	return GlobalPools.CSVFields.Get()
}

// PutCSVFields returns a string slice to the global pool
func PutCSVFields(fields []string) {
	GlobalPools.CSVFields.Put(fields)
}

// GetCSVRecord gets a record map from the global pool
func GetCSVRecord() map[string]string {
	return GlobalPools.CSVRecord.Get()
}

// PutCSVRecord returns a record map to the global pool
func PutCSVRecord(record map[string]string) {
	GlobalPools.CSVRecord.Put(record)
}
