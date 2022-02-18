package cases

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/zag13/dl/internal/core"
)

var (
	loginURL  = "https://account.geekbang.org/signin"
	courseURL = "https://time.geekbang.org/dashboard/course"
	username  = "username"
	password  = "password"
	nodes     []*cdp.Node
)

func Geekbang(action string) {
	for i := 0; i < 3; i++ {
		if action[i]&1 == 1 {
			switch i {
			case 0:
				fmt.Println("Geekbang: download html")
				dlGeekbang()
			case 1:
				fmt.Println("Geekbang: convert html to pdf")
				core.Html2pdf(ardBlogHtml, ardBlogPdf)
			case 2:
				fmt.Println("Geekbang: merged into a pdf with bookmarks")
				core.MergePdf(ardBlogPdf, ardBlogPdfName)
			}
		}
	}
}

func dlGeekbang() {
	loginGeekbang()
}

func loginGeekbang() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,
		// 使用账号密码登陆
		chromedp.Tasks{
			chromedp.Navigate(loginURL),
			chromedp.Click(`div.login-form > div:nth-of-type(3) > a`),
			chromedp.Sleep(time.Second),
			chromedp.SendKeys(`input[name='cellphone']`, username),
			chromedp.SendKeys(`input[name='password']`, password),
			chromedp.Click(`#agree`, chromedp.NodeVisible),
			chromedp.Click(`div.Button_button_3onsJ`),
		},
		// 使用验证码登陆
		chromedp.Tasks{
			chromedp.Navigate(loginURL),
			chromedp.Click(`#agree`, chromedp.NodeVisible),
			chromedp.Click(`'li[alias="wechat"]'`),
		},
		chromedp.Tasks{
			chromedp.Sleep(3 * time.Second),
			chromedp.Navigate(courseURL),
			chromedp.Nodes(`#app > div > div:nth-of-type(2) > div:nth-of-type(2) > div > div:nth-of-type(4)`, &nodes),
			chromedp.ActionFunc(func(ctx context.Context) error {
				fmt.Println(nodes)
				return nil
			}),
			chromedp.ActionFunc(func(c context.Context) error {
				// depth -1 for the entire subtree
				// do your best to limit the size of the subtree
				return dom.RequestChildNodes(nodes[0].NodeID).WithDepth(-1).Do(c)
			}),
			chromedp.Sleep(time.Second),
			chromedp.ActionFunc(func(c context.Context) error {
				fmt.Println(nodes)
				return nil
			}),
		},
	)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(111)
}
