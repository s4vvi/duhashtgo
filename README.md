# duhashtgo

Go library for querying https://github.com/s4vvi/duhashtsrv.

Plese see https://github.com/s4vvi/duhashtsrv for server setup & https://github.com/s4vvi/duhashtcli for CLI queries.

Add:
```
go get github.com/s4vvi/duhashtgo@506375f
```

Example:
```go
package main

import (
	"log"
	"github.com/s4vvi/duhashtgo"
)

const IP = "127.0.0.1"
const PORT = 1337

func main() {
	client := duhashtgo.New(IP, PORT)

	conn, err := client.Connect()
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// Just an example
	// Default is 4K
	conn.SetRate(1)

	hashes := []string{
		"0009775147D3A6A65D348E4ED7D2ABD8",
		"0009775147D3A6A65D348E4ED7D2ABD1",
	}

	hashes, err = conn.Query(hashes)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Got %d misses.\n", len(hashes))
	log.Println(hashes)

	n, err := conn.Update(hashes)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Added %d hashes.\n", n)
}
```
