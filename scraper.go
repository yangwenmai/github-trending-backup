package main

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var tempDate = time.Now().Format("2006-01-02")

func main() {
	//loop
	//for {
	//set monitor targets
	targets := []string{"go", "python", "javascript", "swift", "objective-c", "ruby"}

	var content, readme string
	jobs := make(chan string, 10)
	backs := make(chan string, 10)

	for w := 1; w <= 5; w++ {
		go scrape(jobs, backs)
	}

	for j := 0; j < len(targets); j++ {
		println(targets[j] + " is added to jobs.")
		jobs <- targets[j]
	}

	for a := 0; a < len(targets); a++ {
		content = content + <-backs
	}
	content = "### " + tempDate + "\n" + content
	//close the channels
	close(jobs)
	close(backs)

	//create markdown file
	writeMarkDown(tempDate, content)
	println(tempDate + ".md is completed.")

	readme = "# Scraper\n\nTracking the most popular Github repos, updated daily.\n\nWe scrape the trending page and push a markdown everyday.\n\n"
	readme = readme + "Last Updated: " + time.Now().Format("2006-01-02 15:04:05")
	writeMarkDown("README", readme)
	println("README.md is updated.")

	//gitPull()
	gitAddAll()
	gitCommit()
	gitPush()

	//	time.Sleep(time.Duration(24) * time.Hour)
	//}
}

//interface to string
func interface2string(inter interface{}) string {
	var tempStr string
	switch inter.(type) {
	case string:
		tempStr = inter.(string)
		break
	case float64:
		tempStr = strconv.FormatFloat(inter.(float64), 'f', -1, 64)
		break
	case int64:
		tempStr = strconv.FormatInt(inter.(int64), 10)
		break
	case int:
		tempStr = strconv.Itoa(inter.(int))
		break
	}
	return tempStr
}

func writeMarkDown(fileName, content string) {
	// open output file
	fo, err := os.Create(fileName + ".md")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()
	// make a write buffer
	w := bufio.NewWriter(fo)
	w.WriteString(content)
	w.Flush()
}

func scrape(jobs chan string, backs chan<- string) {
	defer func() {
		if r := recover(); r != nil {
			println("Recovered for", interface2string(r))
			jobs <- interface2string(r)
			go scrape(jobs, backs)
		}
	}()
	for j := range jobs {
		language := j
		var doc *goquery.Document
		var e error
		result := "\n#### " + language + "\n"

		if doc, e = goquery.NewDocument("https://github.com/trending?l=" + language); e != nil {
			println("Error:", e.Error())
			panic(language)
		}

		doc.Find("ol.repo-list li").Each(func(i int, s *goquery.Selection) {
			title := s.Find("h3 a").Text()
			description := s.Find("p.col-9").Text()
			url, _ := s.Find("h3 a").Attr("href")
			url = "https://github.com" + url
			var stars = "0"
			var forks = "0"
			s.Find("div.f6.text-gray.mt-2 a.muted-link.tooltipped.tooltipped-s.mr-3").Each(func(i int, contentSelection *goquery.Selection) {
				if temp, ok := contentSelection.Attr("aria-label"); ok {
					switch temp {
					case "Stargazers":
						stars = contentSelection.Text()
					case "Forks":
						forks = contentSelection.Text()
					}
				}
			})
			result = result + "* [" + strings.Replace(strings.TrimSpace(title), " ", "", -1) + " (" + strings.TrimSpace(stars) + "s/" + strings.TrimSpace(forks) + "f)](" + url + ") : " + strings.TrimSpace(description) + "\n"
		})
		println(language + " is responsed to backs.")
		backs <- result
	}
}

func gitPull() {
	app := "git"
	arg0 := "pull"
	arg1 := "origin"
	arg2 := "master"
	cmd := exec.Command(app, arg0, arg1, arg2)
	out, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return
	}

	print(string(out))
}

func gitAddAll() {
	app := "git"
	arg0 := "add"
	arg1 := "."
	cmd := exec.Command(app, arg0, arg1)
	out, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return
	}

	print(string(out))
}

func gitCommit() {
	app := "git"
	arg0 := "commit"
	arg1 := "-am"
	arg2 := tempDate
	cmd := exec.Command(app, arg0, arg1, arg2)
	out, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return
	}

	print(string(out))
}

func gitPush() {
	app := "git"
	arg0 := "push"
	arg1 := "origin"
	arg2 := "master"
	cmd := exec.Command(app, arg0, arg1, arg2)
	out, err := cmd.Output()

	if err != nil {
		println(err.Error())
		return
	}

	print(string(out))
}
