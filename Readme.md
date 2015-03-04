# buffer

Package go-disk-buffer provides an io.Writer as a 1:N on-disk buffer, publishing
flushed files to a channel for processing. Files may be flushed via interval, write count, or byte size.

## Usage

#### type Buffer

```go
type Buffer struct {
	*Config

	sync.Mutex
}
```

Buffer represents a 1:N on-disk buffer.

#### func  New

```go
func New(path string, config *Config) (*Buffer, error)
```
New buffer at `path`. The path given is used for the base of the filenames
created, which append ".{pid}.{id}".

#### func (*Buffer) Bytes

```go
func (b *Buffer) Bytes() int64
```
Bytes returns the number of bytes made to the current file.

#### func (*Buffer) Close

```go
func (b *Buffer) Close() error
```
Close the underlying file. TODO: flush

#### func (*Buffer) Flush

```go
func (b *Buffer) Flush() error
```
Flush forces a flush.

#### func (*Buffer) FlushReason

```go
func (b *Buffer) FlushReason(reason Reason) error
```
FlushReason flushes for the given reason.

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
	FlushWrites   int64
	FlushBytes    int64
	FlushInterval time.Duration
	Queue         chan *Flush
	Verbosity     int
	Logger        *log.Logger
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
