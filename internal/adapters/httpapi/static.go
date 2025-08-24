package httpapi

import (
	"net/http"
	"path/filepath"
)

func ServeStatic(mux *http.ServeMux, root string) {
	fs := http.FileServer(http.Dir(root))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	}))
}
