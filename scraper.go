package main

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var tempDate = time.Now().Format("2006-01-02")

// Alert type
type Alert struct {
	Title, Content, URL, Priority string
}

//SendAlert to send notification
func (a *Alert) SendAlert(source, receiver string) {
	data := url.Values{
		"source":   {source},
		"receiver": {receiver},
		"title":    {a.Title},
		"content":  {a.Content},
		"url":      {a.URL},
		"priority": {a.Priority},
	}
	resp, err := http.PostForm("https://api.alertover.com/v1/alert", data)
	if err != nil {
		println("alertover send massage error: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		println("alertover send message success.")
	}
}

func main() {
	//loop
	for {
		if time.Now().Day() == 10 {
			if ok, err := collectDocs(); ok {
				println("Collect the .md files: OK!")
			} else {
				println("collectDocs() is failed. ", err)
			}
		}
		//set monitor targets
		targets := []string{"go", "python", "javascript", "swift", "objective-c", "ruby"}

		var content, readme string
		jobs := make(chan string, 10)
		backs := make(chan string, 10)

		for w := 1; w <= 6; w++ {
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

		readme = "# Scraper\n\nWe scrape the github trending page of these languages: "
		for _, v := range targets {
			readme = readme + v + ", "
		}
		readme = readme + "and push a markdown result everyday.\n\n"
		readme = readme + "Last Updated: " + time.Now().Format("2006-01-02 15:04:05")
		writeMarkDown("README", readme)
		println("README.md is updated.")

		gitPull()
		gitAddAll()
		gitCommit()
		gitPush()

		source := "s-4e12c729-0a0c-4491-bd1a-107de60e"
		receiver := "u-fca8a4e9-1ba7-4e94-8fa5-fc2a934c"
		alert := Alert{
			Title:    "Ok",
			Content:  tempDate + ".md is completed.",
			URL:      "https://github.com/henson/Scraper",
			Priority: "1", //优先级：0 普通，1 紧急
		}
		alert.SendAlert(source, receiver)

		time.Sleep(time.Duration(24) * time.Hour)
	}
}

//collectDocs
func collectDocs() (ok bool, err error) {
	today := time.Now()
	lastMonth := today.AddDate(0, -1, 0)
	docName := lastMonth.Format("2006/01")
	regType := lastMonth.Format("2006-01")
	docPath, err := os.Getwd()
	if err != nil {
		return false, err
	}
	mdFiles, err := listDir(docPath, ".md")
	if err != nil {
		return false, err
	}
	var mdNewFiles []string
	for _, v := range mdFiles {
		if ok, _ := regexp.MatchString(regType, v); ok {
			mdNewFiles = append(mdNewFiles, v)
		}
	}
	err = os.MkdirAll(docName, 0666)
	if err != nil {
		return false, err
	}
	for _, v := range mdNewFiles {
		err = os.Rename(v, docName+string(os.PathSeparator)+v)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

//listDir
func listDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	suffix = strings.ToUpper(suffix)
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, fi.Name())
		}
	}
	return files, nil
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
