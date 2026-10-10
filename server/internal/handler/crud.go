// Package handler 业务接口。crud.go 提供通用的增删改查/分页/筛选/导出能力，
// 各业务模块通过声明 Resource 配置复用，新增模块只需几十行代码。
package handler

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qdmz/village-mgmt/internal/db"
	mw "github.com/qdmz/village-mgmt/internal/middleware"
	"github.com/qdmz/village-mgmt/internal/util"
	"gorm.io/gorm"
)

type Col[T any] struct {
	Title string
	Get   func(*T) string
}

type Resource[T any] struct {
	Title       string            // 中文名，用于导出文件名
	Search      []string          // kw 模糊搜索字段
	Filters     map[string]string // 查询参数 -> 字段（精确匹配）
	LikeFilters map[string]string // 查询参数 -> 字段（包含匹配）
	DateCol     string            // start/end 日期区间字段
	Order       string
	GridScope   bool   // 网格员仅能访问所负责网格（grid_id）
	OwnerCol    string // 网格员仅能访问本人数据的字段（如 user_id）
	ReadRoles   []string
	WriteRoles  []string
	DeleteRoles []string
	Prepare     func(c *gin.Context, item *T, old *T) error // 保存前处理（old 为 nil 表示新增）
	AfterSave   func(c *gin.Context, item *T, created bool)
	AfterDelete func(c *gin.Context, item *T)
	AfterList   func(c *gin.Context, items []T)
	AfterGet    func(c *gin.Context, item *T) any
	ExtraWhere  func(c *gin.Context, q *gorm.DB) *gorm.DB
	Export      []Col[T]
	Import      []ImportCol // 配置后自动提供 POST /xxx/import（CSV，前端可将 xlsx 转换后上传）
}

type ImportCol struct {
	Title string // 表头
	Tag   string // json 字段名；特殊值 grid_name（按网格名称匹配）
}

func hasRole(c *gin.Context, roles []string) bool {
	if len(roles) == 0 {
		return true
	}
	u := mw.CurrentUser(c)
	for _, r := range roles {
		if u != nil && u.Role == r {
			return true
		}
	}
	return false
}

func getUintField(v any, name string) uint {
	f := reflect.ValueOf(v).Elem().FieldByName(name)
	if !f.IsValid() {
		return 0
	}
	return uint(f.Uint())
}

func setUintField(v any, name string, val uint) {
	f := reflect.ValueOf(v).Elem().FieldByName(name)
	if f.IsValid() && f.CanSet() {
		f.SetUint(uint64(val))
	}
}

func colToField(col string) string {
	parts := strings.Split(col, "_")
	for i, p := range parts {
		if p == "id" {
			parts[i] = "ID"
		} else if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// query 构建列表查询（含权限范围与筛选条件）
func (r *Resource[T]) query(c *gin.Context) *gorm.DB {
	q := db.DB.Model(new(T))
	if r.GridScope {
		if ids := mw.MyGridIDs(c); ids != nil {
			q = q.Where("grid_id IN ?", ids)
		}
	}
	if r.OwnerCol != "" && mw.IsGrid(c) {
		q = q.Where(r.OwnerCol+" = ?", mw.CurrentUser(c).ID)
	}
	if kw := strings.TrimSpace(c.Query("kw")); kw != "" && len(r.Search) > 0 {
		conds := make([]string, len(r.Search))
		args := make([]any, len(r.Search))
		for i, s := range r.Search {
			conds[i] = s + " LIKE ?"
			args[i] = "%" + kw + "%"
		}
		q = q.Where("("+strings.Join(conds, " OR ")+")", args...)
	}
	for p, col := range r.Filters {
		if v := c.Query(p); v != "" {
			q = q.Where(col+" = ?", v)
		}
	}
	for p, col := range r.LikeFilters {
		if v := c.Query(p); v != "" {
			q = q.Where(col+" LIKE ?", "%"+v+"%")
		}
	}
	if r.DateCol != "" {
		if v := c.Query("start"); v != "" {
			q = q.Where(r.DateCol+" >= ?", v)
		}
		if v := c.Query("end"); v != "" {
			if r.DateCol == "created_at" && len(v) == 10 {
				v += " 23:59:59"
			}
			q = q.Where(r.DateCol+" <= ?", v)
		}
	}
	if r.ExtraWhere != nil {
		q = r.ExtraWhere(c, q)
	}
	return q
}

func (r *Resource[T]) order() string {
	if r.Order != "" {
		return r.Order
	}
	return "id DESC"
}

func (r *Resource[T]) List(c *gin.Context) {
	if !hasRole(c, r.ReadRoles) {
		util.Forbidden(c)
		return
	}
	page, size := util.Page(c)
	q := r.query(c)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		util.ServerError(c, err)
		return
	}
	items := []T{}
	if err := q.Order(r.order()).Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		util.ServerError(c, err)
		return
	}
	if r.AfterList != nil {
		r.AfterList(c, items)
	}
	util.OK(c, gin.H{"list": items, "total": total, "page": page, "size": size})
}

