package storage

import (
	"database/sql"
	"fmt"

	"github.com/pelago/common"
	_ "github.com/mattn/go-sqlite3"
)

// PackageStorage ...
type PackageStorage struct {
	db *sql.DB
}

// NewStorage ...
func NewPackageStorage(db *sql.DB) *PackageStorage {
	return &PackageStorage{
		db: db,
	}
}

// Search returns list of packages
func (s *PackageStorage) Search(name string) []common.Package {
	packages := []common.Package{}
	rows, err := s.db.Query(`SELECT name, version FROM package WHERE name LIKE '%' || $1 || '%'`, name)
	if err != nil {
		fmt.Println("error querying", err)
		return packages
	}
	var nameData string
	var versionData string
	for rows.Next() {
		rows.Scan(&nameData, &versionData)
		packageData := common.Package{
			Name: nameData,
			Version: versionData,
		}
		packages = append(packages, packageData)
	}

	return packages
}
