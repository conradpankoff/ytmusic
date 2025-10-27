package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// WriteJSONResponse writes a JSON response
func WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteXMLResponse writes an XML response
func WriteXMLResponse(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	
	// Add XML header
	if _, err := w.Write([]byte(xml.Header)); err != nil {
		return err
	}
	
	return xml.NewEncoder(w).Encode(data)
}

// WriteErrorResponse writes an error response in the appropriate format
func WriteErrorResponse(w http.ResponseWriter, status int, message, details string) {
	accept := w.Header().Get("Accept")
	
	errorResp := APIError{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	}
	
	if details != "" {
		errorResp.Message = fmt.Sprintf("%s: %s", message, details)
	}
	
	if strings.Contains(accept, "application/xml") {
		WriteXMLResponse(w, status, errorResp)
	} else {
		WriteJSONResponse(w, status, errorResp)
	}
}

// GetPaginationParams extracts pagination parameters from request
func GetPaginationParams(r *http.Request) (page, limit int, err error) {
	page = 1
	limit = 50 // default limit
	
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err != nil {
			return 0, 0, fmt.Errorf("invalid page parameter: %s", pageStr)
		} else if p < 1 {
			return 0, 0, fmt.Errorf("page must be >= 1")
		} else {
			page = p
		}
	}
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err != nil {
			return 0, 0, fmt.Errorf("invalid limit parameter: %s", limitStr)
		} else if l < 1 {
			return 0, 0, fmt.Errorf("limit must be >= 1")
		} else if l > 1000 {
			return 0, 0, fmt.Errorf("limit must be <= 1000")
		} else {
			limit = l
		}
	}
	
	return page, limit, nil
}

// GetIDFromPath extracts ID from URL path
func GetIDFromPath(r *http.Request, paramName string) (int, error) {
	// This would typically use a router like mux to extract path parameters
	// For now, we'll assume the ID is passed as a query parameter or extracted by the router
	
	idStr := r.URL.Query().Get(paramName)
	if idStr == "" {
		return 0, fmt.Errorf("missing %s parameter", paramName)
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %s", paramName, idStr)
	}
	
	if id < 1 {
		return 0, fmt.Errorf("%s must be >= 1", paramName)
	}
	
	return id, nil
}

// DetermineResponseFormat determines the response format based on Accept header and query params
func DetermineResponseFormat(r *http.Request) string {
	// Check explicit format parameter first
	if format := r.URL.Query().Get("format"); format != "" {
		switch strings.ToLower(format) {
		case "json":
			return "json"
		case "xml":
			return "xml"
		}
	}
	
	// Check Accept header
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/xml") {
		return "xml"
	}
	
	// Default to JSON
	return "json"
}

// WriteResponse writes a response in the appropriate format
func WriteResponse(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	format := DetermineResponseFormat(r)
	
	var err error
	switch format {
	case "xml":
		err = WriteXMLResponse(w, status, data)
	default:
		err = WriteJSONResponse(w, status, data)
	}
	
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}