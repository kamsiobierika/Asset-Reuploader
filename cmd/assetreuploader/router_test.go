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
	if !shouldEmitDone(0, false, true, false) {
		t.Fatal("expected done signal after a completed job")
	}

	if shouldEmitDone(0, false, true, true) {
		t.Fatal("expected done signal to be emitted only once")
	}

	if shouldEmitDone(2, false, true, false) {
		t.Fatal("expected done signal to wait until the response payload is drained")
	}
}

func TestServeStateTransitions(t *testing.T) {
	state := &serveState{finished: true}

	if !state.startReupload() {
		t.Fatal("expected a new reupload to start")
	}

	if state.startReupload() {
		t.Fatal("expected a second reupload to be rejected while busy")
	}

	state.finishReupload()
	if state.busy || !state.finished {
		t.Fatal("expected the state to be reset after completion")
	}
}
