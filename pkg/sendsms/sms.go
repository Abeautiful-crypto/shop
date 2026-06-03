package sendsms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const strUrl = "https://api.ihuyi.com/sms/Submit.json"
const APPID = ""  //APIID(用户中心【文本短信】-【验证码/通知短信】-【产品总览】查看)
const APIKEY = "" //1、APIKEY(用户中心【文本短信】-【验证码/通知短信】

func SendSms(code string, mobile string) (T, error) {
	v := url.Values{}
	v.Set("account", APPID)                            //APIID(用户中心【文本短信】-【验证码/通知短信】-【产品总览】查看)
	v.Set("password", APIKEY)                          //1、APIKEY(用户中心【文本短信】-【验证码/通知短信】-【产品总览】查看)2、动态密码(生成动态密码方式请看该文档末尾的说明)
	v.Set("mobile", mobile)                            //根据发送方式不同：1、完整内容方式提交完整的短信内容，如：您的验证码是：1234。请不要把验证码泄露给其他人。2、模板变量方式模板中的变量内容，多个变量以英文竖线（|）隔开①单变量示例模板内容：您的验证码是：【变量】。请不要把验证码泄露给其他人。参数写法：content=1234最终短信为：您的验证码是：1234。请不要把验证码泄露给其他人。②多变量示例模板内容：订单号：【变量1】，联系人：【变量2】，手机号：【变量3】，金额：【变量4】。参数写法：content=20180515006|张三|136xxxxxxxx|100元最终短信为：订单号：20180515006，联系人：张三，手机号：136xxxxxxxx，金额：100元。支持500字以内的长短信，长短信按多条计费
	v.Set("content", "您的验证码是："+code+"。请不要把验证码泄露给其他人。") //短信内容，注：模板ID为空时必填
	//v.Set("templateid", "1")                           //短信模板ID（使用模板变量方式发送时必填）调试阶段可使用系统默认模板ID：1（模板内容为：您的验证码是：【变量】。请不要把验证码泄露给其他人。）
	//v.Set("time", "1623643787")                        //Unix时间戳（10位整型数字，当使用动态密码方式时为必填）

	body := strings.NewReader(v.Encode()) //把form数据编码
	client := &http.Client{}
	req, _ := http.NewRequest("POST", strUrl, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req) //发送
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close() //一定要关闭resp.Body
	res, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(res)) //短信通道返回的json数据 {"code":1,"msg":"ok","smsid":11111}

	var t T
	json.Unmarshal(res, &t)
	return t, err
}

type T struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Smsid string `json:"smsid"`
}
