package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kartFr/Asset-Reuploader/internal/app/assets"
	"github.com/kartFr/Asset-Reuploader/internal/app/request"
	"github.com/kartFr/Asset-Reuploader/internal/app/response"
	"github.com/kartFr/Asset-Reuploader/internal/color"
	"github.com/kartFr/Asset-Reuploader/internal/files"
	"github.com/kartFr/Asset-Reuploader/internal/roblox"
)

var CompatiblePluginVersion = ""

func getOutputFileName(reuploadType string) string {
	t := time.Now()
	return fmt.Sprintf("Output_%s_%s.json", reuploadType, t.Format("2006-01-02_15-04-05"))
}

type serveState struct {
	mu       sync.Mutex
	busy     bool
	finished bool
	doneSent bool
	exportJSON bool
}

func (s *serveState) startReupload() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.busy || !s.finished {
		return false
	}

	s.busy = true
	s.finished = false
	s.doneSent = false
	s.exportJSON = false
	return true
}

func (s *serveState) finishReupload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.busy = false
	s.finished = true
	s.doneSent = false
}

func (s *serveState) canEmitResults(responseLen int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return responseLen > 0 && !s.busy && !s.finished
}

func (s *serveState) canEmitDone(responseLen int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return responseLen == 0 && !s.busy && s.finished && !s.doneSent
}

func (s *serveState) markDoneSent() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.doneSent = true
	s.exportJSON = false
}

func (s *serveState) setExportJSON(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.exportJSON = enabled
}

func (s *serveState) exportEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exportJSON
}

func shouldEmitResults(responseLen int, busy bool, finished bool) bool {
	return responseLen > 0 && !busy && !finished
}

func shouldEmitDone(responseLen int, busy bool, finished bool, doneSent bool) bool {
	return responseLen == 0 && !busy && finished && !doneSent
}

func serve(c *roblox.Client) error {
	var exportedJSONName string
	state := &serveState{finished: true}

	respHistory := make([]response.ResponseItem, 0)
	resp := response.New(func(i response.ResponseItem) {
		if state.exportEnabled() {
			respHistory = append(respHistory, i)

			j, err := json.Marshal(respHistory)
			if err != nil {
				log.Fatal(err)
			}

			if err := files.Write(exportedJSONName, string(j)); err != nil {
				log.Fatal(err)
			}
		}
	})

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		busy := false
		finished := true
		doneSent := false
		state.mu.Lock()
		busy = state.busy
		finished = state.finished
		doneSent = state.doneSent
		state.mu.Unlock()

		if shouldEmitResults(resp.Len(), busy, finished) {
			if err := resp.EncodeJSON(json.NewEncoder(w)); err != nil {
				log.Fatal(err)
				return
			}
			resp.Clear()
			return
		}

		if shouldEmitDone(resp.Len(), busy, finished, doneSent) {
			state.markDoneSent()
			resp.Clear()
			respHistory = make([]response.ResponseItem, 0)

			fmt.Fprint(w, "done")
			fmt.Println("Finished reuploading. (you can rerun without restarting)")
			return
		}
	})

	http.HandleFunc("POST /reupload", func(w http.ResponseWriter, r *http.Request) {
		if !state.startReupload() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		var req request.RawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			state.finishReupload()
			color.Error.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if CompatiblePluginVersion != "" && req.PluginVersion != CompatiblePluginVersion {
			state.finishReupload()
			w.WriteHeader(http.StatusConflict)
			return
		}

		if exists := assets.DoesModuleExist(req.AssetType); !exists {
			state.finishReupload()
			w.WriteHeader(http.StatusNotFound)
			return
		}

		startReupload, err := assets.NewReuploadHandlerWithType(req.AssetType, c, &req, resp)
		if err != nil {
			state.finishReupload()
			color.Error.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.ExportJSON {
			state.setExportJSON(true)
			exportedJSONName = getOutputFileName(req.AssetType)
		} else {
			state.setExportJSON(false)
		}

		go func() {
			start := time.Now()
			err := startReupload()
			state.finishReupload()
			if err != nil {
				color.Error.Println("Failed to start reuploading: ", err)
				return
			}

			duration := time.Since(start)
			fmt.Printf("Reuploading took %d hours, %d minutes, and %d seconds\n", int(duration.Hours()), int(duration.Minutes())%60, int(duration.Seconds())%60)
			fmt.Println("Waiting for client to finish changing ids...")
		}()

		w.WriteHeader(http.StatusOK)
	})

	return http.ListenAndServe(":"+port, nil)
}
