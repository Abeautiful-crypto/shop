package utils

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func Md5Str(str string) string {

	data := []byte(str)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has) //将[]byte转成16进制

	fmt.Println(md5str1)

	return md5str1
}

// FilterSensitiveWords 敏感词自动替换
func FilterSensitiveWords(content string) string {
	// 这里写你的敏感词列表，示例：
	// SensitiveWords 极简敏感词库，可直接用于自动过滤整改
	var sensitiveWords = []string{
		// 政治/违规类
		"暴力",
		"赌博",
		"违法",
		"违规",
		"反动",
		"邪教",
		"台独",
		"港独",
		"分裂",
		"叛国",
		"颠覆",
		"恐怖主义",
		"极端主义",

		// 低俗/擦边类
		"色情",
		"黄色",
		"裸聊",
		"嫖娼",
		"卖淫",
		"约炮",
		"一夜情",
		"裸贷",
		"涉黄",
		"低俗",
		"擦边",
		"诱惑",
		"挑逗",
		"成人内容",
		"涩图",
		"福利姬",

		// 广告/导流/诈骗类
		"加微信",
		"加v",
		"加vx",
		"私我",
		"私聊",
		"私信我",
		"加我",
		"引流",
		"导流",
		"推广",
		"刷单",
		"刷信誉",
		"兼职刷单",
		"返现",
		"返利",
		"高佣金",
		"日赚",
		"躺赚",
		"稳赚",
		"包赚",
		"无本万利",
		"轻松赚钱",
		"彩票",
		"博彩",
		"赌博",
		"网赌",
		"时时彩",
		"六合彩",
		"私彩",

		// 违规服务/产品类
		"外挂",
		"破解",
		"破解版",
		"盗版",
		"翻墙",
		"VPN",
		"科学上网",
		"翻墙软件",
		"翻墙工具",
		"代练",
		"代打",
		"代刷",
		"刷量",
		"刷赞",
		"刷粉",
		"刷评论",
		"刷播放",
		"刷阅读",

		// 其他违规词
		"辱骂",
		"脏话",
		"侮辱",
		"诽谤",
		"造谣",
		"传谣",
		"恶意攻击",
		"人身攻击",
		"地域黑",
		"种族歧视",
		"性别歧视",
	}
	//sensitiveWords := []string{"敏感词1", "敏感词2"}
	cleaned := content
	for _, word := range sensitiveWords {
		cleaned = strings.ReplaceAll(cleaned, word, "***")
	}
	return cleaned
}
