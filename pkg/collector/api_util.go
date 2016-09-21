/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
