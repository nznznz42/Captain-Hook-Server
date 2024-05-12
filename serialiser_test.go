package main

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"
)

func TestSerializeAndDeserializeRequest(t *testing.T) {
	mockReq, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString("test body"))
	if err != nil {
		t.Fatalf("Failed to create mock request: %v", err)
	}
	mockReq.Header.Set("Content-Type", "application/json")
	mockReq.Header.Add("X-Custom-Header", "value1")
	mockReq.Header.Add("X-Custom-Header", "value2")

	serializedData, err := SerializeRequest(mockReq)
	if err != nil {
		t.Fatalf("SerializeRequest failed: %v", err)
	}

	deserializedReq, err := DeserializeRequest(serializedData)
	if err != nil {
		t.Fatalf("DeserializeRequest failed: %v", err)
	}

	if !reflect.DeepEqual(mockReq.Method, deserializedReq.Method) {
		t.Errorf("Method mismatch: expected %s, got %s", mockReq.Method, deserializedReq.Method)
	}
	if !reflect.DeepEqual(mockReq.URL.String(), deserializedReq.URL.String()) {
		t.Errorf("URL mismatch: expected %s, got %s", mockReq.URL.String(), deserializedReq.URL.String())
	}
	if !reflect.DeepEqual(mockReq.Header, deserializedReq.Header) {
		t.Errorf("Header mismatch: expected %+v, got %+v", mockReq.Header, deserializedReq.Header)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(mockReq.Body)
	mockBody := buf.String()
	buf.Reset()
	buf.ReadFrom(deserializedReq.Body)
	deserializedBody := buf.String()
	if mockBody != deserializedBody {
		t.Errorf("Body mismatch: expected %s, got %s", mockBody, deserializedBody)
	}
}
