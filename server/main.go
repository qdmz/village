// 乡村振兴村居日常工作管理系统 - 后端入口
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/qdmz/village-mgmt/internal/config"
	"github.com/qdmz/village-mgmt/internal/db"
	mw "github.com/qdmz/village-mgmt/internal/middleware"
	"github.com/qdmz/village-mgmt/internal/router"
	"github.com/qdmz/village-mgmt/internal/seed"
	"github.com/qdmz/village-mgmt/webui"
)

var Version = "v0.1.0"

func main() {
	cfgPath := flag.String("c", "config.yaml", "配置文件路径")
	showVer := flag.Bool("v", false, "显示版本")
	flag.Parse()
	if *showVer {
		fmt.Println(Version)
		return
	}
	if err := config.Load(*cfgPath); err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}
	if err := os.MkdirAll(config.C.Server.UploadDir, 0o755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}
	if err := db.Init(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	if err := seed.Run(); err != nil {
		log.Fatalf("初始化数据失败: %v", err)
	}
	mw.InitSecret()
	r := router.New(webui.FS())
	log.Printf("乡村振兴村居日常工作管理系统 %s 启动，监听 %s（数据库：%s）", Version, config.C.Server.Addr, config.C.Database.Driver)
	if err := r.Run(config.C.Server.Addr); err != nil {
		log.Fatal(err)
	}
}
