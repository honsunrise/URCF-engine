package codec

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJsonRequests(t *testing.T) {
	const jsonStream = `
	{"method": "Ed", "jsonrpc": "2.0", "id": 0, "params": [0]}
`
	const jsonStreamArray = `
	[
		{"method": "Ed", "jsonrpc": "2.0", "id": 0, "params": [0]}
	]
`
	dec := json.NewDecoder(strings.NewReader(jsonStream))
	var r jsonrpcRequests
	err := dec.Decode(&r)
	if err != nil {
		t.Fatalf("%T: %v\n", err, err)
	}
	t.Logf("%T: %v\n", r, r)

	dec = json.NewDecoder(strings.NewReader(jsonStreamArray))
	var rs jsonrpcRequests
	err = dec.Decode(&rs)
	if err != nil {
		t.Fatalf("%T: %v\n", err, err)
	}
	t.Logf("%T: %v\n", rs, rs)
}
