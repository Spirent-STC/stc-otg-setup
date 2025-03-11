package utils

import (
	"fmt"
	"testing"
	"time"
)

func PollStatus(t *testing.T, interval, timeout time.Duration, fetch func(t *testing.T, paramlist []any) (bool, error), fetchparams ...any) (reterr error) {
	start := time.Now()
	reterr = nil
	for {
		if gotit, err := fetch(t, fetchparams[:]); err != nil {
			reterr = fmt.Errorf("pollstatus error: %s", err.Error())
			break
		} else if gotit {
			t.Logf("Poll Status OK!")
			reterr = nil
			break
		}

		elapsed := time.Since(start).Seconds()
		if elapsed > float64(timeout) {
			reterr = fmt.Errorf("pollstatus timeout")
			break
		}

		time.Sleep(interval * time.Second)
	}

	return reterr
}
