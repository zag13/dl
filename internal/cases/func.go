package cases

import (
	"log"
	"os"
	"strings"
)

func initResource(blogHtml, blogHtmlMeta string) {
	wd, _ := os.Getwd()
	wds := strings.Split(wd, "/")
	if wds[len(wds)-1] == "dl" {
		curDir = wd + "/"
	} else {
		curDir = strings.TrimSuffix(wd, "cmd")
	}

	blogHtmlMeta = curDir + blogHtmlMeta

	if _, err := os.Stat(blogHtmlMeta); os.IsNotExist(err) {
		f, err := os.Create(blogHtmlMeta)
		if err != nil {
			log.Fatalf("Error creating meta file: %s", err)
		}
		defer f.Close()
	}

	blogHtml = curDir + blogHtml

	if err := os.MkdirAll(blogHtml, 0755); err != nil {
		log.Fatalf("Error creating directory: %s", err)
	}
	os.MkdirAll(blogHtml+"css/", 0755)
	os.MkdirAll(blogHtml+"assets/", 0755)
}
