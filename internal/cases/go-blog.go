package cases

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/zag13/dl/internal/core"
	"github.com/zag13/go-utils/ujson"
)

var (
	curDir         = ""
	goBlogDomain   = "go.dev"
	goBlogUrl      = "https://go.dev/blog/all"
	goBlogHtml     = "./storage/go.dev/html/"
	goBlogHtmlMeta = "./storage/go.dev/html/meta.json"
	goBlogPdf      = "./storage/go.dev/pdf/"
	goBlogPdfName  = "./storage/go.dev/go.dev"
	goBlogs        = make([]goBlog, 0)
	goBlogsOld     = make([]goBlog, 0)
)

type goBlog struct {
	Title   string
	Url     string
	Href    string
	Date    string
	Author  string
	Summary string
}

func GoBlog(action string) {
	for i := 0; i < 3; i++ {
		if action[i]&1 == 1 {
			switch i {
			case 0:
				fmt.Println("GoBlog: download html")
				dlGoBlog()
			case 1:
				fmt.Println("GoBlog: convert html to pdf")
				core.Html2pdf(goBlogHtml, goBlogPdf)
			case 2:
				fmt.Println("GoBlog: merged into a pdf with bookmarks")
				core.MergePdf(goBlogPdf, goBlogPdfName)
			}
		}
	}
}

func dlGoBlog() {
	initResource()
	updateMeta()

	c := colly.NewCollector(
		colly.AllowedDomains(goBlogDomain),
		colly.MaxDepth(1),
	)

	blogCollector := c.Clone()

	c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		rawURL := e.Request.AbsoluteURL(e.Attr("href"))
		u, _ := url.Parse(rawURL)
		paths := strings.Split(u.Path, "/")
		if paths[len(paths)-2] != "css" {
			return
		}

		fileName := goBlogHtml + "css/" + paths[len(paths)-1]
		if _, err := os.Stat(fileName); err == nil {
			return
		}

		f, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("Error creating file: %s", err)
		}
		defer f.Close()

		res, _ := http.Get(rawURL)
		defer res.Body.Close()
		io.Copy(f, res.Body)
	})

	c.OnHTML("div[id='blogindex']", func(e *colly.HTMLElement) {
		e.ForEach("p[class='blogtitle']", func(_ int, e *colly.HTMLElement) {
			href, ok := e.DOM.Find("a").Attr("href")
			if !ok {
				return
			}

			rawURL := e.Request.AbsoluteURL(href)

			u, _ := url.Parse(rawURL)
			paths := strings.Split(u.Path, "/")
			if searchBlog(paths[len(paths)-1], goBlogsOld) != -1 {
				return
			}

			fmt.Printf("URL: %s\n", u)
			blogCollector.Visit(e.Request.AbsoluteURL(rawURL))
		})
	})

	blogCollector.OnHTML("html", func(e *colly.HTMLElement) {
		e.DOM.Find("header[class='Site-header js-siteHeader']").Remove()
		e.DOM.Find("link[class='alternate']").Remove()
		e.DOM.Find("h1[class='small']").Remove()
		e.DOM.Find("div[class='Article prevnext']").Remove()
		e.DOM.Find("footer[class='Site-footer']").Remove()
		e.DOM.Find("aside[class='NavigationDrawer js-header']").Remove()
		e.DOM.Children().Find("link[rel='stylesheet']").Each(func(_ int, selection *goquery.Selection) {
			val, ok := selection.Attr("href")
			if !ok {
				return
			}
			selection.SetAttr("href", "."+val)
		})
		e.DOM.Children().Find("img[src]").Each(func(_ int, selection *goquery.Selection) {
			val, ok := selection.Attr("src")
			if !ok {
				return
			}
			selection.SetAttr("src", "./assets/"+val)
		})

		rawURL := e.Request.URL.String()
		u, err := url.Parse(rawURL)
		if err != nil {
			log.Fatalf("URL parse error: %s", err)
		}

		paths := strings.Split(u.Path, "/")
		fileName := goBlogHtml + getFilename(paths[len(paths)-1]) + ".html"
		f, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("File create error: %s", err)
		}
		defer f.Close()

		ret, _ := e.DOM.Html()
		io.WriteString(f, ret)
	})

	blogCollector.OnHTML("img[src]", func(e *colly.HTMLElement) {
		rawURL := e.Request.AbsoluteURL(e.Attr("src"))
		u, err := url.Parse(rawURL)
		if err != nil {
			log.Fatalf("URL parse error: %v", err)
		}

		paths := strings.Split(u.Path, "/")
		if paths[1] != "blog" {
			return
		}

		_ = os.MkdirAll(goBlogHtml+"assets/"+paths[len(paths)-2], 0755)
		fileName := goBlogHtml + "assets/" + paths[len(paths)-2] + "/" + paths[len(paths)-1]
		if _, err := os.Stat(fileName); err == nil {
			return
		}

		f, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("Error creating file: %s", err)
		}
		defer f.Close()

		if paths[len(paths)-3] == "assets" {
			urls := strings.Split(rawURL, "/assets")
			rawURL = urls[0] + urls[1]
		}
		res, _ := http.Get(rawURL)
		defer res.Body.Close()
		io.Copy(f, res.Body)
	})

	c.Visit(goBlogUrl)
}

func initResource() {
	wd, _ := os.Getwd()
	wds := strings.Split(wd, "/")
	if wds[len(wds)-1] == "dl" {
		curDir = wd + "/"
	} else {
		curDir = strings.TrimSuffix(wd, "cmd")
	}

	goBlogHtml = curDir + goBlogHtml
	goBlogHtmlMeta = curDir + goBlogHtmlMeta

	if err := os.MkdirAll(goBlogHtml, 0755); err != nil {
		log.Fatalf("Error creating directory: %s", err)
	}
	os.MkdirAll(goBlogHtml+"css/", 0755)
	os.MkdirAll(goBlogHtml+"assets/", 0755)
}

func updateMeta() {
	err := ujson.JsonLoad(goBlogHtmlMeta, &goBlogsOld)
	if err != nil {
		log.Printf("Error loading meta.json: %s", err)
	}

	c := colly.NewCollector(
		colly.AllowedDomains(goBlogDomain),
		colly.MaxDepth(1),
	)

	c.OnHTML("div[id='blogindex']", func(e *colly.HTMLElement) {
		e.ForEach("p[class='blogtitle']", func(_ int, e *colly.HTMLElement) {
			u, ok := e.DOM.Find("a").Attr("href")
			if !ok {
				return
			}
			paths := strings.Split(u, "/")
			goBlogs = append(goBlogs, goBlog{
				Title:   e.DOM.Find("a").Text(),
				Url:     e.Request.AbsoluteURL(u),
				Href:    paths[len(paths)-1],
				Date:    e.DOM.Find("span[class='date']").Text(),
				Author:  e.DOM.Find("span[class='author']").Text(),
				Summary: e.DOM.Next().Text(),
			})
		})
	})

	c.Visit(goBlogUrl)

	ujson.JsonDump(goBlogHtmlMeta, goBlogs)
}

func searchBlog(key string, data []goBlog) (n int) {
	for i, datum := range data {
		if datum.Href == key {
			return i
		}
	}
	return -1
}

func getFilename(key string) (s string) {
	index := 0
	for i, blog := range goBlogs {
		if blog.Href == key {
			index = i
			break
		}
	}
	return fmt.Sprintf("%03d_", len(goBlogs)-index) + key
}
