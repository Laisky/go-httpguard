package dependencysecurity_test

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"testing"
	"time"
)

func TestProtoJSONRejectsMissingValue(t *testing.T) {
	done := make(chan error, 1)
	go func() { done <- protojson.Unmarshal([]byte("{\"\":}"), &anypb.Any{}) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("invalid JSON accepted")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("invalid JSON did not terminate")
	}
}
