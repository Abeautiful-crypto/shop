package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const API_KEY = ""
const SECRET_KEY = ""

func BaiduPass(content string) int {

	url := "https://aip.baidubce.com/rest/2.0/solution/v1/text_censor/v2/user_defined?access_token=" + GetAccessToken()
	payload := strings.NewReader("text=" + content + "&strategyId=1")
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Println(err)
		return 0
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	fmt.Println(string(body))

	var data T
	err = json.Unmarshal(body, &data)
	return data.ConclusionType //1.合规，2.不合规，3.疑似，4.审核失败
}

type T struct {
	Conclusion     string      `json:"conclusion"`
	LogId          string      `json:"log_id"`
	PhoneRisk      interface{} `json:"phoneRisk"`
	IsHitMd5       bool        `json:"isHitMd5"`
	ConclusionType int         `json:"conclusionType"`
}

/**
 * 使用 AK，SK 生成鉴权签名（Access Token）
 * @return string 鉴权签名信息（Access Token）
 */
func GetAccessToken() string {
	url := "https://aip.baidubce.com/oauth/2.0/token"
	postData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", API_KEY, SECRET_KEY)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(postData))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	accessTokenObj := map[string]any{}
	_ = json.Unmarshal([]byte(body), &accessTokenObj)
	return accessTokenObj["access_token"].(string)
}
