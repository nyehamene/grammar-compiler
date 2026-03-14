package server

import (
	"encoding/json"
	"testing"
)

func TestDocumentUriMarshalJSON(t *testing.T) {
	uri, err := ParseURI("file:///test.grammar")
	if err != nil {
		t.Fatalf("Failed to parse URI: %v", err)
	}

	data, err := json.Marshal(uri)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `"file:///test.grammar"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}

func TestDocumentUriUnmarshalJSON(t *testing.T) {
	data := `"file:///test.grammar"`

	var uri DocumentUri
	err := json.Unmarshal([]byte(data), &uri)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	expected := "file:///test.grammar"
	if uri.String() != expected {
		t.Errorf("Expected %s, got %s", expected, uri.String())
	}
}

func TestDocumentUriString(t *testing.T) {
	uri, err := ParseURI("file:///path/to/file.grammar")
	if err != nil {
		t.Fatalf("Failed to parse URI: %v", err)
	}

	str := uri.String()
	expected := "file:///path/to/file.grammar"
	if str != expected {
		t.Errorf("Expected %s, got %s", expected, str)
	}
}

func TestDocumentUriWithQueryAndFragment(t *testing.T) {
	uri, err := ParseURI("file:///path?query=value#fragment")
	if err != nil {
		t.Fatalf("Failed to parse URI: %v", err)
	}

	data, err := json.Marshal(uri)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `"file:///path?query=value#fragment"`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}
}
