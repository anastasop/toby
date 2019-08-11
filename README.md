
Toby is a tool to help you cope with backups on external removable disks.

Disks are not reliable and continuous backup is necessary. If you don't want to use a cloud backup service because of cost and privacy issues, the alternative is to use external usb disks. The problem with them is that everything is manual: You must copy files between them at regular intervals, because nothing is automated, sync them and maintain the directory structure yourself. The last issue is not as easy as it sounds. Eventually directories, files and structure will not be synced any more and you won't be able to answer questions like:

- where is this file stored?
- is it corrupted?
- where have i store the pictures from my trip to Europe?
- what was the name of the pdf with my notes on healthy food?

Toby helps by scanning the disks and creating an sqlite3 database with information about the files to be used as index. It is easier to query this index to find the files you need and then mount the disks to get the files.

The schema is:

```
CREATE TABLE IF NOT EXISTS files (
  id INTEGER PRIMARY KEY,
  tag TEXT NOT NULL,  -- the tag, an identifier for an external disk
  path TEXT NOT NULL, -- the file path
  sha1 TEXT,          -- sha1 hash of file
  size INTEGER,       -- the size of file
  mod_time TEXT,      -- the modification time of the file
  mime_type TEXT,     -- the mime type
  exif_time TEXT,     -- if an images and has exif data, this is Exif Date And Time
  error TEXT          -- if the file could not be read this is a humanized string for the error
);
```

## Installation

Toby is written in Go and tested with go1.12 on linux systems. It can be easily installed with the go tool

```
go get github.com/anastasop/toby
```

## Usage
```
#add the files rooted at /mnt/c/snapshot to the database file summaries.db and tag with backup
toby -t backup -d summaries.db /mnt/c/snapshot

#same as above but paths will be saved ad ./snapshot/path. Flag -v causes prefix /mnt/c to be stripped
toby -t backup -d summaries.db -v /mnt/c /mnt/c/snapshot`

#fuzzy search from paths matching main and report them
toby -d summaries.db -s main

#display the sql for the tables of the database
toby --schema
```

## License

Toby is licensed under the [AGPL](https://www.gnu.org/licenses/agpl-3.0.en.html)

