package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/88250/gulu"
	"github.com/google/uuid"
	"github.com/parnurzeal/gorequest"
)

var logger = gulu.Log.NewLogger(os.Stdout)

const (
	githubUserName = "stanley760"
)

func main() {
	result := map[string]interface{}{}
	uuid, _ := uuid.NewRandom()
	reg := regexp.MustCompile(`[^0-9]+`)
	str := reg.ReplaceAllString(uuid.String(), "")
	response, data, errors := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Post("https://api.juejin.cn/content_api/v1/article/query_list?aid=2608&spider=0&uuid="+str).Timeout(7*time.Second).
		Type("json").
		Send(`{"user_id":"3140618196628622","sort_type":2,"cursor":"0"}`).
		Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36 Edg/115.0.1901.203").
		EndStruct(&result)
	if nil != errors || http.StatusOK != response.StatusCode {
		logger.Fatalf("fetch events failed: %+v, %s", errors, data)
	}

	if 0 != result["err_no"].(float64) || result["data"].([]interface{}) == nil {
		logger.Fatalf("fetch events failed: %s", data)
	}
	buf := &bytes.Buffer{}
	buf.WriteString("\n\n")
	updated := time.Now().Format("2006-01-02 15:04:05")
	buf.WriteString("我的近期动态（[一键三连](https://github.com/" + githubUserName + "/" + githubUserName + ") 将自动刷新，最近更新时间：`" + updated + "`）：\n\n")
	for _, event := range result["data"].([]interface{}) {
		// article
		evt := event.(map[string]interface{})
		// article_info
		info := evt["article_info"].(map[string]interface{})
		id := info["article_id"].(string)
		title := info["title"].(string)
		url := "https://juejin.cn/post/" + id
		content := info["brief_content"].(string)
		tags := evt["tags"].([]interface{})
		var tagstr string
		for i := 0; i < len(tags); i++ {
			tag := tags[i].(map[string]interface{})
			name := tag["tag_name"].(string)
			if len(tags)-1 != i {
				tagstr = name + " | "
			} else {
				tagstr += name
			}
		}
		buf.WriteString("* [" + title + "](" + url + ")：（" + tagstr + "）" + content + "\n")
	}
	buf.WriteString("\n")
	fmt.Println(buf.String())
	readme, err := os.ReadFile("README.md")
	if nil != err {
		logger.Fatalf("read README.md failed: %s", data)
	}
	startFlag := []byte("<!--events start -->")
	beforeStart := readme[:bytes.Index(readme, startFlag)+len(startFlag)]
	newBeforeStart := make([]byte, len(beforeStart))
	copy(newBeforeStart, beforeStart)
	endFlag := []byte("<!--events end -->")
	afterEnd := readme[bytes.Index(readme, endFlag):]
	newAfterEnd := make([]byte, len(afterEnd))
	copy(newAfterEnd, afterEnd)
	newReadme := append(newBeforeStart, buf.Bytes()...)
	newReadme = append(newReadme, newAfterEnd...)

	if err := os.WriteFile("README.md", newReadme, 0644); nil != err {
		logger.Fatalf("write README.md failed: %s", data)
	}
}
