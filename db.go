package main

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `-- this table is created in the primary db
CREATE TABLE IF NOT EXISTS files (
  id INTEGER PRIMARY KEY,
  tag TEXT NOT NULL,  -- the tag, usually an identifier for the disk
  path TEXT NOT NULL, -- the file path
  sha1 TEXT,          -- sha1 hash of file
  size INTEGER,       -- the size of file
  mod_time TEXT,      -- the modification time of the file
  mime_type TEXT,     -- the mime type
  exif_time TEXT,     -- if an images and has exif data, this is Exif Date And Time
  error TEXT,         -- if the file could not be read this is a humanized string for the error

  UNIQUE(tag, path, sha1)
);

CREATE INDEX IF NOT EXISTS files_mime_type ON files(mime_type);
CREATE INDEX IF NOT EXISTS files_tag ON files(tag);
`

const insertFileSql = `INSERT INTO files(tag, path, sha1, size, mod_time, mime_type, exif_time, error) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

const retrievePathsSql = `SELECT tag, path FROM files`

var (
	db *sql.DB

	insertFileStmt    *sql.Stmt
	retrievePathsStmt *sql.Stmt
)

func printSchema(w io.Writer) {
	fmt.Fprintln(w, schema)
}

func openDatabase(dbFile string) (err error) {
	if db, err = sql.Open("sqlite3", "file:"+dbFile); err != nil {
		return fmt.Errorf("Cannot open database %s: %s", dbFile, err)
	}

	if _, err = db.Exec(schema); err != nil {
		return fmt.Errorf("Cannot init database %s: %s", dbFile, err)
	}

	insertFileStmt, err = db.Prepare(insertFileSql)
	if err != nil {
		return fmt.Errorf("Cannot prepare statement to insert files: %s", err)
	}

	retrievePathsStmt, err = db.Prepare(retrievePathsSql)
	if err != nil {
		return fmt.Errorf("Cannot prepare statement to retrieve paths: %s", err)
	}

	return
}

func closeDatabase(dbFile string) (err error) {
	if insertFileStmt != nil {
		insertFileStmt.Close()
	}

	if retrievePathsStmt != nil {
		retrievePathsStmt.Close()
	}

	if db != nil {
		if err = db.Close(); err != nil {
			return fmt.Errorf("Cannot close database %s: %s", dbFile, err)
		}
	}

	return nil
}

func saveSummary(fs *fileSummary) error {
	var errStr sql.NullString
	if fs.err == nil {
		errStr = sql.NullString{Valid: false}
	} else {
		errStr = sql.NullString{Valid: true, String: fs.err.Error()}
	}
	_, err := insertFileStmt.Exec(
		fs.tag,
		fs.path,
		emptyStringToNil(fs.sha1),
		emptyIntToNil(fs.size),
		emptyTimeToNil(fs.modTime),
		emptyStringToNil(fs.mimeType),
		emptyTimeToNil(fs.exifTime),
		errStr)
	if err != nil {
		return fmt.Errorf("Cannot insert file %s: %s", fs.path, err)
	}

	return nil
}

type taggedPath struct {
	tag  string
	path string
}
type paths []taggedPath

func (p paths) Len() int {
	return len(p)
}

func (p paths) String(i int) string {
	return p[i].path
}

func retrievePaths() (paths, error) {
	rows, err := retrievePathsStmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	p := make(paths, 0)
	var tag, path sql.RawBytes
	for rows.Next() {
		rows.Scan(&tag, &path)
		p = append(p, taggedPath{string(tag), string(path)})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return p, nil
}

func emptyStringToNil(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func emptyIntToNil(s int64) interface{} {
	if s == 0 {
		return nil
	}
	return s
}

func emptyTimeToNil(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
