package main

import (
	"log"
	"time"

	"github.com/rakyll/coop"
	"github.com/tj/go-disk-buffer"
)

func main() {
	b, err := buffer.New("/tmp/pets", &buffer.Config{
		FlushBytes:    20 << 20,
		FlushInterval: 15 * time.Second,
		BufferSize:    5 << 10,
		Verbosity:     0,
	})

	if err != nil {
		log.Fatalf("error opening: %s", err)
	}

	go func() {
		for file := range b.Queue {
			log.Printf("flushed %+v", file)
		}
	}()

	ops := 10000000
	con := 80
	per := ops / con
	start := time.Now()

	<-coop.Replicate(con, func() {
		for i := 0; i < per; i++ {
			_, err := b.Write([]byte("Tobi Ferret"))
			if err != nil {
				log.Fatalf("error writing: %s", err)
			}
		}
	})

	err = b.Close()
	if err != nil {
		log.Fatalf("error closing: %s", err)
	}

	log.Printf("ops: %d total, %d per, %d concurrent", ops, per, con)
	log.Printf("duration: %s", time.Since(start))
}
