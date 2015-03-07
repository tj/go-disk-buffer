package buffer

import "github.com/bmizerany/assert"
import "testing"
import "time"
import "fmt"

var config = &Config{
	FlushWrites:   1000,
	FlushBytes:    1000,
	FlushInterval: time.Second,
	Verbosity:     0,
}

func TestOpen(t *testing.T) {
	b, err := New("/tmp/buffer", config)
	assert.Equal(t, nil, err)

	err = b.Close()
	assert.Equal(t, nil, err)
}

func TestWrite(t *testing.T) {
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

func TestFlushWrites(t *testing.T) {
	b, err := New("/tmp/buffer", &Config{
		FlushWrites:   10,
		FlushBytes:    1024,
		FlushInterval: time.Second,
		Verbosity:     0,
	})

	assert.Equal(t, nil, err)

	go func() {
		for i := 0; i < 22; i++ {
			_, err := b.Write([]byte("hello"))
			assert.Equal(t, nil, err)
		}
	}()

	flush := <-b.Queue
	assert.Equal(t, fmt.Sprintf("/tmp/buffer.%d.1", pid), flush.Path)
	assert.Equal(t, int64(10), flush.Writes)
	assert.Equal(t, int64(50), flush.Bytes)
	assert.Equal(t, Writes, flush.Reason)

	flush = <-b.Queue
	assert.Equal(t, fmt.Sprintf("/tmp/buffer.%d.2", pid), flush.Path)
	assert.Equal(t, int64(10), flush.Writes)
	assert.Equal(t, int64(50), flush.Bytes)
	assert.Equal(t, Writes, flush.Reason)

	err = b.Close()
	assert.Equal(t, nil, err)
}

func TestFlushBytes(t *testing.T) {
	b, err := New("/tmp/buffer", &Config{
		FlushWrites:   10000,
		FlushBytes:    1024,
		FlushInterval: time.Second,
		Verbosity:     0,
	})

	assert.Equal(t, nil, err)

	go func() {
		for i := 0; i < 200; i++ {
			_, err := b.Write([]byte("hello world"))
			assert.Equal(t, nil, err)
		}
	}()

	flush := <-b.Queue
	assert.Equal(t, fmt.Sprintf("/tmp/buffer.%d.1", pid), flush.Path)
	assert.Equal(t, int64(94), flush.Writes)
	assert.Equal(t, int64(1034), flush.Bytes)
	assert.Equal(t, Bytes, flush.Reason)

	flush = <-b.Queue
	assert.Equal(t, fmt.Sprintf("/tmp/buffer.%d.2", pid), flush.Path)
	assert.Equal(t, int64(94), flush.Writes)
	assert.Equal(t, int64(1034), flush.Bytes)
	assert.Equal(t, Bytes, flush.Reason)

	err = b.Close()
	assert.Equal(t, nil, err)
}

func BenchmarkWrite(t *testing.B) {
	b, err := New("/tmp/buffer", &Config{
		FlushWrites:   30000,
		FlushBytes:    1 << 30,
		FlushInterval: time.Minute,
		Verbosity:     0,
	})

	if err != nil {
		t.Fatalf("error: %s", err)
	}

	go func() {
		for range b.Queue {
			// whoop
		}
	}()

	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		b.Write([]byte("hello world"))
	}
}

// - bufio
// - bench / race
// - flush on close
// - prefix logs with filename
// - support zero value to ignore Flush* option
// - examples
// - interval
