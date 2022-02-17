package core

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type pdfInfo struct {
	Title    string
	HtmlPath string
	PdfPath  string
}

func Html2pdf(htmlDir string, pdfDir string) (err error) {
	files, err := ioutil.ReadDir(htmlDir)
	if err != nil {
		return
	}

	ex, _ := os.Getwd()
	pdfs := make([]pdfInfo, 0)

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

	if err = os.MkdirAll(pdfDir, 0755); err != nil {
		return
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	for _, p := range pdfs {
		if _, err = os.Stat(p.PdfPath); err == nil {
			continue
		}
		var buf []byte
		if err = chromedp.Run(ctx, printToPDF("file:///"+p.HtmlPath, &buf)); err != nil {
			return err
		}

		if err = ioutil.WriteFile(p.PdfPath, buf, 0644); err != nil {
			return err
		}
		log.Println("生成：", p.PdfPath)

		time.Sleep(100 * time.Millisecond)
	}
	return
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
