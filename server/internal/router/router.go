// Package router 注册全部路由。
package router

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/qdmz/village-mgmt/internal/config"
	h "github.com/qdmz/village-mgmt/internal/handler"
	mw "github.com/qdmz/village-mgmt/internal/middleware"
	"github.com/qdmz/village-mgmt/internal/model"
)

func New(web fs.FS) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.MaxMultipartMemory = 16 << 20
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	api := r.Group("/api")
	api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	api.GET("/public/settings", h.GetSettings)
	api.POST("/auth/login", h.Login)
	api.GET("/public/portal", h.PublicPortal)
	api.GET("/public/notices/:id", h.PublicNotice)
	api.GET("/public/items/:id", h.PublicItem)

	a := api.Group("", mw.Auth(), mw.OpLog())
	ac := mw.Roles(model.RoleAdmin, model.RoleCadre)
	ad := mw.Roles(model.RoleAdmin)

	a.GET("/me", h.Me)
	a.PUT("/me/profile", h.UpdateProfile)
	a.PUT("/me/password", h.ChangePassword)
	a.GET("/dashboard", h.Dashboard)
	a.GET("/org/tree", h.OrgTree)
	a.POST("/upload", h.Upload)
	a.GET("/users/options", h.UserOptions)
	a.GET("/settings", h.GetSettings)
	a.PUT("/settings", ad, h.SaveSettings)

	h.VillageRes.Register(a, "/villages")
	h.GroupRes.Register(a, "/groups")
	h.GridRes.Register(a, "/grids")
	a.POST("/households/import", h.ImportHouseholds)
	h.HouseholdRes.Register(a, "/households")
	a.POST("/villagers/import", h.ImportVillagers)
	h.VillagerRes.Register(a, "/villagers")
	h.PatrolRes.Register(a, "/patrols")
	a.POST("/events/:id/flow", h.EventFlowAction)
	h.EventRes.Register(a, "/events")
	a.POST("/tasks/:id/progress", h.TaskProgress)
	h.TaskRes.Register(a, "/tasks")
	h.WorkLogRes.Register(a, "/worklogs")
	h.NoticeRes.Register(a, "/notices")
	h.ProjectRes.Register(a, "/projects")
	a.GET("/funds/summary", ac, h.FundSummary)
	h.FundRes.Register(a, "/funds")
	h.AssetRes.Register(a, "/assets")
	h.MeetingRes.Register(a, "/meetings")

	a.GET("/party/orgs/tree", h.PartyOrgTree)
	h.PartyOrgRes.Register(a, "/party/orgs")
	h.PartyMemberRes.Register(a, "/party/members")
	a.GET("/welfare/stats", h.WelfareStats)
	h.WelfareRes.Register(a, "/welfare")
	h.DisputeRes.Register(a, "/disputes")
	h.HonorRes.Register(a, "/honors")
	a.GET("/points/rank", h.PointRank)
	h.PointRes.Register(a, "/points")
	h.RichLeaderRes.Register(a, "/leaders")
	a.GET("/funds/stats", ac, h.FundStats)
	h.PortalRes.Register(a, "/portal")
	a.GET("/screen", h.Screen)
	a.GET("/system/backup", ad, h.Backup)
	a.POST("/system/restore", ad, h.Restore)

	users := a.Group("/users", ad)
	users.GET("", h.UserRes.List)
	users.GET("/export", h.UserRes.ExportCSV)
	users.GET("/:id", h.UserRes.Get)
	users.POST("", h.CreateUser)
	users.PUT("/:id", h.UpdateUser)
	users.DELETE("/:id", h.DeleteUser)
	users.POST("/:id/reset-password", h.ResetPassword)

	a.GET("/oplogs", h.OpLogRes.List)
	a.GET("/oplogs/export", h.OpLogRes.ExportCSV)

	r.Static("/uploads", config.C.Server.UploadDir)

	// 前端静态资源（hash 路由，未知路径回落 index.html）
	if d := config.C.Server.StaticDir; d != "" {
		web = os.DirFS(d)
	}
	fileServer := http.FileServer(http.FS(web))
	r.NoRoute(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/api/") {
			c.JSON(404, gin.H{"code": 404, "msg": "接口不存在"})
			return
		}
		clean := strings.TrimPrefix(path.Clean(p), "/")
		if clean != "" {
			if st, err := fs.Stat(web, clean); err == nil && !st.IsDir() {
				if strings.HasPrefix(clean, "assets/") {
					c.Header("Cache-Control", "public, max-age=31536000, immutable")
				}
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}
		b, err := fs.ReadFile(web, "index.html")
		if err != nil {
			c.String(200, "前端尚未构建，请先执行 web 目录下的 npm run build")
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.Data(200, "text/html; charset=utf-8", b)
	})
	return r
}