// load 读取单条并做权限范围校验
func (r *Resource[T]) load(c *gin.Context) (*T, bool) {
	item := new(T)
	if err := db.DB.First(item, util.ParamID(c)).Error; err != nil {
		util.NotFound(c)
		return nil, false
	}
	if !r.canAccess(c, item) {
		util.Forbidden(c)
		return nil, false
	}
	return item, true
}

func (r *Resource[T]) canAccess(c *gin.Context, item *T) bool {
	if !mw.IsGrid(c) {
		return true
	}
	if r.GridScope && !mw.InGrids(mw.MyGridIDs(c), getUintField(item, "GridID")) {
		return false
	}
	if r.OwnerCol != "" && getUintField(item, colToField(r.OwnerCol)) != mw.CurrentUser(c).ID {
		return false
	}
	return true
}

func (r *Resource[T]) Get(c *gin.Context) {
	if !hasRole(c, r.ReadRoles) {
		util.Forbidden(c)
		return
	}
	item, ok := r.load(c)
	if !ok {
		return
	}
	if r.AfterGet != nil {
		util.OK(c, r.AfterGet(c, item))
		return
	}
	util.OK(c, item)
}

func (r *Resource[T]) Create(c *gin.Context) {
	if !hasRole(c, r.WriteRoles) {
		util.Forbidden(c)
		return
	}
	item := new(T)
	if err := c.ShouldBindJSON(item); err != nil {
		util.BadRequest(c, "参数格式错误："+err.Error())
		return
	}
	setUintField(item, "ID", 0)
	if r.OwnerCol != "" && mw.IsGrid(c) {
		setUintField(item, colToField(r.OwnerCol), mw.CurrentUser(c).ID)
	}
	if r.Prepare != nil {
		if err := r.Prepare(c, item, nil); err != nil {
			util.BadRequest(c, err.Error())
			return
		}
	}
	if r.GridScope && mw.IsGrid(c) && !mw.InGrids(mw.MyGridIDs(c), getUintField(item, "GridID")) {
		util.BadRequest(c, "只能在自己负责的网格内录入数据")
		return
	}
	if err := db.DB.Create(item).Error; err != nil {
		util.ServerError(c, err)
		return
	}
	if r.AfterSave != nil {
		r.AfterSave(c, item, true)
	}
	util.OK(c, item)
}

func (r *Resource[T]) Update(c *gin.Context) {
	if !hasRole(c, r.WriteRoles) {
		util.Forbidden(c)
		return
	}
	old, ok := r.load(c)
	if !ok {
		return
	}
	id := getUintField(old, "ID")
	item := new(T)
	*item = *old
	if err := c.ShouldBindJSON(item); err != nil {
		util.BadRequest(c, "参数格式错误："+err.Error())
		return
	}
	setUintField(item, "ID", id)
	reflect.ValueOf(item).Elem().FieldByName("Base").FieldByName("CreatedAt").Set(
		reflect.ValueOf(old).Elem().FieldByName("Base").FieldByName("CreatedAt"))
	if r.OwnerCol != "" && mw.IsGrid(c) {
		setUintField(item, colToField(r.OwnerCol), mw.CurrentUser(c).ID)
	}
	if r.Prepare != nil {
		if err := r.Prepare(c, item, old); err != nil {
			util.BadRequest(c, err.Error())
			return
		}
	}
	if r.GridScope && mw.IsGrid(c) && !mw.InGrids(mw.MyGridIDs(c), getUintField(item, "GridID")) {
		util.BadRequest(c, "只能在自己负责的网格内维护数据")
		return
	}
	if err := db.DB.Save(item).Error; err != nil {
		util.ServerError(c, err)
		return
	}
	if r.AfterSave != nil {
		r.AfterSave(c, item, false)
	}
	util.OK(c, item)
}

func (r *Resource[T]) Delete(c *gin.Context) {
	roles := r.DeleteRoles
	if roles == nil {
		roles = r.WriteRoles
	}
	if !hasRole(c, roles) {
		util.Forbidden(c)
		return
	}
	item, ok := r.load(c)
	if !ok {
		return
	}
	if err := db.DB.Delete(item).Error; err != nil {
		util.ServerError(c, err)
		return
	}
	if r.AfterDelete != nil {
		r.AfterDelete(c, item)
	}
	util.OK(c, nil)
}

// ExportCSV 导出 CSV（带 BOM，Excel 可直接打开）
func (r *Resource[T]) ExportCSV(c *gin.Context) {
	if !hasRole(c, r.ReadRoles) || len(r.Export) == 0 {
		util.Forbidden(c)
		return
	}
	items := []T{}
	if err := r.query(c).Order(r.order()).Limit(20000).Find(&items).Error; err != nil {
		util.ServerError(c, err)
		return
	}
	if r.AfterList != nil {
		r.AfterList(c, items)
	}
	WriteCSV(c, r.Title, r.Export, items)
}

