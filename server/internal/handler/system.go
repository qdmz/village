package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qdmz/village-mgmt/internal/config"
	"github.com/qdmz/village-mgmt/internal/db"
	mw "github.com/qdmz/village-mgmt/internal/middleware"
	"github.com/qdmz/village-mgmt/internal/model"
	"github.com/qdmz/village-mgmt/internal/util"
	"golang.org/x/crypto/bcrypt"
)

// ---------------- 登录 ----------------

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var loginFails sync.Map // 简单的登录失败限流：key=ip|user

type failRec struct {
	n     int
	until time.Time
}

func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		util.BadRequest(c, "请输入用户名和密码")
		return
	}
	key := c.ClientIP() + "|" + req.Username
	if v, ok := loginFails.Load(key); ok {
		if r := v.(*failRec); r.n >= 5 && time.Now().Before(r.until) {
			util.Fail(c, 429, "密码错误次数过多，请 5 分钟后再试")
			return
		}
	}
	var u model.User
	if err := db.DB.Where("username = ?", req.Username).First(&u).Error; err != nil ||
		bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		v, _ := loginFails.LoadOrStore(key, &failRec{})
		r := v.(*failRec)
		r.n++
		r.until = time.Now().Add(5 * time.Minute)
		util.Fail(c, 401, "用户名或密码错误")
		return
	}
	if u.Status != 1 {
		util.Fail(c, 403, "账号已停用，请联系管理员")
		return
	}
	loginFails.Delete(key)
	tok, err := mw.IssueToken(&u)
	if err != nil {
		util.ServerError(c, err)
		return
	}
	now := time.Now()
	db.DB.Model(&u).Update("last_login", &now)
	db.DB.Create(&model.OpLog{UserID: u.ID, Username: u.Username, RealName: u.RealName, Method: "POST", Path: "/api/auth/login",
		Module: "登录认证", Action: "登录", IP: c.ClientIP(), Status: 200, CreatedAt: now})
	util.OK(c, gin.H{"token": tok, "user": u, "grid_ids": gridIDsOf(&u)})
}

func gridIDsOf(u *model.User) []uint {
	ids := []uint{}
	if u.Role == model.RoleGrid {
		db.DB.Model(&model.Grid{}).Where("worker_id = ?", u.ID).Pluck("id", &ids)
	}
	return ids
}

func Me(c *gin.Context) {
	u := mw.CurrentUser(c)
	util.OK(c, gin.H{"user": u, "grid_ids": gridIDsOf(u)})
}

type pwdReq struct {
	Old string `json:"old_password"`
	New string `json:"new_password"`
}

func checkPasswordStrength(p string) error {
	if len(p) < 8 {
		return errors.New("新密码至少 8 位")
	}
	kinds := 0
	for _, set := range []string{"abcdefghijklmnopqrstuvwxyz", "ABCDEFGHIJKLMNOPQRSTUVWXYZ", "0123456789"} {
		if strings.ContainsAny(p, set) {
			kinds++
		}
	}
	if kinds < 2 {
		return errors.New("新密码需包含字母、数字中的至少两类字符")
	}
	return nil
}

func ChangePassword(c *gin.Context) {
	u := mw.CurrentUser(c)
	var req pwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, "参数格式错误")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Old)) != nil {
		util.BadRequest(c, "原密码不正确")
		return
	}
	if err := checkPasswordStrength(req.New); err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	h, _ := bcrypt.GenerateFromPassword([]byte(req.New), bcrypt.DefaultCost)
	db.DB.Model(u).Update("password", string(h))
	util.OK(c, nil)
}

type profileReq struct {
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
}

func UpdateProfile(c *gin.Context) {
	u := mw.CurrentUser(c)
	var req profileReq
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.RealName) == "" {
		util.BadRequest(c, "请填写姓名")
		return
	}
	db.DB.Model(u).Updates(map[string]any{"real_name": req.RealName, "phone": req.Phone})
	util.OK(c, nil)
}

// ---------------- 用户管理 ----------------

type userReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
	Phone    string `json:"phone"`
	Position string `json:"position"`
	Status   *int   `json:"status"`
}

var UserRes = &Resource[model.User]{Title: "用户", Search: []string{"username", "real_name", "phone"},
	Filters: map[string]string{"role": "role", "status": "status"}, ReadRoles: adminOnly, WriteRoles: adminOnly,
	Order: "id",
	Export: []Col[model.User]{
		{"账号", func(u *model.User) string { return u.Username }},
		{"姓名", func(u *model.User) string { return u.RealName }},
		{"角色", func(u *model.User) string { return RoleName(u.Role) }},
		{"职务", func(u *model.User) string { return u.Position }},
		{"电话", func(u *model.User) string { return u.Phone }},
		{"状态", func(u *model.User) string { return map[int]string{1: "启用", 0: "停用"}[u.Status] }},
	},
}

