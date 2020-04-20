package main

import (
	"database/sql"
	"fmt"
	"github.com/pelago/storage"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type APIHandler struct {
	storage *storage.PackageStorage
}

func NewAPIHandler(storage *storage.PackageStorage) *APIHandler {
	return &APIHandler{
		storage: storage,
	}
}

func (api *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	packages := api.storage.Search(r.URL.Query()["name"][0])
	fmt.Fprintf(w, "%s", packages)
}

func main() {
	db, _ := sql.Open("sqlite3", "./packages.db")
	packagesStorage := storage.NewPackageStorage(db)

	apiHandler := NewAPIHandler(packagesStorage)

	http.Handle("/search", apiHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
