package apictx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type User interface {
	ID() string
}

type HandlerFunc func(http.ResponseWriter, *http.Request)
type ContextFunc func(ctx *Context) error

type ApiResponse struct {
	Code     int
	Response interface{}
}

type ApiErrorResponse struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
	Cause   error       `json:"-"`
}

// HttpError used to handle generic error for the context
type HttpError struct {
	err        error
	msg        string
	statusCode int
}

func NewHttpError(msg string, err error, statsuCode ...int) *HttpError {
	statusCode := http.StatusBadRequest

	if len(statsuCode) == 1 && statsuCode[0] >= 200 && statsuCode[0] <= 520 {
		statusCode = statsuCode[0]
	}
	return &HttpError{err: err, msg: msg, statusCode: statusCode}
}

func (e HttpError) Error() string {
	return e.msg
}

func (e HttpError) Cause() error {
	return e.err
}

func (e HttpError) Status() int {
	return e.statusCode
}

type Context struct {
	CurrentUser User
	writer      http.ResponseWriter
	request     *http.Request
}

func NewContext(w http.ResponseWriter, r *http.Request, user User) Context {
	return Context{
		CurrentUser: user,
		writer:      w,
		request:     r,
	}
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) Writer() http.ResponseWriter {
	return c.writer
}

func (c *Context) Bind(data interface{}) *HttpError {
	err := c.BindWithoutValidation(data)
	if err != nil {
		return NewHttpError("failed to read inputs", err, http.StatusBadRequest)
	}
	// Validate the data
	v := validator.New()
	err = v.Struct(data)
	if err != nil {
		var errMsgs []string
		for _, e := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, fmt.Sprintf("validation failed for %s", e.Field()))
		}
		return NewHttpError(
			fmt.Sprintf("validation error(s): %s", strings.Join(errMsgs, ", ")),
			nil,
			http.StatusBadRequest,
		)
	}
	return nil
}

func (c *Context) BindWithoutValidation(data interface{}) error {
	// Bind query parameters
	queryParams := c.request.URL.Query()
	err := c.BindQueryParams(data, queryParams)
	if err != nil {
		return err
	}

	// Bind request body
	contentType := c.request.Header.Get("Content-Type")
	if contentType == "application/json" {
		err = c.BindJSONBody(data, c.request.Body)
	} else {
		// Handle other content types like form data
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Context) BindQueryParams(data interface{}, params map[string][]string) error {
	val := reflect.ValueOf(data).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("query")
		if tag != "" {
			paramValues, ok := params[tag]
			if ok && len(paramValues) > 0 {
				paramValue := paramValues[0] // Use the first value
				switch field.Kind() {
				case reflect.String:
					field.SetString(paramValue)
				case reflect.Int:
					intValue, err := strconv.Atoi(paramValue)
					if err != nil {
						return fmt.Errorf("failed to convert parameter %s to int: %s", tag, err)
					}
					field.SetInt(int64(intValue))
					// Add cases for other types as needed
				}
			}
		}
	}

	return nil
}

func (c *Context) BindJSONBody(data interface{}, body io.Reader) error {
	err := json.NewDecoder(body).Decode(data)
	if err != nil {
		return fmt.Errorf("failed to decode JSON body: %s", err)
	}
	return nil
}

func (c *Context) JSON(code int, data interface{}) {
	statusCode := code
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	c.writer.Header().Set("Content-Type", "application/json;charset=utf-8")
	c.writer.WriteHeader(statusCode)
	json.NewEncoder(c.writer).Encode(data)
}

func Handler(c ContextFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// parse uer details or return 403

		ctx := Context{
			writer:  w,
			request: r,
		}

		err := c(&ctx)
		if err != nil {
			HandleError(w, r, err)
			return
		}
	}
}

func HandleError(w http.ResponseWriter, r *http.Request, err error, overRideStatusCode ...int) {
	var errRes ApiErrorResponse
	statusCode := http.StatusInternalServerError

	if len(overRideStatusCode) == 1 {
		statusCode = overRideStatusCode[0]
	}

	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		slog.Debug("api error: "+httpErr.Error(), "error", httpErr.Cause(), r.Method, r.URL)
		statusCode = httpErr.Status()
		errRes = ApiErrorResponse{0x6400, httpErr.Error(), httpErr.Cause()}
	} else {
		slog.Warn("internal error", "error", err, r.Method, r.URL)
		errRes = ApiErrorResponse{0x0, "Internal error", nil}
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errRes)
}
