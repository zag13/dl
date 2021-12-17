package main

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
	"github.com/zag13/go-utils/ujson"
)

var (
	destDomain = "go.dev"
	destUrl    = "https://go.dev/blog/all"
	htmlMeta   = "../../storage/html/go.dev/meta.json"
	htmlDir    = "../../storage/html/go.dev/"
	blogs      = make([]blog, 0)
	oldBlogs   = make([]blog, 0)
)

type blog struct {
	Title   string
	Href    string
	Date    string
	Author  string
	Summary string
}

func dlhtml() {
	initResource()

	c := colly.NewCollector(
		colly.AllowedDomains(destDomain),
		colly.MaxDepth(1),
	)

	blogCollector := colly.NewCollector(
		colly.AllowedDomains(destDomain),
		colly.MaxDepth(1),
	)

	c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		rawURL := e.Request.AbsoluteURL(e.Attr("href"))
		u, _ := url.Parse(rawURL)
		paths := strings.Split(u.Path, "/")
		if paths[len(paths)-2] != "css" {
			return
		}

		fileName := htmlDir + "css/" + paths[len(paths)-1]
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
			if searchBlog(paths[len(paths)-1], oldBlogs) != -1 {
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
		prefixInt := len(blogs) - searchBlog(paths[len(paths)-1], blogs) - 1
		fileName := htmlDir + fmt.Sprintf("%03d_", prefixInt) + paths[len(paths)-1] + ".html"
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

		_ = os.MkdirAll(htmlDir+"assets/"+paths[len(paths)-2], 0755)
		fileName := htmlDir + "assets/" + paths[len(paths)-2] + "/" + paths[len(paths)-1]
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

	c.Visit(destUrl)
}

func initResource() {
	err := os.MkdirAll(htmlDir, 0755)
	if err != nil {
		log.Fatalf("Error creating directory: %s", err)
	}
	_ = os.MkdirAll(htmlDir+"css/", 0755)
	_ = os.MkdirAll(htmlDir+"assets/", 0755)

	err = ujson.JsonLoad(htmlMeta, &oldBlogs)
	if err != nil {
		log.Printf("Error loading meta.json: %s", err)
	}

	c := colly.NewCollector(
		colly.AllowedDomains(destDomain),
		colly.MaxDepth(1),
	)

	c.OnHTML("div[id='blogindex']", func(e *colly.HTMLElement) {
		e.ForEach("p[class='blogtitle']", func(_ int, e *colly.HTMLElement) {
			u, ok := e.DOM.Find("a").Attr("href")
			if !ok {
				return
			}
			paths := strings.Split(u, "/")
			blogs = append(blogs, blog{
				Title:   e.DOM.Find("a").Text(),
				Href:    paths[len(paths)-1],
				Date:    e.DOM.Find("span[class='date']").Text(),
				Author:  e.DOM.Find("span[class='author']").Text(),
				Summary: e.DOM.Next().Text(),
			})
		})
	})

	c.Visit(destUrl)

	ujson.JsonDump(htmlMeta, blogs)
}

func searchBlog(key string, data []blog) (n int) {
	for i, datum := range data {
		if datum.Href == key {
			return i
		}
	}
	return -1
}
