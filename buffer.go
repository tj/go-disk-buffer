//
// Package go-disk-buffer provides an io.Writer as a 1:N on-disk buffer,
// publishing flushed files to a channel for processing.
//
// Files may be flushed via interval, write count, or byte size.
//
package buffer

import "sync/atomic"
import "sync"
import "time"
import "log"
import "fmt"
import "os"

// PID for unique filename.
var pid = os.Getpid()

// Reason for flush.
type Reason string

// Flush reasons.
const (
	Forced   Reason = "forced"
	Writes   Reason = "writes"
	Bytes    Reason = "bytes"
	Interval Reason = "interval"
)

// Flush represents a flushed file.
type Flush struct {
	Reason Reason        `json:"reason"`
	Path   string        `json:"path"`
	Writes int64         `json:"writes"`
	Bytes  int64         `json:"bytes"`
	Opened time.Time     `json:"opened"`
	Age    time.Duration `json:"age"`
}

// Config for disk buffer.
type Config struct {
	FlushWrites   int64
	FlushBytes    int64
	FlushInterval time.Duration
	Queue         chan *Flush
	Verbosity     int
	Logger        *log.Logger
}

// Buffer represents a 1:N on-disk buffer.
type Buffer struct {
	*Config

	verbosity int
	path      string
	ids       int64

	sync.Mutex
	opened time.Time
	writes int64
	bytes  int64
	file   *os.File
}

// New buffer at `path`. The path given is used for the base
// of the filenames created, which append ".{pid}.{id}".
func New(path string, config *Config) (*Buffer, error) {
	b := &Buffer{
		Config:    config,
		path:      path,
		verbosity: 1,
	}

	if b.Logger == nil {
		b.Logger = log.New(os.Stderr, "buffer ", log.LstdFlags)
	}

	if b.Queue == nil {
		b.Queue = make(chan *Flush)
	}

	return b, b.open()
}

// Open a new buffer, closing the previous when present.
func (b *Buffer) open() error {
	path := b.pathname()

	b.log(1, "opening %s", path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	b.Lock()
	defer b.Unlock()

	if b.file != nil {
		b.file.Close()
	}

	b.opened = time.Now()
	b.writes = 0
	b.bytes = 0
	b.file = f

	return nil
}

// Write implements io.Writer.
func (b *Buffer) Write(data []byte) (int, error) {
	b.log(2, "write %s", data)

	n, err := b.write(data)
	if err != nil {
		return n, err
	}

	if b.Writes() >= b.FlushWrites {
		err := b.FlushReason(Writes)
		if err != nil {
			return n, err
		}
	}

	if b.Bytes() >= b.FlushBytes {
		err := b.FlushReason(Bytes)
		if err != nil {
			return n, err
		}
	}

	return n, err
}

// Close the underlying file.
// TODO: flush
func (b *Buffer) Close() error {
	b.Lock()
	defer b.Unlock()
	return b.file.Close()
}

// Flush forces a flush.
func (b *Buffer) Flush() error {
	return b.FlushReason(Forced)
}

// FlushReason flushes for the given reason.
func (b *Buffer) FlushReason(reason Reason) error {
	b.log(1, "flushing (%s)", reason)

	b.Lock()

	flush := &Flush{
		Reason: reason,
		Writes: b.writes,
		Bytes:  b.bytes,
		Opened: b.opened,
		Path:   b.file.Name(),
		Age:    time.Since(b.opened),
	}

	b.Queue <- flush

	b.Unlock()

	return b.open()
}

// Write with metrics.
func (b *Buffer) write(data []byte) (int, error) {
	b.Lock()
	defer b.Unlock()

	b.writes += 1
	b.bytes += int64(len(data))

	return b.file.Write(data)
}

// Writes returns the number of writes made to the current file.
func (b *Buffer) Writes() int64 {
	return atomic.LoadInt64(&b.writes)
}

// Bytes returns the number of bytes made to the current file.
func (b *Buffer) Bytes() int64 {
	return atomic.LoadInt64(&b.bytes)
}

// Pathname for a new buffer.
func (b *Buffer) pathname() string {
	return fmt.Sprintf("%s.%d.%d", b.path, pid, b.id())
}

// Id for a new buffer.
func (b *Buffer) id() int64 {
	return atomic.AddInt64(&b.ids, 1)
}

// Log helper.
func (b *Buffer) log(n int, msg string, args ...interface{}) {
	if b.Verbosity >= n {
		b.Logger.Printf(msg, args...)
	}
}
