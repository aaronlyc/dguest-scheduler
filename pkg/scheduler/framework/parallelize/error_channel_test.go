package parallelize

import (
	"context"
	"errors"
	"testing"
)

func TestErrorChannel(t *testing.T) {
	errCh := NewErrorChannel()

	if actualErr := errCh.ReceiveError(); actualErr != nil {
		t.Errorf("expect nil from err channel, but got %v", actualErr)
	}

	err := errors.New("unknown error")
	errCh.SendError(err)
	if actualErr := errCh.ReceiveError(); actualErr != err {
		t.Errorf("expect %v from err channel, but got %v", err, actualErr)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh.SendErrorWithCancel(err, cancel)
	if actualErr := errCh.ReceiveError(); actualErr != err {
		t.Errorf("expect %v from err channel, but got %v", err, actualErr)
	}

	if ctxErr := ctx.Err(); ctxErr != context.Canceled {
		t.Errorf("expect context canceled, but got %v", ctxErr)
	}
}
