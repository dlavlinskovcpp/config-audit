package httpserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"configaudit/internal/app"
	"configaudit/internal/audit"
	"configaudit/internal/parser"
)

const maxScanBodySize = 1 << 20

type scanResponse struct {
	Problems []audit.Problem `json:"problems"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(scanner app.Scanner) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/_info", infoHandler)
	mux.HandleFunc("/scan", scanHandler(scanner))
	return mux
}

func ListenAndServe(addr string, scanner app.Scanner) error {
	server := &http.Server{
		Addr:              addr,
		Handler:           NewHandler(scanner),
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"name":   "configaudit",
		"status": "ok",
	})
}

func scanHandler(scanner app.Scanner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		format, err := requestFormat(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxScanBodySize))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: fmt.Sprintf("read request body: %v", err)})
			return
		}

		problems, err := scanner.ScanContent("", body, format)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, scanResponse{Problems: problems})
	}
}

func requestFormat(r *http.Request) (parser.Format, error) {
	if format := r.URL.Query().Get("format"); format != "" {
		return parser.ParseFormat(format)
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	switch {
	case strings.Contains(contentType, "json"):
		return parser.FormatJSON, nil
	case strings.Contains(contentType, "yaml"), strings.Contains(contentType, "yml"):
		return parser.FormatYAML, nil
	default:
		return parser.FormatAuto, nil
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
