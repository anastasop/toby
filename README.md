
Toby is a tool to help you cope with backups on external removable disks.

Disks are not reliable and continuous backup is necessary. If you don't want to use a cloud backup service because of cost and privacy issues, the alternative is to use external usb disks. The problem with them is that everything is manual: You must copy files between them at regular intervals, because nothing is automated, sync them and maintain the directory structure yourself. The last issue is not as easy as it sounds. Eventually dirs, files and structure will not be synced any more and you won't be able to answer questions like:

- where is this file stored?
- is it corrupted?
- where have i store the pictures from my trip to Europe?
- what was the name of the pdf with my notes on healthy food?

Toby helps by scanning the disks and creating an sqlite3 database with information about the files to be used as index. It is easier to query this index to find the files you need and then mount the disks to get the files.

Two databases are created: The first one has the metadata of the files and the second contain thumbnails for images and the first page for pdfs. The schema is:

```
CREATE TABLE IF NOT EXISTS files (
  id INTEGER PRIMARY KEY,
  tag TEXT NOT NULL,  -- the tag, usually an identifier for the disk
  path TEXT NOT NULL, -- the file path
  sha1 TEXT,          -- sha1 hash of file
  size INTEGER,       -- the size of file
  mod_time TEXT,      -- the modification time of the file
  mime_type TEXT,     -- the mime type
  exif_time TEXT,     -- if an images and has exif data, this is Exif Date And Time
  error TEXT          -- if the file could not be read this is a humanized string for the error
);

CREATE TABLE IF NOT EXISTS thumbs.thumbnails (
  id INTEGER PRIMARY KEY,
  file_id INTEGER,
  thumbnail BLOB
);
```

## Installation

Toby is written in Go and tested with go1.12 on linux systems. It can be easily installed with the go tool

```
go get github.com/anastasop/toby
```

## Usage

toby -t backup -d summaries.db /mnt/c/snapshot
&nbsp;&nbsp;add the files rooted at /mnt/c/snapshot to the database file summaries.db tagged with backup

toby -t backup -d summaries.db -v /mnt/c /mnt/c/snapshot`
&nbsp;&nbsp;same as above but paths will be saved ad ./snapshot/path. Flag -v causes prefix /mnt/c to be stripped

toby -d summaries.db -s main
&nbsp;&nbsp;fuzzy search from paths matching main and report them

toby --schema
&nbsp;&nbsp;display the sql for the tables of the database

## License

Toby is licensed under the [AGPL](https://www.gnu.org/licenses/agpl-3.0.en.html) This is a requirement by [unidoc](https://unidoc.io/), the creators of [unipdf](https://github.com/unidoc/unipdf), one of the components of toby. You can check the above links for details on licensing.

If you don't have a unidoc license you can use toby but a watermark appears to the bottom of each page. Also it prints `Unlicensed copy of unidoc. To get rid of the watermark - Please get a license on https://unidoc.io`. This is not a major issue as pdfcovers is for personal use and the output is not excepted to be published. If however this is annoying you can [apply to unidoc](https://unidoc.io/pricing/) for a free license and use the `-lf` and `-ln` flag to hide the watermark.
