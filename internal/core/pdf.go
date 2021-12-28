package core

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	pdf "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/zag13/go-utils/ujson"
)

func MergePdf(pdfDir string, pdfName string) (err error) {
	files, err := ioutil.ReadDir(pdfDir)
	if err != nil {
		return
	}

	inFiles := make([]string, 0, len(files))
	for _, f := range files {
		names := strings.Split(f.Name(), ".")
		if names[len(names)-1] == "pdf" {
			inFiles = append(inFiles, pdfDir+f.Name())
		}
	}

	var (
		originalPdf = pdfName + "_original.pdf"
		bookmarkPdf = pdfName + "_bookmark.pdf"
		pdfMeta     = pdfDir + "meta.json"
	)

	if _, err = os.Stat(originalPdf); errors.Is(err, os.ErrNotExist) {
		err = pdf.MergeCreateFile(inFiles, originalPdf, nil)
		if err != nil {
			return err
		}
	}

	var bookmarks []pdfcpu.Bookmark
	if _, err = os.Stat(pdfMeta); err == nil {
		err = ujson.JsonLoad(pdfMeta, &bookmarks)
		if err != nil {
			return
		}
		if err = pdf.AddBookmarksFile(originalPdf, bookmarkPdf, bookmarks, nil); err != nil {
			return
		}
	} else {
		if err = createPdfMeta(inFiles, pdfMeta); err != nil {
			return
		}
		err = ujson.JsonLoad(pdfMeta, &bookmarks)
		if err != nil {
			return
		}
		if err = pdf.AddBookmarksFile(originalPdf, bookmarkPdf, bookmarks, nil); err != nil {
			return
		}
	}

	return
}

func createPdfMeta(files []string, metaFile string) error {
	ff := []*os.File(nil)
	for _, f := range files {
		f, err := os.Open(f)
		if err != nil {
			return err
		}

		ff = append(ff, f)
	}

	f, err := os.Create(metaFile)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			f.Close()
			for _, f := range ff {
				f.Close()
			}
			return
		}
		if err = f.Close(); err != nil {
			return
		}
		for _, f := range ff {
			if err = f.Close(); err != nil {
				return
			}
		}
	}()

	var bookmarks []pdfcpu.Bookmark
	var totalPage = 1

	for _, f := range ff {
		names := strings.Split(f.Name(), "_")
		name := strings.TrimSuffix(names[1], ".pdf")
		bookmarks = append(bookmarks, pdfcpu.Bookmark{
			Title:    strings.Replace(name, "-", " ", -1),
			PageFrom: totalPage,
		})
		fPage, _ := pdf.PageCount(f, nil)
		totalPage += fPage
	}

	v, err := json.Marshal(bookmarks)
	if err != nil {
		return err
	}

	_, err = f.Write(v)
	if err != nil {
		return err
	}

	return nil
}
