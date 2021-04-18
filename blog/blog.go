package blog

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/html"
)

type Entry struct {
	Date       time.Time
	Id         string
	Title      string
	Body       template.HTML
	CoverImage string
}

type Entries []Entry

var (
	All Entries

	entriesIndex = make(map[string]Entry)

	blogIndexTemplate,
	blogShowTemplate *template.Template
)

func init() {
	var indexParseErr, showParseErr error

	blogIndexTemplate, indexParseErr = template.ParseFiles("templates/blog/index.html")
	if indexParseErr != nil {
		log.Fatal(indexParseErr)
	}

	blogShowTemplate, showParseErr = template.ParseFiles("templates/blog/show.html")
	if showParseErr != nil {
		log.Fatal(showParseErr)
	}

	entryFiles, globErr := filepath.Glob("templates/blog/entries/*.html")
	if globErr == nil {
		for _, path := range entryFiles {
			base := filepath.Base(path)
			if len(base) > 8 {
				dateSeg := base[0:8]
				id := strings.Split(base, ".")[0]

				body, readErr := ioutil.ReadFile(path)
				if readErr == nil {

					date, dateParseErr := time.Parse("01-02-06", dateSeg)
					if dateParseErr == nil {

						bodyParts := strings.Split(string(body), "\n")

						coverImage := ""
						body := strings.Join(bodyParts[1:], "\n")

						r := bytes.NewBuffer([]byte(body))
						z := html.NewTokenizer(r)
					searchLoop:
						for {
							tt := z.Next()
							if tt == html.ErrorToken {
								break
							}
							if tt == html.SelfClosingTagToken {
								tagName, _ := z.TagName()
								if string(tagName) == "img" {
								attrLoop:
									for {
										key, val, more := z.TagAttr()
										if string(key) == "src" {
											coverImage = string(val)
											break searchLoop
										}

										if !more {
											break attrLoop
										}
									}
								}
							}
						}

						entry := Entry{
							Id:         id,
							Date:       date,
							Title:      bodyParts[0],
							CoverImage: coverImage,
							Body:       template.HTML(body),
						}
						All = append(All, entry)

						entriesIndex[id] = entry
					}
				}
			}
		}
	} else {
		panic(globErr)
	}

	sort.Sort(All)
}

func (entry Entry) DateFormatted() string {
	return entry.Date.Format("01/02/06")
}

func (e Entries) Len() int {
	return len(e)
}
func (e Entries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e Entries) Less(i, j int) bool {
	return e[i].Date.After(e[j].Date)
}

func IndexHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var (
		toRender Entries

		q = r.URL.Query()

		skipNum int
	)

	skipStr := q.Get("skip")
	if skipStr != "" {
		skip, skipErr := strconv.ParseInt(skipStr, 10, 8)
		if skipErr == nil {
			skipNum = int(skip)
		}
	}

	for i := skipNum; i < len(All) && len(toRender) < 3; i++ {
		toRender = append(toRender, All[i])
	}

	blogIndexTemplate.Execute(w, struct {
		Entries         Entries
		EntriesToRender Entries
		NextSkip        int
	}{
		Entries:         All,
		EntriesToRender: toRender,
		NextSkip:        skipNum + 3,
	})
}

func EntryHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	entry, ok := entriesIndex[id]

	if ok {
		blogShowTemplate.Execute(w, struct {
			Entries []Entry
			Entry   Entry
		}{
			Entries: All,
			Entry:   entry,
		})
	} else {
		w.WriteHeader(404)
		return
	}
}
