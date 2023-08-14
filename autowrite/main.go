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
	juejin = "7125845246609049122" ) 
func main() { 
	result := map[string]interface{}{}
	uuid,_ := uuid.NewRandom();
	reg := regexp.MustCompile(`[^0-9]+`)
	str := reg.ReplaceAllString(uuid.String(), "")
	response, data, errors := gorequest.New().TLSClientConfig(&tls.Config{InsecureSkipVerify: true}). 
			Post("https://api.juejin.cn/recommend_api/v1/article/recommend_cate_feed").Param("uuid", str).Param("aid", "6587").Timeout(7*time.Second). 
			Send(struct{
				limit int64
				client_type int64
				cursor string
				id_type int64
				cate_id string
				sort_type int64
				}{10,6587,"0",2,"6809637769959178254",200}).
			Set("Content-Type", "application/json;charset=utf-8").EndStruct(&result) 
	if nil != errors || http.StatusOK != response.StatusCode { 
		logger.Fatalf("fetch events failed: %+v, %s", errors, data) 
	} 
	
	if 0 != result["err_no"].(float64) { 
		logger.Fatalf("fetch events failed: %s", data) 
	}
	buf := &bytes.Buffer{} 
	buf.WriteString("\n\n") 
	
	buf.WriteString("juejin最新推荐blog: ") 
	for _, event := range result["data"].([]interface{}) { 
		// article
		evt := event.(map[string]interface{}) 
		// article_info
		info := evt["article_info"].(map[string]interface{}) 
		id := info["article_id"].(string)
		title := info["title"].(string) 
		url := "https://juejin.cn/post/"+ id + "?utm_source=gold_browser_extension"
		content := evt["content"].(string)
		tags := evt["tags"].([]interface{})
		var tagstr string
		for i := 0; i < len(tags); i++ {
			tag := tags[i].(map[string]interface{})
			name := tag["tag_name"].(string)
			if len(tags) - 1 != i {
				tagstr = name + " | "
			}
		}
		buf.WriteString("* [" + tagstr + "](" + url + ")：（" + title + "）" + content + "\n") 
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