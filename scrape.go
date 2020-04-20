package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"

	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/kelindar/loader"
	"github.com/stapelberg/godebiancontrol"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pelago/common"
)

const (
	CranProject = "https://cran.r-project.org/src/contrib/PACKAGES"
	MaxPackages = 50
)


func setUpDB() *sql.DB {
	database, _ := sql.Open("sqlite3", "./packages.db")
	statement, _ := database.Prepare(`CREATE TABLE IF NOT EXISTS package (name VARCHAR PRIMARY KEY,
		version VARCHAR,
		publicationTime BIGINT,
		title VARCHAR,
		description VARCHAR,
		authors VARCHAR,
		maintainers VARCHAR)`)
	statement.Exec()

	return database
}

func main() {
	fmt.Println("Scraping ...")
	dir, err := ioutil.TempDir(".", "packageDir-")
	if err != nil {
		fmt.Println("Error creating temp directory")
	}
	defer func() { _ = os.RemoveAll(dir) }()

	db := setUpDB()

	l := loader.New()

	data, err := l.Load(context.Background(), CranProject)
	if err != nil {
		fmt.Println("Error reading page")
	}

	rawString := string(data)
	packagesStr := strings.Split(rawString, "\n\n")

	total := 0

	packages := make([]common.Package, MaxPackages)

	for i, packageStr := range packagesStr {
		packageData := parsePackageStr(packageStr)
		err := downloadPackageZip(l, dir, packageData)
		if err != nil {
			fmt.Printf("Error downloading package", err)
			continue
		}

		packageData = parseDownloadedFile(dir, packageData.Name)
		packages[i] = packageData

		total += 1
		if total >= MaxPackages {
			break
		}
	}

	fmt.Println(packages)
	flushToDB(db, packages)
}

func flushToDB(db *sql.DB, packages []common.Package) {
	for _, packageData := range packages {
		statement, _ := db.Prepare(`INSERT INTO package (name, version, publicationTime, title, description, authors, maintainers)
VALUES (?, ?, ?, ?, ?, ?, ?)`)
		statement.Exec(packageData.Name,
			packageData.Version,
			packageData.PublicationTime,
			packageData.Title,
			packageData.Description,
			packageData.Authors,
			packageData.Maintainers)
	}
}

func parsePackageStr(packageStr string) common.Package {
	const NameLine = 0
	const VersionLine = 1

	packageInfo := strings.Split(packageStr, "\n")
	// Assume that they exist
	name := strings.Split(packageInfo[NameLine], " ")[1]
	version := strings.Split(packageInfo[VersionLine], " ")[1]

	return common.Package{
		Name: name,
		Version: version,
	}
}

func downloadPackageZip(l *loader.Loader, directory string, packageData common.Package) error {
	url := fmt.Sprintf("https://cran.r-project.org/src/contrib/%s_%s.tar.gz", packageData.Name, packageData.Version)
	fmt.Println(url)
	zippedBytes, err := l.Load(context.Background(), url)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s/%s.tar.gz", directory, packageData.Name), zippedBytes, 0644)
	if err != nil {
		fmt.Println("Error Writing file", err)
		return err
	}

	time.Sleep(1 * time.Second)

	return nil
}

func parseDownloadedFile(dir, packageName string) common.Package {
	f, err := os.Open(fmt.Sprintf("%s/%s.tar.gz", dir, packageName))
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		fmt.Println(err)
	}

	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
		}

		name := header.Name
		if name == fmt.Sprintf("%s/DESCRIPTION", packageName) {
			paragraphs, err := godebiancontrol.Parse(tarReader)
			if err != nil {
				fmt.Println("parsing err", err)
			}
			if len(paragraphs) <= 0 {
				fmt.Println("Data not found")
			}

			f.Close()
			unix, err := time.Parse("2006-01-02 15:04:05", paragraphs[0]["Date/Publication"])
			if err != nil {
				unix, _ = time.Parse("2006-01-02 15:04:05 MST", paragraphs[0]["Date/Publication"])
			}
			return common.Package{
				Name: packageName,
				Version: paragraphs[0]["Version"],
				Authors: paragraphs[0]["Author"],
				Maintainers: paragraphs[0]["Maintainer"],
				Title: paragraphs[0]["Title"],
				PublicationTime: unix.Unix(),
				Description: paragraphs[0]["Description"],
			}
		}
	}

	return common.Package{}
}
