package duhashtgo_test

import (
	"testing"

	"github.com/s4vvi/duhashtgo"
)

const IP = "127.0.0.1"
const PORT = 1337

func TestQueryMisses(t *testing.T) {
	client := duhashtgo.New(IP, PORT)

	conn, err := client.Connect()
	if err != nil {
		t.Errorf("%s\n", err)
		return
	}
	defer conn.Close()

	conn.SetRate(1)

	hashes := []string{
		"0009775147D3A6A65D348E4ED7D2ABD8",
		"0009775147D3A6A65D348E4ED7D2ABD1",
	}

	_, err = conn.Query(hashes)
	if err != nil {
		t.Errorf("%s\n", err)
	}
}

func TestUpdate(t *testing.T) {
	client := duhashtgo.New(IP, PORT)

	conn, err := client.Connect()
	if err != nil {
		t.Errorf("%s\n", err)
		return
	}
	defer conn.Close()

	conn.SetRate(1)

	hashes := []string{
		"0009775147D3A6A65D348E4ED7D2ABD8",
		"0009775147D3A6A65D348E4ED7D2ABD1",
	}

	_, err = conn.Update(hashes)
	if err != nil {
		t.Errorf("%s\n", err)
	}
}
