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
	ardBlogDomain   = "www.ardanlabs.com"
	ardBlogXML      = "https://www.ardanlabs.com/blog/index.xml"
	ardBlogHtml     = "./storage/ardanlabs/html/"
	ardBlogHtmlMeta = "./storage/ardanlabs/html/meta.json"
	ardBlogPdf      = "./storage/ardanlabs/pdf/"
	ardBlogPdfName  = "./storage/ardanlabs/ardanlabs"
	ardBlogs        = make([]ardBlog, 0)
	ardBlogsOld     = make([]ardBlog, 0)
)

type ardBlog struct {
	Title string
	Url   string
	Href  string
	Date  string
}

func Ardanlabs(action string) {
	for i := 0; i < 3; i++ {
		if action[i]&1 == 1 {
			switch i {
			case 0:
				fmt.Println("ArdBlog: download html")
				dlArdBlog()
			case 1:
				fmt.Println("ArdBlog: convert html to pdf")
				core.Html2pdf(ardBlogHtml, ardBlogPdf)
			case 2:
				fmt.Println("ArdBlog: merged into a pdf with bookmarks")
				core.MergePdf(ardBlogPdf, ardBlogPdfName)
			}
		}
	}
}

func dlArdBlog() {
	initResource(ardBlogHtml, ardBlogHtmlMeta)
	updateArdMeta()

	c := colly.NewCollector(
		colly.AllowedDomains(ardBlogDomain),
		colly.MaxDepth(1),
	)

	blogCollector := c.Clone()

	c.OnXML("//rss/channel/item", func(e *colly.XMLElement) {
		if searchArdBlog(e.ChildText("link"), ardBlogsOld) != -1 {
			return
		}

		fmt.Printf("URL: %s\n", e.ChildText("link"))
		blogCollector.Visit(e.ChildText("link"))
	})

	blogCollector.OnHTML("html", func(e *colly.HTMLElement) {
		e.DOM.Find("script").Remove()
		e.DOM.Find("header[class='blog-header']").Remove()
		e.DOM.Find("section[class='ribbon-slider']").Remove()
		e.DOM.Find("section[class='alert-wrapper']").Remove()
		e.DOM.Find("div[class='nav-wrap affix-top']").Remove()
		e.DOM.Find("div[class='blog-links']").Remove()
		e.DOM.Find("div[class='nav-mobile']").Remove()
		e.DOM.Find("div[class='author-flex']").Remove()
		e.DOM.Find("ul[class='breadcrumbs']").Remove()
		e.DOM.Find("div[class='container-fluid']").Remove()
		e.DOM.Find("footer").Remove()
		e.DOM.Find("div[class='bottom-footer']").Remove()
		e.DOM.Children().Find("img[src]").Each(func(_ int, selection *goquery.Selection) {
			val, ok := selection.Attr("src")
			if !ok {
				return
			}
			if strings.Contains(val, "http") {
				return
			}
			val = strings.Replace(val, "../../../", "", -1)
			selection.SetAttr("src", "./assets/"+val)
		})

		rawURL := e.Request.URL.String()
		u, err := url.Parse(rawURL)
		if err != nil {
			log.Fatalf("URL parse error: %s", err)
		}

		paths := strings.Split(u.Path, "/")
		name := paths[len(paths)-1]
		fileName := ardBlogHtml + getArdFileName(name[:len(name)-5]) + ".html"
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

		fileName := ardBlogHtml + "assets" + u.Path
		if _, err := os.Stat(fileName); err == nil {
			return
		}

		p := strings.Split(fileName, "/")
		dir := strings.Join(p[:len(p)-1], "/")
		os.MkdirAll(dir, os.ModePerm)

		f, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("Error creating file: %s", err)
		}
		defer f.Close()

		res, _ := http.Get(rawURL)
		defer res.Body.Close()
		io.Copy(f, res.Body)
	})

	c.Visit(ardBlogXML)
}

func updateArdMeta() {
	err := ujson.JsonLoad(ardBlogHtmlMeta, &ardBlogsOld)
	if err != nil {
		log.Printf("Error loading meta.json: %s", err)
	}

	c := colly.NewCollector(
		colly.AllowedDomains(ardBlogDomain),
	)

	c.OnXML("//rss/channel/item", func(e *colly.XMLElement) {
		hrefs := strings.Split(e.ChildText("link"), "/")
		href := hrefs[len(hrefs)-1]

		ardBlogs = append(ardBlogs, ardBlog{
			Title: e.ChildText("title"),
			Url:   e.ChildText("link"),
			Href:  strings.Replace(href, ".html", "", -1),
			Date:  e.ChildText("pubDate"),
		})
	})

	c.Visit(ardBlogXML)

	ujson.JsonDump(ardBlogHtmlMeta, ardBlogs)
}

func searchArdBlog(key string, data []ardBlog) (n int) {
	for i, datum := range data {
		if datum.Url == key {
			return i
		}
	}
	return -1
}

func getArdFileName(key string) (s string) {
	index := 0
	for i, blog := range ardBlogs {
		if blog.Href == key {
			index = i
			break
		}
	}
	return fmt.Sprintf("%03d_", len(ardBlogs)-index) + key
}