func RoleName(r string) string {
	return map[string]string{model.RoleAdmin: "管理员", model.RoleCadre: "村干部", model.RoleGrid: "网格员"}[r]
}

func validRole(r string) bool {
	return r == model.RoleAdmin || r == model.RoleCadre || r == model.RoleGrid
}

func CreateUser(c *gin.Context) {
	var req userReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, "参数格式错误")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.RealName == "" || !validRole(req.Role) {
		util.BadRequest(c, "请填写账号、姓名并选择角色")
		return
	}
	if err := checkPasswordStrength(req.Password); err != nil {
		util.BadRequest(c, strings.Replace(err.Error(), "新密码", "密码", 1))
		return
	}
	var n int64
	db.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&n)
	if n > 0 {
		util.BadRequest(c, "账号已存在")
		return
	}
	h, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u := model.User{Username: req.Username, Password: string(h), RealName: req.RealName, Role: req.Role, Phone: req.Phone, Position: req.Position, Status: 1}
	if req.Status != nil {
		u.Status = *req.Status
	}
	if err := db.DB.Create(&u).Error; err != nil {
		util.ServerError(c, err)
		return
	}
	util.OK(c, u)
}

func UpdateUser(c *gin.Context) {
	var u model.User
	if err := db.DB.First(&u, util.ParamID(c)).Error; err != nil {
		util.NotFound(c)
		return
	}
	var req userReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, "参数格式错误")
		return
	}
	me := mw.CurrentUser(c)
	if req.Role != "" {
		if !validRole(req.Role) {
			util.BadRequest(c, "角色不正确")
			return
		}
		if u.ID == me.ID && req.Role != model.RoleAdmin {
			util.BadRequest(c, "不能修改自己的管理员角色")
			return
		}
		u.Role = req.Role
	}
	if req.RealName != "" {
		u.RealName = req.RealName
	}
	u.Phone, u.Position = req.Phone, req.Position
	if req.Status != nil {
		if u.ID == me.ID && *req.Status != 1 {
			util.BadRequest(c, "不能停用自己的账号")
			return
		}
		u.Status = *req.Status
	}
	if err := db.DB.Save(&u).Error; err != nil {
		util.ServerError(c, err)
		return
	}
	db.DB.Model(&model.Grid{}).Where("worker_id = ?", u.ID).Update("worker_name", u.RealName)
	util.OK(c, u)
}

func ResetPassword(c *gin.Context) {
	var u model.User
	if err := db.DB.First(&u, util.ParamID(c)).Error; err != nil {
		util.NotFound(c)
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Password == "" {
		req.Password = util.RandPassword(12)
	}
	if err := checkPasswordStrength(req.Password); err != nil {
		util.BadRequest(c, strings.Replace(err.Error(), "新密码", "密码", 1))
		return
	}
	h, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	db.DB.Model(&u).Update("password", string(h))
	util.OK(c, gin.H{"password": req.Password})
}

func DeleteUser(c *gin.Context) {
	id := util.ParamID(c)
	if id == mw.CurrentUser(c).ID {
		util.BadRequest(c, "不能删除当前登录账号")
		return
	}
	var u model.User
	if err := db.DB.First(&u, id).Error; err != nil {
		util.NotFound(c)
		return
	}
	db.DB.Delete(&u)
	db.DB.Model(&model.Grid{}).Where("worker_id = ?", id).Updates(map[string]any{"worker_id": 0, "worker_name": ""})
	util.OK(c, nil)
}

// UserOptions 供选择责任人/处理人使用
func UserOptions(c *gin.Context) {
	type opt struct {
		ID       uint   `json:"id"`
		RealName string `json:"real_name"`
		Role     string `json:"role"`
		Position string `json:"position"`
	}
	var list []opt
	q := db.DB.Model(&model.User{}).Where("status = 1")
	if r := c.Query("role"); r != "" {
		q = q.Where("role = ?", r)
	}
	q.Order("id").Scan(&list)
	util.OK(c, list)
}

// ---------------- 系统设置 ----------------

var settingKeys = []string{"village_name", "system_title", "logo_text", "town_name", "contact_phone", "footer_text"}

func defaultSettings() map[string]string {
	return map[string]string{"village_name": "青山村", "system_title": "乡村振兴村居日常工作管理系统", "logo_text": "青",
		"town_name": "", "contact_phone": "", "footer_text": "数字乡村 · 网格治理"}
}

func GetSettings(c *gin.Context) {
	out := defaultSettings()
	for _, k := range settingKeys {
		if v := db.GetSetting(k, ""); v != "" {
			out[k] = v
		}
	}
	util.OK(c, out)
}

func SaveSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, "参数格式错误")
		return
	}
	for _, k := range settingKeys {
		if v, ok := req[k]; ok {
			if len([]rune(v)) > 100 {
				util.BadRequest(c, "设置内容过长")
				return
			}
			_ = db.SetSetting(k, strings.TrimSpace(v))
		}
	}
	GetSettings(c)
}

