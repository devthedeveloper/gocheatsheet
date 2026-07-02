// json: struct ⇄ JSON, both directions.
package main

import (
	"encoding/json"
	"fmt"
)

type Note struct {
	ID   int      `json:"id"`
	Text string   `json:"text"`
	Tags []string `json:"tags,omitempty"`
}

func main() {
	// struct → JSON (marshal)
	n := Note{ID: 1, Text: "hello",
		Tags: []string{"go", "demo"}}
	out, _ := json.MarshalIndent(n, "", "  ")
	fmt.Println(string(out))

	// JSON → struct (unmarshal)
	raw := `{"id": 2, "text": "from the wire"}`
	var m Note
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", m)
	// Tags omitted in input → nil slice,
	// and omitempty hid it in the output
}
