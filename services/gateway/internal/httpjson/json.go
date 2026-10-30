package httpjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"reflect"
	"strings"
)

const maxJSONBodyBytes int64 = 1 << 20 // 1 MiB

var (
	errUnsupportedMediaType = errors.New("unsupported media type")
	errBodyTooLarge         = errors.New("JSON body too large")
	errEmptyBody            = errors.New("empty JSON body")
	errInvalidJSON          = errors.New("invalid JSON body")
	errMultipleJSONValues   = errors.New("multiple JSON values")
	errInvalidDecodeTarget  = errors.New("decode target must be a non-nil pointer")
)

type errorResponse struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MyJSONDecode reads and validates JSON.
// If decoding fails, it sends an error response and returns false.
func MyJSONDecode(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := decodeJSON(w, r, dst); err != nil {
		writeDecodeError(w, err)
		return false
	}

	return true
}

// WriteJSONError sends a JSON error response.
func WriteJSONError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	payload := errorResponse{
		Error: errorDetail{
			Code:    code,
			Message: message,
		},
	}

	if err := MyJSONResponse(w, status, payload); err != nil {
		slog.Error("write JSON error response", "error", err)
	}
}

// MyJSONResponse sends a JSON response with the given HTTP status.
func MyJSONResponse(
	w http.ResponseWriter,
	status int,
	payload any,
) error {
	if status < 200 || status > 599 {
		return writeInternalError(
			w,
			fmt.Errorf("invalid HTTP status: %d", status),
		)
	}

	switch status {
	case http.StatusNoContent,
		http.StatusResetContent,
		http.StatusNotModified:

		if payload != nil {
			return writeInternalError(
				w,
				fmt.Errorf("status %d cannot contain a response body", status),
			)
		}

		w.WriteHeader(status)
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return writeInternalError(
			w,
			fmt.Errorf("marshal JSON response: %w", err),
		)
	}

	setJSONHeaders(w)
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("write JSON response: %w", err)
	}

	return nil
}

// Everything below is private to this package.

func decodeJSON(
	w http.ResponseWriter,
	r *http.Request,
	dst any,
) error {
	if r == nil {
		return errors.New("nil HTTP request")
	}

	if dst == nil {
		return errInvalidDecodeTarget
	}

	target := reflect.ValueOf(dst)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return errInvalidDecodeTarget
	}

	mediaType, _, err := mime.ParseMediaType(
		r.Header.Get("Content-Type"),
	)
	if err != nil ||
		(!strings.EqualFold(mediaType, "application/json") &&
			!strings.HasSuffix(strings.ToLower(mediaType), "+json")) {
		return errUnsupportedMediaType
	}

	if r.Body == nil {
		return errEmptyBody
	}

	if r.ContentLength > maxJSONBodyBytes {
		return errBodyTooLarge
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return fmt.Errorf(
				"%w: limit is %d bytes",
				errBodyTooLarge,
				maxJSONBodyBytes,
			)
		}

		return fmt.Errorf("read request body: %w", err)
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return errEmptyBody
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("%w: %v", errInvalidJSON, err)
	}

	var extra any
	err = decoder.Decode(&extra)

	if err == nil {
		return errMultipleJSONValues
	}

	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: %v", errInvalidJSON, err)
	}

	return nil
}

func writeDecodeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errUnsupportedMediaType):
		WriteJSONError(
			w,
			http.StatusUnsupportedMediaType,
			"unsupported_media_type",
			"Content-Type must be application/json",
		)

	case errors.Is(err, errBodyTooLarge):
		WriteJSONError(
			w,
			http.StatusRequestEntityTooLarge,
			"request_body_too_large",
			"Request body must not exceed 1 MiB",
		)

	case errors.Is(err, errEmptyBody),
		errors.Is(err, errInvalidJSON),
		errors.Is(err, errMultipleJSONValues):

		WriteJSONError(
			w,
			http.StatusBadRequest,
			"invalid_request_body",
			"Request body must contain one valid JSON value",
		)

	default:
		slog.Error("decode JSON request", "error", err)

		WriteJSONError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"Internal server error",
		)
	}
}

func setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func writeInternalError(w http.ResponseWriter, cause error) error {
	setJSONHeaders(w)
	w.WriteHeader(http.StatusInternalServerError)

	const fallback = `{"error":{"code":"internal_error","message":"Internal server error"}}`

	if _, err := io.WriteString(w, fallback); err != nil {
		return fmt.Errorf("%v; write fallback response: %w", cause, err)
	}

	return cause
}
