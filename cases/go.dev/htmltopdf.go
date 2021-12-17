package main

import (
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var (
	pdfDir = "../../storage/pdf/go.dev/"
	pdfs   = make([]pdfInfo, 0)
)

type pdfInfo struct {
	Title    string
	HtmlPath string
	PdfPath  string
}

func htmltopdf() {
	initHtmlToPdf()

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	for _, p := range pdfs {
		if _, err := os.Stat(p.PdfPath); err == nil {
			log.Println("已经存在：", p.PdfPath)
			continue
		}
		// capture pdf
		var buf []byte
		if err := chromedp.Run(ctx, printToPDF("file:///"+p.HtmlPath, &buf)); err != nil {
			log.Fatal(err)
		}

		if err := ioutil.WriteFile(p.PdfPath, buf, 0644); err != nil {
			log.Fatal(err)
		}
		log.Println("生成：", p.PdfPath)
	}
}

func initHtmlToPdf() {
	err := os.MkdirAll(pdfDir, 0755)
	if err != nil {
		log.Fatalf("Error creating directory: %s", err)
	}

	ex, err := os.Getwd()
	if err != nil {
		log.Fatalf("os.Getwd err: %v", err)
	}

	files, err := ioutil.ReadDir(htmlDir)
	if err != nil {
		log.Fatalf("read dir error: %v", err)
	}

	for _, f := range files {
		names := strings.Split(f.Name(), ".")
		if names[len(names)-1] == "html" {
			pdfs = append(pdfs, pdfInfo{
				Title:    strings.TrimSuffix(f.Name(), ".html"),
				HtmlPath: ex + "/" + htmlDir + f.Name(),
				PdfPath:  pdfDir + strings.TrimSuffix(f.Name(), ".html") + ".pdf",
			})
		}
	}
}

// print a specific pdf page.
func printToPDF(urlstr string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(true).
				WithMarginLeft(0.2).WithMarginRight(0.2).
				Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}
