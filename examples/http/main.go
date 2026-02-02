package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/tpauling/handgover"
)

type GetCustomersByIDRequest struct {
	CustomerID int `path:"id"`
}

type GetCustomersList struct {
	Count int `query:"count"`
}

type PostNewCustomer struct {
	Body struct {
		Name string `json:"name"`
	} `body:"application/json"`
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/customers/{id}", func(w http.ResponseWriter, r *http.Request) {
		var parsedRequest GetCustomersByIDRequest
		if err := ParseFrom(r, &parsedRequest); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println("Customer ID:", parsedRequest.CustomerID) // Customer ID: 123
	})
	mux.HandleFunc("/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			var parsedRequest GetCustomersList
			if err := ParseFrom(r, &parsedRequest); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Println("Count:", parsedRequest.Count) //  Count: 100
		}

		if r.Method == http.MethodPost {
			var parsedRequest PostNewCustomer
			if err := ParseFrom(r, &parsedRequest); err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			log.Println("Customer:", parsedRequest.Body) // Customer: {John Doe}
		}
	})

	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/customers/123", nil))
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/customers?count=100", nil))
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/customers", bytes.NewBufferString(`{"name" : "John Doe"}`)))
}

func ParseFrom(req *http.Request, obj interface{}) error {
	sources := []handgover.Source{
		handgover.NewSource(
			"body",
			func(field string) (handgover.Valuer, error) {
				defer req.Body.Close()
				data, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				return handgover.Value(string(data)), nil
			},
		),
		handgover.NewSource(
			"query",
			func(field string) (handgover.Valuer, error) {
				return handgover.Value(req.URL.Query()[field]...), nil
			},
		),
		handgover.NewSource(
			"path",
			func(field string) (handgover.Valuer, error) {
				return handgover.Value(req.PathValue(field)), nil
			},
		),
	}
	return handgover.From(sources).To(obj)
}
