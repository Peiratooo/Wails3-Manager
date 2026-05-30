package desktop

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const localFilePath = "/local/file"

func LocalFileMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == localFilePath {
			ServeLocalFile(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ServeLocalFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	absPath := getLocalFilePathFromRequest(r)
	absPath = cleanLocalPath(absPath)

	if absPath == "" {
		http.NotFound(w, r)
		return
	}

	if !isReadableLocalFile(absPath) {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, absPath)
}

func getLocalFilePathFromRequest(r *http.Request) string {
	if value := r.URL.Query().Get("path"); value != "" {
		return value
	}

	raw := strings.TrimSpace(r.URL.RawQuery)
	if raw == "" {
		return ""
	}

	if strings.HasPrefix(raw, "path=") {
		raw = strings.TrimPrefix(raw, "path=")
	}

	value, err := url.QueryUnescape(raw)
	if err != nil {
		return raw
	}

	return value
}

func cleanLocalPath(path string) string {
	path = strings.TrimSpace(strings.Trim(path, `"`))
	if path == "" {
		return ""
	}

	if !filepath.IsAbs(path) {
		return ""
	}

	return filepath.Clean(path)
}

func isReadableLocalFile(absPath string) bool {
	if absPath == "" || !filepath.IsAbs(absPath) {
		return false
	}

	info, err := os.Stat(absPath)
	if err != nil || info.IsDir() {
		return false
	}

	return true
}
