package duhashtgo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
)

const (
	// Amount of hashes sent per query
	DEFAULT_RATE = 4096
	// Maximum error size sent by the server
	// Note this is an arbitrary value, most errors are shorter.
	MAX_ERROR_SIZE = 4096
)

const (
	// Protocol version
	PROTO_VERSION = '1';

	// Protocol query command
	PROTO_QUERY_COMMAND = 'q';
	// Protocol update command
	PROTO_UPDATE_COMMAND = 'u';
	// Protocol exit command
	PROTO_EXIT_COMMAND = 'e';

	// Protocol response success
	PROTO_RESPONSE_SUCCESS = 's';
	// Protocol response error
	PROTO_RESPONSE_ERROR = 'e';

	// Protocol qurery respond true
	PROTO_QUERY_TRUE = 1
	// Protocol qurery respond false
	PROTO_QUERY_FALSE = 0
)

const (
	ERROR_INVALID_RESPONSE = "invalid response"
	ERROR_INVALID_HASH = "invalid hash"
)

// Represents client configuration
type Client struct {
	address string
}

//
// Instantiate a new client.
//
func New(ip string, port uint16) *Client {
	return &Client {
		address: fmt.Sprintf("%s:%d", ip, port),
	}
}

// Represents a session / connection.
type Session struct {
	rate uint16
	conn net.Conn
}

//
// Establish a connection with the server.
//
func (c *Client) Connect() (*Session, error) {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return nil, err
	}

	return &Session{
		conn: conn,
		rate: DEFAULT_RATE,
	}, nil
}

//
// Sets the rate at which hashes are sent.
// Default is defined as `DEFAULT_RATE`.
//
func (c *Session) SetRate(rate uint16) {
	c.rate = rate
}

//
// Execute a raw command & handle error.
// Why? Useful for more custom logic.
//
func (s *Session) RawCommand(
	command byte,
	args []byte,
	resp_size int,
) ([]byte, error) {
	query := []byte{PROTO_VERSION, command}
	query = append(query, args...)

	_, err := s.conn.Write(query)
	if err != nil {
		return nil, err
	}

	err = s.ReadResponseStatus()
	if err != nil {
		return nil, err
	}

	response := make([]byte, resp_size)

	_, err = s.conn.Read(response)
	if err != nil {
		return nil, err
	}

	return response, nil 
}

//
// Send a query request to the server.
// Returns a slice of hashes that were not found.
//
func (s *Session) Query(hashes []string) ([]string, error) {
	var result []string

	hash_amount := len(hashes)
	query_amount := int(math.Ceil(float64(hash_amount) / float64(s.rate)))

	for i := range query_amount {

		window_start := int(s.rate) * i
		window_end := min(window_start + int(s.rate) - 1, hash_amount - 1)

		query := []byte{PROTO_VERSION, PROTO_QUERY_COMMAND}

		query = binary.BigEndian.AppendUint16(query, uint16(window_end) - uint16(window_start) + 1)

		for j := window_start; j < window_end + 1; j++ {
			n1, n2, err := MD5ToPairUint16(hashes[j])
			if err != nil {
				return nil, err
			}
			query = binary.BigEndian.AppendUint64(query, n1)
			query = binary.BigEndian.AppendUint64(query, n2)
		}

		// Send query
		_, err := s.conn.Write(query)
		if err != nil {
			return nil, err
		}

		//
		// Read status
		// Checks for errors
		//
		err = s.ReadResponseStatus()
		if err != nil {
			return nil, err
		}

		//
		// Read response 
		//
		response := make([]byte, window_end - window_start + 1)

		_, err = s.conn.Read(response)
		if err != nil {
			return nil, err
		}

		for result_pos, result_val := range response {
			switch result_val {
			case PROTO_QUERY_FALSE:
				result = append(result, hashes[window_start + result_pos])
			case PROTO_QUERY_TRUE:
			default:
				return nil, errors.New(ERROR_INVALID_RESPONSE)
			}
		}
	} 

	return result, nil
}

//
// Send update request.
// Returns the amount of hashes updated & error.
//
func (s *Session) Update(hashes []string) (int, error) {
	result := 0

	hash_amount := len(hashes)
	query_amount := int(math.Ceil(float64(hash_amount) / float64(s.rate)))

	for i := range query_amount {

		window_start := int(s.rate) * i
		window_end := min(window_start + int(s.rate) - 1, hash_amount - 1)

		query := []byte{PROTO_VERSION, PROTO_UPDATE_COMMAND}

		query = binary.BigEndian.AppendUint16(query, uint16(window_end) - uint16(window_start) + 1)

		for j := window_start; j < window_end + 1; j++ {
			n1, n2, err := MD5ToPairUint16(hashes[j])
			if err != nil {
				return result, err
			}
			query = binary.BigEndian.AppendUint64(query, n1)
			query = binary.BigEndian.AppendUint64(query, n2)
		}

		// Send query
		_, err := s.conn.Write(query)
		if err != nil {
			return result, err
		}

		//
		// Read status
		// Checks for errors
		//
		err = s.ReadResponseStatus()
		if err != nil {
			return result, err
		}

		//
		// Read response
		//
		response := make([]byte, 2)
		_, err = s.conn.Read(response)
		if err != nil {
			return result, err
		}

		result += int(binary.BigEndian.Uint16(response))
	} 

	return result, nil
}

//
// Sends the exit sequence to terminate the session.
//
func (s Session) Close() {
	s.conn.Write([]byte{PROTO_VERSION, PROTO_EXIT_COMMAND})
}

//
// Reads a single byte and matches for expected result.
// On error returns an error, on success returns nil.
//
func (s Session) ReadResponseStatus() error {
	status := make([]byte, 1)
	_, err := s.conn.Read(status)

	if err != nil {
		return err
	}

	switch status[0] {
	case PROTO_RESPONSE_SUCCESS: return nil
	case PROTO_RESPONSE_ERROR:
		buf := make([]byte, MAX_ERROR_SIZE)

		n, err := s.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
		}

		return errors.New(string(buf[:n]))
	default:
		return errors.New(ERROR_INVALID_RESPONSE)
	}
}

//
// Convert HEX MD5 string to two Uint16 integers.
//
func MD5ToPairUint16(hash string) (uint64, uint64, error) {
	if len(hash) != 32 {
		return 0, 0, errors.New(ERROR_INVALID_HASH)
	}

	n1, err := strconv.ParseUint(hash[:16], 16, 64)
	if err != nil {
		return 0, 0, err
	}

	n2, err := strconv.ParseUint(hash[16:], 16, 64)
	if err != nil {
		return 0, 0, err
	}
	
	return n1, n2, nil
}