// ---------------- 组织树 ----------------

func OrgTree(c *gin.Context) {
	var v model.Village
	db.DB.First(&v)
	var groups []model.Group
	db.DB.Order("sort, id").Find(&groups)
	var grids []model.Grid
	db.DB.Order("sort, id").Find(&grids)
	type cnt struct {
		GridID uint
		N      int
	}
	var hc, pc []cnt
	db.DB.Model(&model.Household{}).Select("grid_id, COUNT(*) AS n").Group("grid_id").Scan(&hc)
	db.DB.Model(&model.Villager{}).Select("grid_id, COUNT(*) AS n").Group("grid_id").Scan(&pc)
	hm, pm := map[uint]int{}, map[uint]int{}
	for _, x := range hc {
		hm[x.GridID] = x.N
	}
	for _, x := range pc {
		pm[x.GridID] = x.N
	}
	gChildren := map[uint][]gin.H{}
	for _, g := range grids {
		gChildren[g.GroupID] = append(gChildren[g.GroupID], gin.H{"key": fmt.Sprintf("grid-%d", g.ID), "id": g.ID, "type": "grid",
			"label": g.Name, "worker": g.WorkerName, "worker_id": g.WorkerID, "households": hm[g.ID], "people": pm[g.ID], "data": g})
	}
	children := []gin.H{}
	for _, gr := range groups {
		ch := gChildren[gr.ID]
		if ch == nil {
			ch = []gin.H{}
		}
		children = append(children, gin.H{"key": fmt.Sprintf("group-%d", gr.ID), "id": gr.ID, "type": "group", "label": gr.Name,
			"leader": gr.Leader, "children": ch, "data": gr})
	}
	if ch := gChildren[0]; len(ch) > 0 {
		children = append(children, gin.H{"key": "group-0", "id": 0, "type": "group", "label": "未分组网格", "children": ch})
	}
	util.OK(c, []gin.H{{"key": "village", "id": v.ID, "type": "village", "label": v.Name, "children": children, "data": v}})
}

// ---------------- 文件上传 ----------------

var allowExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}

func Upload(c *gin.Context) {
	max := config.C.Server.MaxUpload
	if max <= 0 {
		max = 5
	}
	c.Request.Body = http_MaxBytes(c, max<<20+1024)
	f, err := c.FormFile("file")
	if err != nil {
		util.BadRequest(c, "请选择文件（不超过 "+fmt.Sprint(max)+"MB）")
		return
	}
	if f.Size > max<<20 {
		util.BadRequest(c, fmt.Sprintf("文件不能超过 %dMB", max))
		return
	}
	ext := strings.ToLower(filepath.Ext(f.Filename))
	if !allowExt[ext] {
		util.BadRequest(c, "仅支持 jpg/png/gif/webp 图片")
		return
	}
	src, err := f.Open()
	if err != nil {
		util.ServerError(c, err)
		return
	}
	defer src.Close()
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	ct := httpDetect(head[:n])
	if !strings.HasPrefix(ct, "image/") {
		util.BadRequest(c, "文件内容不是有效图片")
		return
	}
	sub := time.Now().Format("200601")
	dir := filepath.Join(config.C.Server.UploadDir, sub)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		util.ServerError(c, err)
		return
	}
	name := util.RandHex(12) + ext
	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		util.ServerError(c, err)
		return
	}
	defer dst.Close()
	_, _ = dst.Write(head[:n])
	if _, err := io.Copy(dst, src); err != nil {
		util.ServerError(c, err)
		return
	}
	util.OK(c, gin.H{"url": "/uploads/" + sub + "/" + name, "name": f.Filename})
}

// ---------------- CSV 导入 ----------------

func readCSV(c *gin.Context) ([]map[string]string, error) {
	f, err := c.FormFile("file")
	if err != nil {
		return nil, errors.New("请选择 CSV 文件")
	}
	if f.Size > 10<<20 {
		return nil, errors.New("文件不能超过 10MB")
	}
	src, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	b, _ := io.ReadAll(src)
	s := strings.TrimPrefix(string(b), "\xEF\xBB\xBF")
	r := csv.NewReader(strings.NewReader(s))
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	rows, err := r.ReadAll()
	if err != nil {
		return nil, errors.New("CSV 解析失败，请使用 UTF-8 编码：" + err.Error())
	}
	if len(rows) < 2 {
		return nil, errors.New("文件中没有数据行")
	}
	head := rows[0]
	out := []map[string]string{}
	for _, row := range rows[1:] {
		m := map[string]string{}
		empty := true
		for i, h := range head {
			if i < len(row) {
				m[strings.TrimSpace(h)] = strings.TrimSpace(row[i])
				if m[strings.TrimSpace(h)] != "" {
					empty = false
				}
			}
		}
		if !empty {
			out = append(out, m)
		}
	}
	return out, nil
}

