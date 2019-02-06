package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var mainUrl = "https://www.mzitu.com/hot/"

type subject struct {
	id   int
	name string
	url  string
}

var subjects = make(map[int]*subject)

func main() {
	main := colly.NewCollector(colly.Async(true))
	detail := colly.NewCollector(colly.Async(true))

	main.OnHTML("#pins li", func(e *colly.HTMLElement) {
		detailURL, _ := e.DOM.Children().Eq(0).Attr("href")
		detailName := e.DOM.Children().Eq(1).Text()

		idslc := strings.Split(detailURL, "/")
		id, _ := strconv.Atoi(idslc[len(idslc)-1])

		sub := &subject{
			id:   id,
			name: detailName,
			url:  detailURL,
		}

		subjects[id] = sub

		err := os.MkdirAll("pic/"+detailName, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		ctx := colly.NewContext()
		ctx.Put("id", id)

		detail.Request("GET", detailURL, nil, ctx, nil)
	})

	main.OnHTML(".next.page-numbers", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	main.OnError(func(r *colly.Response, e error) {
		fmt.Println(e)
	})

	detail.OnHTML(".content p img", func(e *colly.HTMLElement) {
		picURL := e.Attr("src")
		request, err := http.NewRequest("GET", picURL, nil)
		if err != nil {
			fmt.Println(e)
			return
		}

		request.Header.Add("Referer", e.Request.URL.String())
		client := http.Client{}
		resp, err := client.Do(request)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		s1 := strings.Split(picURL, "/")
		s := s1[len(s1)-1]

		bytes, _ := ioutil.ReadAll(resp.Body)

		file, _ := os.Create("pic/" + subjects[e.Request.Ctx.GetAny("id").(int)].name + "/" + s)
		file.Write(bytes)
		defer file.Close()
	})

	detail.OnHTML(".pagenavi a:last-child", func(e *colly.HTMLElement) {
		if e.DOM.Children().Eq(0).Text() == "下一页»" {
			fmt.Printf("正在抓取%s\n", e.Request.URL.String())
			e.Request.Visit(e.Attr("href"))
		}
	})

	detail.OnError(func(r *colly.Response, e error) {
		fmt.Println(e)
	})

	main.Visit(mainUrl)
	main.Wait()
	detail.Wait()
}
