package main

import (
	"testing"

	"github.com/nats-io/go-nats-streaming"
)

func Test_procVote(t *testing.T) {
	type args struct {
		msg *stan.Msg
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			procVote(tt.args.msg)
		})
	}
}
