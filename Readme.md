# buffer

Package go-disk-buffer provides an io.Writer as a 1:N on-disk buffer, publishing
flushed files to a channel for processing.

Files may be flushed via interval, write count, or byte size.

All exported methods are thread-safe.

## Usage

#### type Buffer

```go
type Buffer struct {
	Config

	sync.Mutex
}
```

Buffer represents a 1:N on-disk buffer.

#### func  New

```go
func New(path string, config Config) (*Buffer, error)
```
New buffer at `path`. The path given is used for the base of the filenames
created, which append ".{pid}.{id}.{fid}".

#### func (*Buffer) Bytes

```go
func (b *Buffer) Bytes() int64
```
Bytes returns the number of bytes made to the current file.

#### func (*Buffer) Close

```go
func (b *Buffer) Close() error
```
Close the underlying file after flushing.

#### func (*Buffer) Flush

```go
func (b *Buffer) Flush() error
```
Flush forces a flush.

#### func (*Buffer) FlushReason

```go
func (b *Buffer) FlushReason(reason Reason) error
```
FlushReason flushes for the given reason and re-opens.

#### func (*Buffer) Write

```go
func (b *Buffer) Write(data []byte) (int, error)
```
Write implements io.Writer.

#### func (*Buffer) Writes

```go
func (b *Buffer) Writes() int64
```
Writes returns the number of writes made to the current file.

#### type Config

```go
type Config struct {
	FlushWrites   int64         // Flush after N writes, zero to disable
	FlushBytes    int64         // Flush after N bytes, zero to disable
	FlushInterval time.Duration // Flush after duration, zero to disable
	Queue         chan *Flush   // Queue of flushed files
	Verbosity     int           // Verbosity level, 0-3
	Logger        *log.Logger   // Logger instance
}
```

Config for disk buffer.

#### type Flush

```go
type Flush struct {
	Reason Reason        `json:"reason"`
	Path   string        `json:"path"`
	Writes int64         `json:"writes"`
	Bytes  int64         `json:"bytes"`
	Opened time.Time     `json:"opened"`
	Closed time.Time     `json:"closed"`
	Age    time.Duration `json:"age"`
}
```

Flush represents a flushed file.

#### type Reason

```go
type Reason string
```

Reason for flush.

```go
const (
	Forced   Reason = "forced"
	Writes   Reason = "writes"
	Bytes    Reason = "bytes"
	Interval Reason = "interval"
)
```
Flush reasons.
