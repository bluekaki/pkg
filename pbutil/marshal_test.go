package pbutil

import (
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	now := time.Now()
	ts, err := NewTimestamp(&now)
	if err != nil {
		t.Fatal(err)
	}

	req := &HelloRequest{
		TrackId:   "0987654321",
		Message:   "hello world !",
		Timestamp: ts,
		Duration:  NewDuration(time.Minute * 2),
		Status:    HelloRequest_Closing,
		Payloads: []*HelloRequest_Payload{
			&HelloRequest_Payload{Raw: []byte("xxxxxx")},
		},
		Name: &HelloRequest_EnName{EnName: "minami"},
		Meta: map[int32]bool{
			1: true,
		},
	}

	raw, err := ProtoMessage2JSON(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(raw))

	mp, err := ProtoMessage2Map(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(mp)
}
