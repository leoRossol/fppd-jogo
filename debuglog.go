package main

import (
	"io"
	"log"
	"os"
)

var dbg = log.New(io.Discard, "", log.LstdFlags)

func init() {
	// se debug=1 manda pra stderr (apenas se quiser ver logs na tela)
	if os.Getenv("DEBUG") == "1" {
		dbg.SetOutput(os.Stderr)
	}

	// painel de debug
	if os.Getenv("DEBUG_PANEL") == "1" {
		dbg.SetOutput(termboxBufferWriter{})
		return
	}

	// se client_log estiver defninido, escreve um arquivo (nao mostra na tela)
	if path := os.Getenv("CLIENT_LOG"); path != "" {
		if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			dbg.SetOutput(f)
		}
	}
}
