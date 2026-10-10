package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qdmz/village-mgmt/internal/db"
	"github.com/qdmz/village-mgmt/internal/model"
)

var moduleNames = map[string]string{
	"auth": "登录认证", "me": "个人中心", "villages": "村情信息", "groups": "村民小组", "grids": "网格管理",
	"households": "户籍台账", "villagers": "人口台账", "patrols": "网格巡查", "events": "事件上报",
	"tasks": "任务督办", "worklogs": "工作日志", "notices": "通知公告", "projects": "产业项目",
	"funds": "资金台账", "assets": "集体资产", "meetings": "会议活动", "users": "用户管理",
	"settings": "系统设置", "upload": "文件上传",
}

// OpLog 记录所有写操作
func OpLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method == "GET" {
			return
		}
		u := CurrentUser(c)
		if u == nil {
			return
		}
		path := c.Request.URL.Path
		parts := strings.Split(strings.TrimPrefix(path, "/api/"), "/")
		mod := moduleNames[parts[0]]
		if mod == "" {
			mod = parts[0]
		}
		act := map[string]string{"POST": "新增", "PUT": "修改", "DELETE": "删除", "PATCH": "修改"}[c.Request.Method]
		if len(parts) > 1 {
			last := parts[len(parts)-1]
			switch last {
			case "import":
				act = "导入"
			case "flow":
				act = "流转"
			case "progress":
				act = "进度更新"
			case "password":
				act = "修改密码"
			case "reset-password":
				act = "重置密码"
			}
		}
		db.DB.Create(&model.OpLog{UserID: u.ID, Username: u.Username, RealName: u.RealName, Method: c.Request.Method,
			Path: path, Module: mod, Action: act, IP: c.ClientIP(), Status: c.Writer.Status(), CreatedAt: time.Now()})
	}
}
