// Package util 通用工具：统一响应、分页、脱敏等。
package util

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": data})
}

func Fail(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"code": status, "msg": msg})
}

func BadRequest(c *gin.Context, msg string) { Fail(c, http.StatusBadRequest, msg) }
func Forbidden(c *gin.Context)             { Fail(c, http.StatusForbidden, "无权限执行该操作") }
func NotFound(c *gin.Context)              { Fail(c, http.StatusNotFound, "记录不存在") }
func ServerError(c *gin.Context, err error) {
	Fail(c, http.StatusInternalServerError, "服务器错误："+err.Error())
}

// Page 解析分页参数
func Page(c *gin.Context) (page, size int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 500 {
		size = 20
	}
	return
}

func ParamID(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id)
}

// MaskIDCard 身份证脱敏：保留前6后4
func MaskIDCard(s string) string {
	if len(s) < 10 {
		return s
	}
	return s[:6] + strings.Repeat("*", len(s)-10) + s[len(s)-4:]
}

func MaskPhone(s string) string {
	if len(s) != 11 {
		return s
	}
	return s[:3] + "****" + s[7:]
}

func RandHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// RandPassword 生成强随机密码（大小写+数字+符号）
func RandPassword(n int) string {
	const sets = "ABCDEFGHJKLMNPQRSTUVWXYZ|abcdefghijkmnpqrstuvwxyz|23456789|@#%+=!"
	groups := strings.Split(sets, "|")
	all := strings.Join(groups, "")
	out := make([]byte, 0, n)
	for _, g := range groups {
		out = append(out, g[ri(len(g))])
	}
	for len(out) < n {
		out = append(out, all[ri(len(all))])
	}
	for i := len(out) - 1; i > 0; i-- {
		j := ri(i + 1)
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

func ri(n int) int {
	v, _ := rand.Int(rand.Reader, big.NewInt(int64(n)))
	return int(v.Int64())
}