func WriteCSV[T any](c *gin.Context, title string, cols []Col[T], items []T) {
	name := fmt.Sprintf("%s_%s.csv", title, time.Now().Format("20060102150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\"export.csv\"; filename*=UTF-8''"+url.PathEscape(name))
	_, _ = c.Writer.Write([]byte("\xEF\xBB\xBF"))
	w := csv.NewWriter(c.Writer)
	head := make([]string, len(cols))
	for i, col := range cols {
		head[i] = col.Title
	}
	_ = w.Write(head)
	for i := range items {
		row := make([]string, len(cols))
		for j, col := range cols {
			row[j] = col.Get(&items[i])
		}
		_ = w.Write(row)
	}
	w.Flush()
}

// ImportFile 通用导入
func (r *Resource[T]) ImportFile(c *gin.Context) {
	if !hasRole(c, r.WriteRoles) {
		util.Forbidden(c)
		return
	}
	rows, err := readCSV(c)
	if err != nil {
		util.BadRequest(c, err.Error())
		return
	}
	grids := gridNames()
	byName := map[string]uint{}
	for id, n := range grids {
		byName[n] = id
	}
	okN, errs := 0, []string{}
	for i, m := range rows {
		line := i + 2
		item := new(T)
		bad := ""
		for _, col := range r.Import {
			v, ok := m[col.Title]
			if !ok || v == "" {
				continue
			}
			if col.Tag == "grid_name" {
				id, ok := byName[v]
				if !ok {
					bad = "网格「" + v + "」不存在"
					break
				}
				setUintField(item, "GridID", id)
				continue
			}
			if err := setByTag(item, col.Tag, v); err != nil {
				bad = col.Title + "：" + err.Error()
				break
			}
		}
		if bad == "" && r.OwnerCol != "" && mw.IsGrid(c) {
			setUintField(item, colToField(r.OwnerCol), mw.CurrentUser(c).ID)
		}
		if bad == "" && r.Prepare != nil {
			if err := r.Prepare(c, item, nil); err != nil {
				bad = err.Error()
			}
		}
		if bad == "" && r.GridScope && mw.IsGrid(c) && !mw.InGrids(mw.MyGridIDs(c), getUintField(item, "GridID")) {
			bad = "无权导入到该网格"
		}
		if bad == "" {
			if err := db.DB.Create(item).Error; err != nil {
				bad = err.Error()
			}
		}
		if bad != "" {
			errs = append(errs, fmt.Sprintf("第%d行：%s", line, bad))
			continue
		}
		if r.AfterSave != nil {
			r.AfterSave(c, item, true)
		}
		okN++
	}
	util.OK(c, gin.H{"success": okN, "failed": len(errs), "errors": errs})
}

func setByTag(item any, tag, v string) error {
	rv := reflect.ValueOf(item).Elem()
	var f reflect.Value
	var walk func(reflect.Value) bool
	walk = func(sv reflect.Value) bool {
		st := sv.Type()
		for i := 0; i < st.NumField(); i++ {
			sf := st.Field(i)
			if sf.Anonymous {
				if walk(sv.Field(i)) {
					return true
				}
				continue
			}
			if strings.Split(sf.Tag.Get("json"), ",")[0] == tag {
				f = sv.Field(i)
				return true
			}
		}
		return false
	}
	if !walk(rv) {
		return nil
	}
	v = strings.TrimSpace(v)
	switch f.Kind() {
	case reflect.String:
		f.SetString(v)
	case reflect.Int, reflect.Int64, reflect.Int32:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("应为数字")
		}
		f.SetInt(int64(n))
	case reflect.Uint, reflect.Uint64, reflect.Uint32:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return fmt.Errorf("应为数字")
		}
		f.SetUint(n)
	case reflect.Float64, reflect.Float32:
		n, err := strconv.ParseFloat(strings.ReplaceAll(v, ",", ""), 64)
		if err != nil {
			return fmt.Errorf("应为数字")
		}
		f.SetFloat(n)
	case reflect.Bool:
		f.SetBool(v == "是" || v == "1" || strings.EqualFold(v, "true"))
	}
	return nil
}

func (r *Resource[T]) Register(g *gin.RouterGroup, path string) {
	if len(r.Import) > 0 {
		g.POST(path+"/import", r.ImportFile)
	}
	g.GET(path, r.List)
	g.GET(path+"/export", r.ExportCSV)
	g.GET(path+"/:id", r.Get)
	g.POST(path, r.Create)
	g.PUT(path+"/:id", r.Update)
	g.DELETE(path+"/:id", r.Delete)
}

// 常用格式化
func S(s string) string     { return s }
func F(f float64) string    { return fmt.Sprintf("%.2f", f) }
func I(i int) string        { return fmt.Sprintf("%d", i) }
func TS(t time.Time) string { return t.Format("2006-01-02 15:04") }
func B(b bool) string {
	if b {
		return "是"
	}
	return "否"
}
