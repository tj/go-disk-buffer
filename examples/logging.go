package main

import "github.com/tj/go-disk-buffer"
import "time"
import "log"

func main() {
	b, err := buffer.New("/tmp/pets", buffer.Config{
		FlushWrites:   250,
		FlushBytes:    1 << 20,
		FlushInterval: 10 * time.Second,
		Verbosity:     1,
	})

	if err != nil {
		log.Fatalf("error opening: %s", err)
	}

	go func() {
		for range b.Queue {
			// discard
		}
	}()

	for i := 0; i < 100000; i++ {
		_, err := b.Write([]byte("Tobi Ferret"))
		if err != nil {
			log.Fatalf("error writing: %s", err)
		}
	}

	err = b.Close()
	if err != nil {
		log.Fatalf("error closing: %s", err)
	}
}
