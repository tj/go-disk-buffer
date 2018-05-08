package buffer

import (
	"testing"
	"time"

	"github.com/tj/assert"
)

var config = &Config{
	Queue:         make(chan *Flush, 100),
	FlushWrites:   1000,
	FlushBytes:    1000,
	FlushInterval: time.Second,
	Verbosity:     0,
}

func discard(b *Buffer) {
	go func() {
		for range b.Queue {

		}
	}()
}

func write(buffer *Buffer, n int, b []byte) {
	go func() {
		for i := 0; i < n; i++ {
			_, err := buffer.Write(b)
			if err != nil {
				panic(err)
			}
		}
	}()
}

// Test immediate open / close.
func TestBuffer_Open(t *testing.T) {
	b, err := New("/tmp/buffer", config)
	assert.Equal(t, nil, err)

	err = b.Close()
	assert.Equal(t, nil, err)
}

// Test buffer writes.
func TestBuffer_Write(t *testing.T) {
	b, err := New("/tmp/buffer", config)
	assert.Equal(t, nil, err)

	n, err := b.Write([]byte("hello"))
	assert.Equal(t, nil, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, int64(1), b.writes)
	assert.Equal(t, int64(5), b.bytes)

	n, err = b.Write([]byte("world"))
	assert.Equal(t, nil, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, int64(2), b.writes)
	assert.Equal(t, int64(10), b.bytes)

	err = b.Close()
	assert.Equal(t, nil, err)
}

// Test flushing on write count.
func TestBuffer_Write_FlushOnWrites(t *testing.T) {
	b, err := New("/tmp/buffer", &Config{
		Queue:         make(chan *Flush, 100),
		FlushWrites:   10,
		FlushBytes:    1024,
		FlushInterval: time.Second,
		Verbosity:     0,
	})

	assert.Equal(t, nil, err)

	write(b, 25, []byte("hello"))

	flush := <-b.Queue
	assert.Equal(t, int64(10), flush.Writes)
	assert.Equal(t, int64(50), flush.Bytes)
	assert.Equal(t, Writes, flush.Reason)

	flush = <-b.Queue
	assert.Equal(t, int64(10), flush.Writes)
	assert.Equal(t, int64(50), flush.Bytes)
	assert.Equal(t, Writes, flush.Reason)

	err = b.Close()
	assert.Equal(t, nil, err)
}

// Test flushing on byte count.
func TestBuffer_Write_FlushOnBytes(t *testing.T) {
	b, err := New("/tmp/buffer", &Config{
		Queue:         make(chan *Flush, 100),
		FlushWrites:   10000,
		FlushBytes:    1024,
		FlushInterval: time.Second,
		Verbosity:     0,
	})

	assert.Equal(t, nil, err)

	write(b, 250, []byte("hello world"))
	flush := <-b.Queue
	assert.Equal(t, int64(94), flush.Writes)
	assert.Equal(t, int64(1034), flush.Bytes)
	assert.Equal(t, Bytes, flush.Reason)

	flush = <-b.Queue
	assert.Equal(t, int64(94), flush.Writes)
	assert.Equal(t, int64(1034), flush.Bytes)
	assert.Equal(t, Bytes, flush.Reason)

	err = b.Close()
	assert.Equal(t, nil, err)
}

// Test flushing on interval.
func TestBuffer_Write_FlushOnInterval(t *testing.T) {
	b, err := New("/tmp/buffer", &Config{
		Queue:         make(chan *Flush, 100),
		FlushInterval: time.Second,
	})

	assert.Equal(t, nil, err)

	b.Write([]byte("hello world"))
	b.Write([]byte("hello world"))

	flush := <-b.Queue
	assert.Equal(t, int64(2), flush.Writes)
	assert.Equal(t, int64(22), flush.Bytes)
	assert.Equal(t, Interval, flush.Reason)

	err = b.Close()
	assert.Equal(t, nil, err)
}

// Test config validation.
func TestConfig_Validate(t *testing.T) {
	_, err := New("/tmp/buffer", &Config{})
	assert.Equal(t, "at least one flush mechanism must be non-zero", err.Error())
}

// Benchmark buffer writes.
func BenchmarkBuffer_Write(t *testing.B) {
	b, err := New("/tmp/buffer", &Config{
		FlushWrites:   30000,
		FlushBytes:    1 << 30,
		FlushInterval: time.Minute,
		Verbosity:     0,
	})

	if err != nil {
		t.Fatalf("error: %s", err)
	}

	discard(b)

	t.ResetTimer()

	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			b.Write([]byte("hello world"))
		}
	})
}

// Benchmark buffer writes with bufio.
func BenchmarkBuffer_Write_Bufio(t *testing.B) {
	b, err := New("/tmp/buffer", &Config{
		FlushWrites:   30000,
		FlushBytes:    1 << 30,
		FlushInterval: time.Minute,
		BufferSize:    1 << 10,
		Verbosity:     0,
	})

	if err != nil {
		t.Fatalf("error: %s", err)
	}

	discard(b)

	t.ResetTimer()

	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			b.Write([]byte("hello world"))
		}
	})
}
