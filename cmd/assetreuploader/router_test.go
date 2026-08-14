package main

import "testing"

func TestShouldEmitResults(t *testing.T) {
	if !shouldEmitResults(2, false, false) {
		t.Fatal("expected results to be emitted while the job is finishing")
	}

	if shouldEmitResults(0, false, false) {
		t.Fatal("expected no result payload when the response is empty")
	}
}

func TestShouldEmitDone(t *testing.T) {
	if !shouldEmitDone(0, false, true) {
		t.Fatal("expected done signal after a completed job")
	}

	if shouldEmitDone(2, false, true) {
		t.Fatal("expected done signal to wait until the response payload is drained")
	}
}
