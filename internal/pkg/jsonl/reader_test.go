package jsonl

import (
	"strings"
	"testing"
)

func TestReader_ValidJSONL(t *testing.T) {
	input := `{"name":"Alice","age":30}
{"name":"Bob","age":25}
`
	reader := NewReaderFromReader(strings.NewReader(input))

	rec1, err := reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if rec1["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", rec1["name"])
	}

	rec2, err := reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if rec2["name"] != "Bob" {
		t.Errorf("expected Bob, got %v", rec2["name"])
	}

	_, err = reader.Next()
	if err == nil {
		t.Error("expected EOF, got nil")
	}
}

func TestReader_EmptyLines(t *testing.T) {
	input := `{"name":"Alice"}

{"name":"Bob"}
`
	reader := NewReaderFromReader(strings.NewReader(input))

	rec1, err := reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if rec1["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", rec1["name"])
	}

	rec2, err := reader.Next()
	if err != nil {
		t.Fatal(err)
	}
	if rec2["name"] != "Bob" {
		t.Errorf("expected Bob, got %v", rec2["name"])
	}
}

func TestReader_ParseError(t *testing.T) {
	input := `{"name":"Alice"}
not valid json
{"name":"Bob"}
`
	reader := NewReaderFromReader(strings.NewReader(input))

	_, err := reader.Next()
	if err != nil {
		t.Fatal(err)
	}

	_, err = reader.Next()
	if err == nil {
		t.Error("expected parse error, got nil")
	}
}

func TestMustGetField(t *testing.T) {
	rec := Record{"name": "Alice", "age": 30}

	if MustGetField(rec, "name") != "Alice" {
		t.Error("expected Alice")
	}
	if MustGetField(rec, "age") != 30 {
		t.Error("expected 30")
	}
	if MustGetField(rec, "missing") != nil {
		t.Error("expected nil for missing field")
	}
}

func TestHasField(t *testing.T) {
	rec := Record{"name": "Alice", "age": 30}

	if !HasField(rec, "name") {
		t.Error("expected true for existing field")
	}
	if HasField(rec, "missing") {
		t.Error("expected false for missing field")
	}
}
