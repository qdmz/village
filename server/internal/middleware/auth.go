// Package middleware 鉴权、角色控制与操作日志。
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/qdmz/village-mgmt/internal/config"
	"github.com/qdmz/village-mgmt/internal/db"
	"github.com/qdmz/village-mgmt/internal/model"
	"github.com/qdmz/village-mgmt/internal/util"
)

var secret []byte

// InitSecret 读取或生成 JWT 密钥
func InitSecret() {
	s := config.C.JWT.Secret
	if s == "" {
		s = db.GetSetting("jwt_secret", "")
		if s == "" {
			s = util.RandHex(32)
			_ = db.SetSetting("jwt_secret", s)
		}
	}
	secret = []byte(s)
}

type Claims struct {
	UID  uint   `json:"uid"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func IssueToken(u *model.User) (string, error) {
	h := config.C.JWT.ExpireHours
	if h <= 0 {
		h = 72
	}
	cl := Claims{UID: u.ID, Role: u.Role, RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(h) * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   u.Username,
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, cl).SignedString(secret)
}

// Auth 校验 Bearer Token 并加载当前用户
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if tok == "" {
			tok = c.Query("token") // 用于导出下载
		}
		if tok == "" {
			util.Fail(c, http.StatusUnauthorized, "请先登录")
			return
		}
		cl := &Claims{}
		t, err := jwt.ParseWithClaims(tok, cl, func(*jwt.Token) (any, error) { return secret, nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !t.Valid {
			util.Fail(c, http.StatusUnauthorized, "登录已过期，请重新登录")
			return
		}
		var u model.User
		if err := db.DB.First(&u, cl.UID).Error; err != nil || u.Status != 1 {
			util.Fail(c, http.StatusUnauthorized, "账号不存在或已停用")
			return
		}
		c.Set("user", &u)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) *model.User {
	v, _ := c.Get("user")
	u, _ := v.(*model.User)
	return u
}

// Roles 仅允许指定角色访问
func Roles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := CurrentUser(c)
		if u == nil {
			util.Fail(c, http.StatusUnauthorized, "请先登录")
			return
		}
		for _, r := range roles {
			if u.Role == r {
				c.Next()
				return
			}
		}
		util.Forbidden(c)
	}
}

// MyGridIDs 网格员负责的网格ID；非网格员返回 nil（表示不限制）
func MyGridIDs(c *gin.Context) []uint {
	u := CurrentUser(c)
	if u == nil || u.Role != model.RoleGrid {
		return nil
	}
	if v, ok := c.Get("grid_ids"); ok {
		return v.([]uint)
	}
	ids := []uint{}
	db.DB.Model(&model.Grid{}).Where("worker_id = ?", u.ID).Pluck("id", &ids)
	if len(ids) == 0 {
		ids = []uint{0} // 无网格时不返回任何数据
	}
	c.Set("grid_ids", ids)
	return ids
}

func IsGrid(c *gin.Context) bool {
	u := CurrentUser(c)
	return u != nil && u.Role == model.RoleGrid
}

func InGrids(ids []uint, id uint) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
