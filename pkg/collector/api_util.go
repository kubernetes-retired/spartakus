package collector

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// contentTypeMiddleware wraps and returns a httprouter.Handle, validating the request
// content type is compatible with the contentTypes list.
// It writes a HTTP 415 error if that fails.
//
// Only PUT, POST, and PATCH requests are considered.
func contentTypeMiddleware(handle httprouter.Handle, contentTypes ...string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if !(r.Method == "PUT" || r.Method == "POST" || r.Method == "PATCH") {
			handle(w, r, p)
			return
		}

		for _, ct := range contentTypes {
			if isContentType(r.Header, ct) {
				handle(w, r, p)
				return
			}
		}
		msg := fmt.Sprintf("Unsupported content type %q; expected one of %q", r.Header.Get("Content-Type"), contentTypes)
		http.Error(w, msg, http.StatusUnsupportedMediaType)
	}
}

// isContentType validates the Content-Type header
// is contentType. That is, its type and subtype match.
func isContentType(h http.Header, contentType string) bool {
	ct := h.Get("Content-Type")
	if i := strings.IndexRune(ct, ';'); i != -1 {
		ct = ct[0:i]
	}
	return ct == contentType
}

// writeError writes an error value.
func writeError(w http.ResponseWriter, code int, err error) error {
	w.WriteHeader(code)

	_, err = w.Write([]byte("Error: " + err.Error()))
	if err != nil {
		return err
	}
	return nil
}
