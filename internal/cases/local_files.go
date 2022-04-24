package cases

import (
	"fmt"

	"github.com/zag13/dl/internal/core"
)

var (
	localFilesHtml    = "./storage/local_files/html/"
	localFilesPdf     = "./storage/local_files/pdf/"
	localFilesPdfName = "./storage/local_files/local_files"
)

func LocalFiles(action string) {
	for i := 0; i < 3; i++ {
		if action[i]&1 == 1 {
			switch i {
			case 0:
				fmt.Println("NOPE!!!")
			case 1:
				fmt.Println("LocalFiles: convert html to pdf")
				core.Html2pdf(localFilesHtml, localFilesPdf)
			case 2:
				fmt.Println("LocalFiles: merged into a pdf with bookmarks")
				core.MergePdf(localFilesPdf, localFilesPdfName)
			}
		}
	}
}
