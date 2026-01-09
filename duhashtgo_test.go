package duhashtgo_test

import (
	"encoding/binary"
	"fmt"
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

func TestRawCommand(t *testing.T) {
	client := duhashtgo.New(IP, PORT)

	conn, err := client.Connect()
	if err != nil {
		t.Errorf("%s\n", err)
		return
	}
	defer conn.Close()

	hash := "0009775147D3A6A65D348E4ED7D2ABD8" 
	h1, h2, err := duhashtgo.MD5ToPairUint16(hash)
	if err != nil {
		t.Errorf("%s\n", err)
		return
	}

	var args []byte
	args = binary.BigEndian.AppendUint16(args, 1)
	args = binary.BigEndian.AppendUint64(args, h1)
	args = binary.BigEndian.AppendUint64(args, h2)

	resp_size := 1

	resp, err := conn.RawCommand(
		duhashtgo.PROTO_QUERY_COMMAND,
		args,
		resp_size,
	)

	fmt.Printf("Got result: %v\n", resp)

	_, err = conn.RawCommand(
		'x',
		args,
		resp_size,
	)

	if err.Error() != "ERROR_INVALID_COMMAND" {
		t.Errorf("Expected error ERROR_INVALID_COMMAND got %s\n", err)
		return
	}
}
