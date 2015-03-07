package main

import "github.com/tj/go-disk-buffer"
import "io/ioutil"
import "time"
import "log"

func main() {
	b, err := buffer.New("/tmp/pets", buffer.Config{
		FlushWrites:   250,
		FlushBytes:    1 << 20,
		FlushInterval: 10 * time.Second,
		Verbosity:     0,
	})

	if err != nil {
		log.Fatalf("error opening: %s", err)
	}

	go func() {
		for file := range b.Queue {
			log.Printf("flushed %s", file)

			b, err := ioutil.ReadFile(file.Path)
			if err != nil {
				log.Fatalf("error reading: %s", err)
			}

			log.Printf("%q is %d bytes", file.Path, len(b))
		}
	}()

	for i := 0; i < 10000; i++ {
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
