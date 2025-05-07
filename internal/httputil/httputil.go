package httputil

import (
	"net/http"
	"net/url"
)

func redirectWithParameter(rw http.ResponseWriter, r *http.Request, baseURL, name, value string) {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	q := u.Query()
	q.Set(name, value)
	u.RawQuery = q.Encode()

	http.Redirect(rw, r, u.String(), http.StatusFound)
}

func RedirectWithError(rw http.ResponseWriter, r *http.Request, baseURL, message string) {
	redirectWithParameter(rw, r, baseURL, "error", message)
}

func RedirectWithSuccess(rw http.ResponseWriter, r *http.Request, baseURL, message string) {
	redirectWithParameter(rw, r, baseURL, "success", message)
}

func RedirectWithInformation(rw http.ResponseWriter, r *http.Request, baseURL, message string) {
	redirectWithParameter(rw, r, baseURL, "information", message)
}

func NotFound(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not found", http.StatusNotFound)
}
