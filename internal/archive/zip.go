package archive

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	archive "github.com/matthewmueller/go-archive"
	"github.com/pkg/errors"
)

var modtime = time.Date(2018, 9, 14, 12, 18, 0, 0, time.UTC)

var transform = archive.TransformFunc(func(r io.Reader, i os.FileInfo) (io.Reader, os.FileInfo) {
	i = archive.Info{
		Name: i.Name(),
		Size: i.Size(),
		Mode: i.Mode() | 0555,
		// Consistent modtime to have a consistent hash
		Modified: modtime,
		Dir:      i.IsDir(),
	}.FileInfo()
	return r, i
})

// Zip the given `dir`.
func Zip(dir string) (io.ReadCloser, *archive.Stats, error) {
	gitignore, err := read(".gitignore")
	if err != nil {
		return nil, nil, errors.Wrap(err, "reading .gitignore")
	}
	defer gitignore.Close()

	npmignore, err := read(".npmignore")
	if err != nil {
		return nil, nil, errors.Wrap(err, "reading .npmignore")
	}
	defer npmignore.Close()

	r := io.MultiReader(
		strings.NewReader(".*\n"),
		gitignore,
		strings.NewReader("\n"),
		npmignore,
		strings.NewReader("\n!node_modules\n"),
		strings.NewReader("\n!main\n!_proxy.js\n!byline.js\n"))

	filter, err := archive.FilterPatterns(r)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing ignore patterns")
	}

	buf := new(bytes.Buffer)
	zip := archive.NewZip(buf).
		WithFilter(filter).
		WithTransform(transform)

	if err := zip.Open(); err != nil {
		return nil, nil, errors.Wrap(err, "opening")
	}

	if err := zip.AddDir(dir); err != nil {
		return nil, nil, errors.Wrap(err, "adding dir")
	}

	if err := zip.Close(); err != nil {
		return nil, nil, errors.Wrap(err, "closing")
	}

	return ioutil.NopCloser(buf), zip.Stats(), nil
}

// read file.
func read(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)

	if os.IsNotExist(err) {
		return ioutil.NopCloser(bytes.NewReader(nil)), nil
	}

	if err != nil {
		return nil, err
	}

	return f, nil
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}
