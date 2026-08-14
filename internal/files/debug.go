package files

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type DebugWriter struct {
	assetType string
	mutex     sync.Mutex
	lines     []string
}

func NewDebugWriter(assetType string) *DebugWriter {
	return &DebugWriter{
		assetType: assetType,
		lines:     make([]string, 0),
	}
}

func (dw *DebugWriter) WriteLine(line string) {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()
	dw.lines = append(dw.lines, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05.000"), line))
}

func (dw *DebugWriter) SaveToFile() error {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	if len(dw.lines) == 0 {
		return nil
	}

	filename := fmt.Sprintf("%s.debug.txt", dw.assetType)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range dw.lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func (dw *DebugWriter) Clear() {
	dw.mutex.Lock()
	defer dw.mutex.Unlock()
	dw.lines = dw.lines[:0]
}
