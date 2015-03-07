package main

import "github.com/tj/go-disk-buffer"
import "github.com/rakyll/coop"
import "time"
import "log"

func main() {
	b, err := buffer.New("/tmp/pets", buffer.Config{
		FlushWrites:   200000,
		FlushBytes:    50 << 20,
		FlushInterval: 15 * time.Second,
		Verbosity:     0,
	})

	if err != nil {
		log.Fatalf("error opening: %s", err)
	}

	go func() {
		for file := range b.Queue {
			log.Printf("flushed %s", file)
		}
	}()

	ops := 1000000
	con := 30
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
