package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
)

func ParseId(r *http.Request) (int64, error) {
	// First try query parameter (useful for tests and simple clients)
	if q := r.URL.Query().Get("id"); q != "" {
		rawId := q
		id, err := strconv.ParseInt(rawId, 10, 64)
		if err != nil {
			return 0, errors.New("id must be a valid integer")
		}
		return id, nil
	}

	rawId := r.PathValue("id")
	if rawId == "" {
		return 0, errors.New("id is required")
	}

	id, err := strconv.ParseInt(rawId, 10, 64)
	if err != nil {
		return 0, errors.New("id must be a valid integer")
	}

	return id, nil
}

const defaultMaxBodyBytes = 1024 * 1024 // 1MB

func ParseJSON[T any](r *http.Request) (*T, error) {
	return ParseJSONWithLimit[T](r, defaultMaxBodyBytes)
}

func ParseJSONWithLimit[T any](r *http.Request, maxBytes int64) (*T, error) {
	var data T

	limited := io.LimitReader(r.Body, maxBytes)
	dec := json.NewDecoder(limited)
	dec.DisallowUnknownFields()

	err := dec.Decode(&data)
	if errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	if err != nil {
		return nil, errors.New("invalid request body")
	}

	defer r.Body.Close()

	return &data, nil
}

func ParseMultipartForm[T any](r *http.Request, maxMemory int64, fileField string) (*T, multipart.File, *multipart.FileHeader, error) {

	var result T

	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return nil, nil, nil, err
	}

	val := reflect.ValueOf(&result).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Tag.Get("form")
		if fieldName == "" {
			fieldName = field.Name
		}

		formValue := r.FormValue(fieldName)
		if formValue == "" {
			continue
		}

		switch val.Field(i).Kind() {
		case reflect.String:
			val.Field(i).SetString(formValue)
		case reflect.Int, reflect.Int64:
			if intVal, err := strconv.ParseInt(formValue, 10, 64); err == nil {
				val.Field(i).SetInt(intVal)
			}
		case reflect.Float64:
			if floatVal, err := strconv.ParseFloat(formValue, 64); err == nil {
				val.Field(i).SetFloat(floatVal)
			}
			// Add more types as needed
		}
	}

	file, fileHeader, err := r.FormFile(fileField)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return &result, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	return &result, file, fileHeader, nil
}