// ImportHouseholds 列：户号,户主,户类型,网格,家庭住址,联系电话,家庭年收入,住房情况,备注
func ImportHouseholds(c *gin.Context) {
	rows, err := readCSV(c)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	gridByName := map[string]model.Grid{}
	var grids []model.Grid
	db.DB.Find(&grids)
	for _, g := range grids {
		gridByName[g.Name] = g
	}
	my := mw.MyGridIDs(c)
	okN, errs := 0, []string{}
	for i, m := range rows {
		line := i + 2
		g, ok := gridByName[m["网格"]]
		if !ok {
			errs = append(errs, fmt.Sprintf("第%d行：网格「%s」不存在", line, m["网格"]))
			continue
		}
		if my != nil && !mw.InGrids(my, g.ID) {
			errs = append(errs, fmt.Sprintf("第%d行：无权导入到网格「%s」", line, g.Name))
			continue
		}
		if m["户主"] == "" {
			errs = append(errs, fmt.Sprintf("第%d行：户主为空", line))
			continue
		}
		h := model.Household{HouseholdNo: m["户号"], HeadName: m["户主"], Type: m["户类型"], GridID: g.ID, GroupID: g.GroupID,
			Address: m["家庭住址"], Phone: m["联系电话"], Housing: m["住房情况"], Remark: m["备注"]}
		fmt.Sscanf(m["家庭年收入"], "%f", &h.Income)
		if h.Type == "" {
			h.Type = "一般户"
		}
		if h.HouseholdNo != "" {
			var n int64
			db.DB.Model(&model.Household{}).Where("household_no = ?", h.HouseholdNo).Count(&n)
			if n > 0 {
				errs = append(errs, fmt.Sprintf("第%d行：户号 %s 已存在", line, h.HouseholdNo))
				continue
			}
		} else {
			h.HouseholdNo = fmt.Sprintf("H%s%04d", time.Now().Format("0601"), (time.Now().UnixNano()/1000+int64(i))%10000)
		}
		if err := db.DB.Create(&h).Error; err != nil {
			errs = append(errs, fmt.Sprintf("第%d行：%v", line, err))
			continue
		}
		okN++
	}
	util.OK(c, gin.H{"success": okN, "failed": len(errs), "errors": errs})
}

// ImportVillagers 列：户号,姓名,身份证号,性别,出生日期,与户主关系,民族,政治面貌,文化程度,就业状态,务工地点,联系电话,特殊人群标签
func ImportVillagers(c *gin.Context) {
	rows, err := readCSV(c)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	my := mw.MyGridIDs(c)
	okN, errs := 0, []string{}
	touched := map[uint]bool{}
	for i, m := range rows {
		line := i + 2
		var h model.Household
		if db.DB.Where("household_no = ?", m["户号"]).First(&h).Error != nil {
			errs = append(errs, fmt.Sprintf("第%d行：户号「%s」不存在，请先导入家庭户", line, m["户号"]))
			continue
		}
		if my != nil && !mw.InGrids(my, h.GridID) {
			errs = append(errs, fmt.Sprintf("第%d行：无权导入该户", line))
			continue
		}
		v := model.Villager{HouseholdID: h.ID, Name: m["姓名"], IDCard: m["身份证号"], Gender: m["性别"], Birth: m["出生日期"],
			Relation: m["与户主关系"], Nation: m["民族"], Political: m["政治面貌"], Education: m["文化程度"], Employment: m["就业状态"],
			WorkPlace: m["务工地点"], Phone: m["联系电话"], Tags: m["特殊人群标签"]}
		if v.Nation == "" {
			v.Nation = "汉族"
		}
		if err := prepareVillager(&v); err != nil {
			errs = append(errs, fmt.Sprintf("第%d行：%s", line, err.Error()))
			continue
		}
		if v.IDCard != "" {
			var n int64
			db.DB.Model(&model.Villager{}).Where("id_card = ?", v.IDCard).Count(&n)
			if n > 0 {
				errs = append(errs, fmt.Sprintf("第%d行：身份证号已存在", line))
				continue
			}
		}
		if err := db.DB.Create(&v).Error; err != nil {
			errs = append(errs, fmt.Sprintf("第%d行：%v", line, err))
			continue
		}
		touched[h.ID] = true
		okN++
	}
	for id := range touched {
		recountMembers(id)
	}
	util.OK(c, gin.H{"success": okN, "failed": len(errs), "errors": errs})
}
