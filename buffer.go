// Package go-disk-buffer provides an io.Writer as a 1:N on-disk buffer,
// publishing flushed files to a channel for processing.
//
// Files may be flushed via interval, write count, or byte size.
//
// All exported methods are thread-safe.
package buffer

import "sync/atomic"
import "sync"
import "time"
import "log"
import "fmt"
import "os"

// PID for unique filename.
var pid = os.Getpid()

// Ids for unique filename.
var ids = int64(0)

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
	Closed time.Time     `json:"closed"`
	Age    time.Duration `json:"age"`
}

// Config for disk buffer.
type Config struct {
	FlushWrites   int64         // Flush after N writes, zero to disable
	FlushBytes    int64         // Flush after N bytes, zero to disable
	FlushInterval time.Duration // Flush after duration, zero to disable
	Queue         chan *Flush   // Queue of flushed files
	Verbosity     int           // Verbosity level, 0-2
	Logger        *log.Logger   // Logger instance
}

// Buffer represents a 1:N on-disk buffer.
type Buffer struct {
	*Config

	verbosity int
	path      string
	ids       int64
	id        int64

	sync.Mutex
	opened time.Time
	writes int64
	bytes  int64
	file   *os.File
}

// New buffer at `path`. The path given is used for the base
// of the filenames created, which append ".{pid}.{id}.{fid}".
func New(path string, config *Config) (*Buffer, error) {
	id := atomic.AddInt64(&ids, 1)

	b := &Buffer{
		Config:    config,
		path:      path,
		id:        id,
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

// Open a new buffer.
func (b *Buffer) open() error {
	path := b.pathname()

	b.log(1, "opening %s", path)
	f, err := os.Create(path)
	if err != nil {
		return err
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

	if b.FlushWrites != 0 && b.Writes() >= b.FlushWrites {
		err := b.FlushReason(Writes)
		if err != nil {
			return n, err
		}
	}

	if b.FlushBytes != 0 && b.Bytes() >= b.FlushBytes {
		err := b.FlushReason(Bytes)
		if err != nil {
			return n, err
		}
	}

	return n, err
}

// Close the underlying file after flushing.
func (b *Buffer) Close() error {
	b.Lock()
	defer b.Unlock()
	return b.flush(Forced)
}

// Flush forces a flush.
func (b *Buffer) Flush() error {
	return b.FlushReason(Forced)
}

// FlushReason flushes for the given reason and re-opens.
func (b *Buffer) FlushReason(reason Reason) error {
	b.Lock()
	defer b.Unlock()

	err := b.flush(reason)
	if err != nil {
		return err
	}

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

// Flush for the given reason without re-open.
func (b *Buffer) flush(reason Reason) error {
	b.log(1, "flushing (%s)", reason)

	err := b.close()
	if err != nil {
		return err
	}

	b.Queue <- &Flush{
		Reason: reason,
		Writes: b.writes,
		Bytes:  b.bytes,
		Opened: b.opened,
		Closed: time.Now(),
		Path:   b.file.Name() + ".closed",
		Age:    time.Since(b.opened),
	}

	return nil
}

// Close existing file after a rename.
func (b *Buffer) close() error {
	if b.file == nil {
		return nil
	}

	path := b.file.Name()

	b.log(2, "renaming %q", path)
	err := os.Rename(path, path+".closed")
	if err != nil {
		return err
	}

	b.log(2, "closing %q", path)
	return b.file.Close()
}

// Pathname for a new buffer.
func (b *Buffer) pathname() string {
	fid := atomic.AddInt64(&b.ids, 1)
	return fmt.Sprintf("%s.%d.%d.%d", b.path, pid, b.id, fid)
}

// Log helper.
func (b *Buffer) log(n int, msg string, args ...interface{}) {
	if b.Verbosity >= n {
		b.Logger.Printf(msg, args...)
	}
}
