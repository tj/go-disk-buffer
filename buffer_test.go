package buffer

import "github.com/bmizerany/assert"
import "testing"
import "time"

var config = &Config{
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

func TestOpen(t *testing.T) {
	b, err := New("/tmp/buffer", config)
	assert.Equal(t, nil, err)

	discard(b)

	err = b.Close()
	assert.Equal(t, nil, err)
}

func TestWrite(t *testing.T) {
	b, err := New("/tmp/buffer", config)
	assert.Equal(t, nil, err)

	discard(b)

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

	quit := make(chan bool)

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				_, err := b.Write([]byte("hello"))
				if err != nil {
					t.Fatalf("error: %s", err)
				}
			}
		}
	}()

	flush := <-b.Queue
	assert.Equal(t, int64(10), flush.Writes)
	assert.Equal(t, int64(50), flush.Bytes)
	assert.Equal(t, Writes, flush.Reason)

	flush = <-b.Queue
	assert.Equal(t, int64(10), flush.Writes)
	assert.Equal(t, int64(50), flush.Bytes)
	assert.Equal(t, Writes, flush.Reason)

	quit <- true
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

	quit := make(chan bool)

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				_, err := b.Write([]byte("hello world"))
				assert.Equal(t, nil, err)
			}
		}
	}()

	flush := <-b.Queue
	assert.Equal(t, int64(94), flush.Writes)
	assert.Equal(t, int64(1034), flush.Bytes)
	assert.Equal(t, Bytes, flush.Reason)

	flush = <-b.Queue
	assert.Equal(t, int64(94), flush.Writes)
	assert.Equal(t, int64(1034), flush.Bytes)
	assert.Equal(t, Bytes, flush.Reason)

	quit <- true
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

	discard(b)

	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		b.Write([]byte("hello world"))
	}
}

// - bufio
// - bench / race
// - prefix logs with filename
// - support zero value to ignore Flush* option
// - examples
// - interval
