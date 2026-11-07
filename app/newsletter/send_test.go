package newsletter

import (
	"errors"
	"testing"

	"github.com/AlexGalhardo/vukem/queue"
)

func TestResendErrorRetriesTransientStatuses(t *testing.T) {
	for status, permanent := range map[int]bool{408: false, 409: false, 429: false, 500: false, 503: false, 400: true, 401: true, 422: true} {
		err := resendError(status, []byte("x"))
		if err == nil || errors.Is(err, queue.ErrPermanent) != permanent {
			t.Errorf("status %d: err=%v, want permanent=%v", status, err, permanent)
		}
	}
	if resendError(200, nil) != nil {
		t.Error("2xx must be nil")
	}
}
