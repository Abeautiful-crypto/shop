package middleware

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ==================== 配置 ===========================================================================================
const (
	SignKey        = "7e2b8f1a9d5c0374620b8e3d1f9a6c5b47802d9s1k3j5h7g6f8d2s4a" // 签名密钥
	ExpireSec      = 300                                                        // 请求过期时间 5分钟
	SignField      = "sign"                                                     // 签名字段
	TimestampField = "timestamp"                                                // 时间戳字段
	NonceField     = "nonce"                                                    // 随机数字段
)

// =========================== 签名工具 =================================================================================

// GenerateSign 生成 MD5 签名
// 规则：参数按字典序排序 + 拼接 + 密钥
func GenerateSign(params map[string]string, key string) string {
	// 1. 取出所有 key
	var keys []string
	for k := range params {
		// 跳过 sign 自身
		if k == SignField {
			continue
		}
		keys = append(keys, k)
	}
	fmt.Println("获取所有的key:", keys)

	// 2. 字典序排序
	sort.Strings(keys)
	fmt.Println("参数按照字典升序排列:", keys)

	// 3. 拼接字符串 k1=v1&k2=v2&...
	var strList []string
	for _, k := range keys {
		strList = append(strList, fmt.Sprintf("%s=%s", k, params[k]))
	}
	signStr := strings.Join(strList, "&")
	fmt.Println("拼接字符串:", signStr)

	// 4. 最后拼接密钥
	signStr = signStr + key
	fmt.Println("拼接字符串和密钥后:", signStr)

	// 5. MD5 加密并转小写
	hash := md5.Sum([]byte(signStr))
	sign := strings.ToLower(hex.EncodeToString(hash[:]))
	fmt.Println("加密计算后的sign:", signStr)
	return sign
}

// SignMiddleware ==================== 签名中间件 ========================================================================
func SignMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// ===================== 核心改造：统一获取所有参数 =====================
		params := make(map[string]string)

		// 1. 获取 URL 参数
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}

		// 读取 body（必须读，且必须归还）
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = c.GetRawData()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		contentType := c.GetHeader("Content-Type")

		// --- 2. 解析 form-data（关键！你之前缺失这个）---
		if strings.Contains(contentType, "multipart/form-data") {
			// 解析 form-data
			if err := c.Request.ParseMultipartForm(32 << 20); err == nil {
				for k, v := range c.Request.MultipartForm.Value {
					if len(v) > 0 {
						params[k] = v[0]
					}
				}
			}
		}

		// --- 3. 解析 form-urlencoded ---
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			_ = c.Request.ParseForm()
			for k, v := range c.Request.PostForm {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}
		}

		// --- 4. 解析 JSON ---
		if strings.Contains(contentType, "application/json") && len(bodyBytes) > 0 {
			var jsonMap map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &jsonMap); err == nil {
				for k, v := range jsonMap {
					params[k] = fmt.Sprint(v)
				}
			}
		}

		// 调试输出（看看现在能不能拿到！）
		fmt.Println("=================解析到的参数===================================================================")
		fmt.Println(params)
		fmt.Println("===============================================================================================")

		// 2. 检查必传字段
		if params[TimestampField] == "" || params[NonceField] == "" || params[SignField] == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "缺少 timestamp / nonce / sign 参数"})
			c.Abort()
			return
		}

		// 3. 校验时间戳是否过期
		ts, err := strconv.ParseInt(params[TimestampField], 10, 64)
		if err != nil || time.Now().Unix()-ts > ExpireSec {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求已过期"})
			c.Abort()
			return
		}

		// 4. 生成服务端签名
		serverSign := GenerateSign(params, SignKey)

		// 5. 对比签名
		if serverSign != params[SignField] {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "签名验证失败"})
			c.Abort()
			return
		}

		// 验证通过，继续执行
		c.Next()
	}
}
