package main

import (
	"flag"
	"fmt"

	"github.com/zag13/dl/internal/cases"
)

var (
	website string
	action  string
)

func init() {
	flag.StringVar(&website, "website", "go-blog", "website to download")
	flag.StringVar(&action, "action", "111", "Three-digit bit value.\n"+
		"If the first digit is 1, download html;\n"+
		"If the second digit is 1, convert html to pdf;\n"+
		"If the third digit is 1, merged into a pdf with bookmarks.")
}

func main() {
	flag.Parse()
	switch website {
	case "go-blog":
		cases.GoBlog(action)
	case "ardanlabs":
		cases.Ardanlabs(action)
	default:
		fmt.Println("no website")
	}
}
