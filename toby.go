package main

import (
	"crypto/sha1"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/sahilm/fuzzy"
)

var (
	errStat     = errors.New("Stat")
	errOpen     = errors.New("Open")
	errRead     = errors.New("Read")
	errSeek     = errors.New("Seek")
	errExif     = errors.New("Exif")
	errExifTime = errors.New("ExifTime")
)

type fileSummary struct {
	tag      string
	path     string
	sha1     string
	size     int64
	modTime  time.Time
	mimeType string
	exifTime time.Time
	err      error
}

func newFileSummary(tag, path string, info os.FileInfo) *fileSummary {
	fs := &fileSummary{tag: tag, path: normalizePath(path)}

	if info == nil {
		fs.err = errStat
		return fs
	} else {
		fs.size = info.Size()
		fs.modTime = info.ModTime()
	}

	fin, err := os.Open(path)
	if err != nil {
		fs.err = errOpen
		return fs
	}
	defer fin.Close()

	hdr, err := ioutil.ReadAll(&io.LimitedReader{fin, 4096})
	if err != nil {
		fs.err = errRead
		return fs
	}
	fs.mimeType = http.DetectContentType(hdr)

	if err := addSha1(fs, fin); err != nil {
		return fs
	}

	if strings.HasPrefix(fs.mimeType, "image") {
		addExifMetadata(fs, fin)
	}

	return fs
}

func addSha1(fs *fileSummary, rs io.ReadSeeker) error {
	if _, err := rs.Seek(0, os.SEEK_SET); err != nil {
		fs.err = errSeek
		return err
	}

	h := sha1.New()
	if _, err := io.Copy(h, rs); err != nil {
		fs.err = errRead
		return err
	}

	fs.sha1 = fmt.Sprintf("%x", h.Sum(nil))
	return nil
}

func addExifMetadata(fs *fileSummary, rs io.ReadSeeker) {
	if _, err := rs.Seek(0, os.SEEK_SET); err != nil {
		fs.err = errSeek
		return
	}

	if ex, err := exif.Decode(rs); err == nil {
		if m, err := ex.DateTime(); err == nil {
			fs.exifTime = m
		} else {
			fs.err = errExifTime
		}
	} else {
		fs.err = errExif
	}
}

var usageMessage = `usage: toby -t tag -d dbfile [options] [<dir> ...]

Walks the directories recursively and for each regular file
write a summary to an sqlite3 database. The summary contains
the path name, the file type, update times and a tag for each
file which is used to differentiate same paths from different
root directories.

Examples:

toby -t backup -d summaries.db /mnt/c/snapshot
  add the files rooted at /mnt/c/snapshot to the database file summaries.db and tag with backup

toby -t backup -d summaries.db -v /mnt/c /mnt/c/snapshot
  same as above but paths will be saved as ./snapshot/path. Flag -v causes prefix /mnt/c to be stripped

toby -d summaries.db -s main
  fuzzy search from paths matching main and report them

toby --schema
  display the sql for the tables of the database

Flags:
`

func usage() {
	fmt.Fprintf(os.Stderr, usageMessage)
	flag.PrintDefaults()
	os.Exit(2)
}

var dbFile = flag.String("d", "", "the sqlite3 database file. If needed it is created and initialized with the schema")
var tag = flag.String("t", "", "the tag for paths. Used to differentiate same paths from different origins ex archives from different disks")
var volume = flag.String("v", "", "a prefix to strip from paths before saving in the database. Usually the mount point of a disk")
var schemaOnly = flag.Bool("schema", false, "print the sqlite3 schema and exit")
var search = flag.String("s", "", "fuzzy search the database for paths matching the argument")

func normalizePath(path string) string {
	if *volume == "" {
		return path
	}

	s, err := filepath.Rel(*volume, path)
	if err != nil {
		log.Printf("path %s: failed to normalize: %s", path, err)
		return path
	}
	return filepath.ToSlash(s)
}

func scanDir(tag, root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Failed to stat %s: %s", path, err)
			return nil
		}

		if info == nil && err == nil {
			log.Fatal("path %s: info is nil but err is not", path)
		}

		if _, elem := filepath.Split(path); elem != "" {
			// Skip "hidden" files or directories.
			if elem[0] == '.' {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.Mode().IsRegular() {
			if err := saveSummary(newFileSummary(tag, path, info)); err != nil {
				log.Println(err)
			}
		}

		return nil
	})
	return err
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("toby: ")
	flag.Usage = usage
	flag.Parse()

	if *schemaOnly {
		printSchema(os.Stdout)
		os.Exit(1)
	}

	if *dbFile == "" || *tag == "" {
		usage()
	} else {
		if err := openDatabase(*dbFile); err != nil {
			log.Fatal(err)
		}
		defer closeDatabase(*dbFile)
	}

	if *search != "" {
		paths, err := retrievePaths()
		if err != nil {
			log.Fatal(err)
		}

		matches := fuzzy.FindFrom(*search, paths)
		for _, match := range matches {
			path := paths[match.Index]
			fmt.Println(path.tag, path.path)
		}
		os.Exit(0)
	}

	for _, root := range flag.Args() {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			log.Println("Failed to make absolute path for %q: %s", root, err)
		} else if err := scanDir(*tag, absRoot); err != nil {
			log.Println("Failed to scan directory %s: %s", absRoot, err)
		}
	}
}
