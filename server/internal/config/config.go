// Package config 负责读取 config.yaml 配置。
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Addr      string `yaml:"addr"`       // 监听地址，如 127.0.0.1:8095
		StaticDir string `yaml:"static_dir"` // 可选：外部前端目录，留空则使用内嵌前端
		UploadDir string `yaml:"upload_dir"` // 上传文件目录
		MaxUpload int64  `yaml:"max_upload_mb"`
	} `yaml:"server"`
	Database struct {
		Driver string `yaml:"driver"` // sqlite | mysql
		DSN    string `yaml:"dsn"`    // sqlite 文件路径或 mysql DSN
	} `yaml:"database"`
	JWT struct {
		Secret      string `yaml:"secret"` // 留空时自动生成并保存到数据库
		ExpireHours int    `yaml:"expire_hours"`
	} `yaml:"jwt"`
	Seed struct {
		Demo          bool   `yaml:"demo"`           // 首次启动是否写入演示数据
		AdminPassword string `yaml:"admin_password"` // 初始管理员密码，留空则随机生成并打印到日志
	} `yaml:"seed"`
}

var C Config

func Load(path string) error {
	C.Server.Addr = "127.0.0.1:8095"
	C.Server.UploadDir = "./data/uploads"
	C.Server.MaxUpload = 5
	C.Database.Driver = "sqlite"
	C.Database.DSN = "./data/village.db"
	C.JWT.ExpireHours = 72
	C.Seed.Demo = true
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 零配置运行
		}
		return err
	}
	return yaml.Unmarshal(b, &C)
}
