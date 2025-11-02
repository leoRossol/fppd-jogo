package main

import (
	"os"
	"strings"
	"sync"
)

var debugPanelEnabled = os.Getenv("DEBUG_PANEL") == "1"

// canal para receber linhas de log do writer
var debugPanelCh = make(chan string, 1024)

// termboxBufferWriter implementa io.Writer e envia linhas para o painel
type termboxBufferWriter struct{}

func (w termboxBufferWriter) Write(p []byte) (int, error) {
	s := string(p)
	for _, ln := range strings.Split(s, "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		select {
		case debugPanelCh <- ln:
		default:
			// descarta se cheio
		}
	}
	return len(p), nil
}

// ring buffer simples para manter histórico
var debugRing = struct {
	mu    sync.Mutex
	lines []string
	max   int
}{max: 300}

func debugPanelDrain() {
	for {
		select {
		case ln := <-debugPanelCh:
			debugRing.mu.Lock()
			if len(debugRing.lines) >= debugRing.max {
				copy(debugRing.lines, debugRing.lines[1:])
				debugRing.lines[len(debugRing.lines)-1] = ln
			} else {
				debugRing.lines = append(debugRing.lines, ln)
			}
			debugRing.mu.Unlock()
		default:
			return
		}
	}
}

func getDebugLines(n int) []string {
	debugRing.mu.Lock()
	defer debugRing.mu.Unlock()
	if n > len(debugRing.lines) {
		n = len(debugRing.lines)
	}
	out := make([]string, n)
	copy(out, debugRing.lines[len(debugRing.lines)-n:])
	return out
}
