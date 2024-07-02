# API Context Package

## Overview

The `apictx` package provides a set of utilities to streamline handling HTTP requests and responses in a Go web application. It includes features for request context management, error handling, validation, and more.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [Creating a Context](#creating-a-context)
  - [Binding Request Data](#binding-request-data)
  - [Returning JSON Responses](#returning-json-responses)
  - [Error Handling](#error-handling)
- [Examples](#examples)
- [Contributing](#contributing)

## Installation

To use the `apictx` package, you need to install it first. Add it to your project using `go get`:

```bash
go get github.com/sivsivsree/apictx
```

## Usage

### Creating a Context

The `Context` struct is used to manage request and response objects, as well as the current user. Create a new context by using the `NewContext` function:

```go
func NewContext(w http.ResponseWriter, r *http.Request, user User) Context {
    return Context{
        CurrentUser: user,
        writer:      w,
        request:     r,
    }
}
```

### Binding Request Data

The `Context` struct provides methods to bind request data to Go structs. The `Bind` method binds and validates the request data:

```go
func (c *Context) Bind(data interface{}) *HttpError
```

To bind without validation, use the `BindWithoutValidation` method:

```go
func (c *Context) BindWithoutValidation(data interface{}) error
```

### Returning JSON Responses

The `Context` struct provides a method to send JSON responses:

```go
func (c *Context) JSON(code int, data interface{})
```

### Error Handling

The package includes an `HttpError` struct for handling HTTP errors:

```go
type HttpError struct {
    err        error
    msg        string
    statusCode int
}

func NewHttpError(msg string, err error, statsuCode ...int) *HttpError
```

Use the `HandleError` function to handle errors in your handlers:

```go
func HandleError(w http.ResponseWriter, r *http.Request, err error, overRideStatusCode ...int)
```

### Handler Wrapper

The `Handler` function wraps your context function, making it compatible with `http.HandlerFunc`:

```go
func Handler(c ContextFunc) http.HandlerFunc
```

## Examples

Here are a few examples to help you get started:

### Basic Usage

```go
package main

import (
    "net/http"
    "github.com/sivsivsree/apictx"
)

func main() {
    http.HandleFunc("/example", apictx.Handler(ExampleHandler))
    http.ListenAndServe(":8080", nil)
}

func ExampleHandler(ctx *apictx.Context) error {
    var data struct {
        Name string `query:"name" validate:"required"`
        Age  int    `query:"age" validate:"gte=0"`
    }

    if err := ctx.Bind(&data); err != nil {
        return err
    }

    ctx.JSON(http.StatusOK, data)
    return nil
}
```

### Error Handling

```go
func ExampleHandler(ctx *apictx.Context) error {
    return apictx.NewHttpError("an error occurred", errors.New("example error"), http.StatusInternalServerError)
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
