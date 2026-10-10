package handler

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qdmz/village-mgmt/internal/db"
	mw "github.com/qdmz/village-mgmt/internal/middleware"
	"github.com/qdmz/village-mgmt/internal/model"
	"github.com/qdmz/village-mgmt/internal/util"
	"gorm.io/gorm"
)

var (
	adminCadre = []string{model.RoleAdmin, model.RoleCadre}
	adminOnly  = []string{model.RoleAdmin}
)

func today() string { return time.Now().Format("2006-01-02") }

// gridNames 网格名称映射（用于导出）
func gridNames() map[uint]string {
	var gs []model.Grid
	db.DB.Find(&gs)
	m := map[uint]string{}
	for _, g := range gs {
		m[g.ID] = g.Name
	}
	return m
}

func groupNames() map[uint]string {
	var gs []model.Group
	db.DB.Find(&gs)
	m := map[uint]string{}
	for _, g := range gs {
		m[g.ID] = g.Name
	}
	return m
}

func userName(id uint) string {
	if id == 0 {
		return ""
	}
	var u model.User
	if db.DB.Select("real_name").First(&u, id).Error != nil {
		return ""
	}
	return u.RealName
}

var idCardRe = regexp.MustCompile(`^\d{17}[\dXx]$`)

// ---------------- 村情与网格 ----------------

var VillageRes = &Resource[model.Village]{Title: "村情信息", Search: []string{"name"}, WriteRoles: adminCadre, DeleteRoles: adminOnly}

var GroupRes = &Resource[model.Group]{Title: "村民小组", Search: []string{"name", "leader"}, Filters: map[string]string{"village_id": "village_id"},
	Order: "sort, id", WriteRoles: adminCadre,
	Prepare: func(c *gin.Context, g *model.Group, old *model.Group) error {
		if strings.TrimSpace(g.Name) == "" {
			return errors.New("请填写小组名称")
		}
		if g.VillageID == 0 {
			var v model.Village
			db.DB.First(&v)
			g.VillageID = v.ID
		}
		return nil
	},
	Export: []Col[model.Group]{{"小组名称", func(g *model.Group) string { return g.Name }}, {"组长", func(g *model.Group) string { return g.Leader }}, {"联系电话", func(g *model.Group) string { return g.Phone }}},
}

var GridRes = &Resource[model.Grid]{Title: "网格", Search: []string{"name", "code", "worker_name", "scope"},
	Filters: map[string]string{"group_id": "group_id", "worker_id": "worker_id"}, Order: "sort, id", WriteRoles: adminCadre,
	Prepare: func(c *gin.Context, g *model.Grid, old *model.Grid) error {
		if strings.TrimSpace(g.Name) == "" {
			return errors.New("请填写网格名称")
		}
		if g.VillageID == 0 {
			var v model.Village
			db.DB.First(&v)
			g.VillageID = v.ID
		}
		g.WorkerName = userName(g.WorkerID)
		return nil
	},
	Export: []Col[model.Grid]{{"网格名称", func(g *model.Grid) string { return g.Name }}, {"编码", func(g *model.Grid) string { return g.Code }},
		{"网格员", func(g *model.Grid) string { return g.WorkerName }}, {"范围", func(g *model.Grid) string { return g.Scope }}},
}

// ---------------- 户籍与人口 ----------------

func recountMembers(hid uint) {
	if hid == 0 {
		return
	}
	var n int64
	db.DB.Model(&model.Villager{}).Where("household_id = ?", hid).Count(&n)
	db.DB.Model(&model.Household{}).Where("id = ?", hid).Update("member_count", n)
}

var HouseholdRes = &Resource[model.Household]{Title: "户籍台账", Search: []string{"household_no", "head_name", "address", "phone"},
	Filters: map[string]string{"grid_id": "grid_id", "group_id": "group_id", "type": "type"}, GridScope: true,
	DeleteRoles: adminCadre,
	Prepare: func(c *gin.Context, h *model.Household, old *model.Household) error {
		if strings.TrimSpace(h.HeadName) == "" {
			return errors.New("请填写户主姓名")
		}
		if h.GridID == 0 {
			return errors.New("请选择所属网格")
		}
		var g model.Grid
		if db.DB.First(&g, h.GridID).Error == nil {
			h.GroupID = g.GroupID
		}
		if h.HouseholdNo == "" {
			h.HouseholdNo = fmt.Sprintf("H%s%04d", time.Now().Format("0601"), time.Now().UnixNano()%10000)
		}
		if h.Type == "" {
			h.Type = "一般户"
		}
		return nil
	},
	AfterDelete: func(c *gin.Context, h *model.Household) {
		db.DB.Where("household_id = ?", h.ID).Delete(&model.Villager{})
	},
	AfterGet: func(c *gin.Context, h *model.Household) any {
		members := []model.Villager{}
		db.DB.Where("household_id = ?", h.ID).Order("id").Find(&members)
		visits := []model.WorkLog{}
		db.DB.Where("household_id = ?", h.ID).Order("log_date DESC, id DESC").Limit(50).Find(&visits)
		return gin.H{"household": h, "members": members, "visits": visits}
	},
}

func init() {
	gn := map[uint]string{}
	HouseholdRes.Export = []Col[model.Household]{
		{"户号", func(h *model.Household) string { return h.HouseholdNo }},
		{"户主", func(h *model.Household) string { return h.HeadName }},
		{"户类型", func(h *model.Household) string { return h.Type }},
		{"所属网格", func(h *model.Household) string {
			if len(gn) == 0 {
				gn = gridNames()
			}
			return gn[h.GridID]
		}},
		{"家庭住址", func(h *model.Household) string { return h.Address }},
		{"联系电话", func(h *model.Household) string { return h.Phone }},
		{"人口数", func(h *model.Household) string { return I(h.MemberCount) }},
		{"家庭年收入(元)", func(h *model.Household) string { return F(h.Income) }},
		{"住房情况", func(h *model.Household) string { return h.Housing }},
		{"备注", func(h *model.Household) string { return h.Remark }},
	}
	HouseholdRes.AfterList = func(c *gin.Context, _ []model.Household) { gn = map[uint]string{} }
}

var VillagerRes = &Resource[model.Villager]{Title: "人口台账", Search: []string{"name", "id_card", "phone"},
	Filters: map[string]string{"household_id": "household_id", "grid_id": "grid_id", "gender": "gender", "political": "political",
		"education": "education", "employment": "employment"},
	LikeFilters: map[string]string{"tag": "tags"}, GridScope: true, DeleteRoles: adminCadre,
	ExtraWhere: func(c *gin.Context, q *gorm.DB) *gorm.DB {
		year := time.Now().Year()
		if v, err := strconv.Atoi(c.Query("age_min")); err == nil {
			q = q.Where("birth <= ?", fmt.Sprintf("%04d-12-31", year-v))
		}
		if v, err := strconv.Atoi(c.Query("age_max")); err == nil {
			q = q.Where("birth >= ?", fmt.Sprintf("%04d-01-01", year-v))
		}
		return q
	},
	Prepare: func(c *gin.Context, v *model.Villager, old *model.Villager) error {
		return prepareVillager(v)
	},
	AfterList: func(c *gin.Context, items []model.Villager) {
		for i := range items {
			items[i].IDCard = util.MaskIDCard(items[i].IDCard)
		}
	},
	AfterSave: func(c *gin.Context, v *model.Villager, created bool) { recountMembers(v.HouseholdID) },
	AfterDelete: func(c *gin.Context, v *model.Villager) { recountMembers(v.HouseholdID) },
	Export: []Col[model.Villager]{
		{"姓名", func(v *model.Villager) string { return v.Name }},
		{"身份证号(脱敏)", func(v *model.Villager) string { return v.IDCard }},
		{"性别", func(v *model.Villager) string { return v.Gender }},
		{"出生日期", func(v *model.Villager) string { return v.Birth }},
		{"与户主关系", func(v *model.Villager) string { return v.Relation }},
		{"民族", func(v *model.Villager) string { return v.Nation }},
		{"政治面貌", func(v *model.Villager) string { return v.Political }},
		{"文化程度", func(v *model.Villager) string { return v.Education }},
		{"就业状态", func(v *model.Villager) string { return v.Employment }},
		{"务工地点", func(v *model.Villager) string { return v.WorkPlace }},
		{"联系电话", func(v *model.Villager) string { return v.Phone }},
		{"特殊人群标签", func(v *model.Villager) string { return v.Tags }},
	},
}

func prepareVillager(v *model.Villager) error {
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		return errors.New("请填写姓名")
	}
	if v.HouseholdID == 0 {
		return errors.New("请选择所属家庭户")
	}
	var h model.Household
	if db.DB.First(&h, v.HouseholdID).Error != nil {
		return errors.New("所属家庭户不存在")
	}
	v.GridID = h.GridID
	v.IDCard = strings.ToUpper(strings.TrimSpace(v.IDCard))
	if v.IDCard != "" {
		if strings.Contains(v.IDCard, "*") {
			return errors.New("请填写完整身份证号")
		}
		if !idCardRe.MatchString(v.IDCard) {
			return errors.New("身份证号格式不正确")
		}
		if v.Birth == "" {
			v.Birth = v.IDCard[6:10] + "-" + v.IDCard[10:12] + "-" + v.IDCard[12:14]
		}
		if v.Gender == "" {
			if (v.IDCard[16]-'0')%2 == 1 {
				v.Gender = "男"
			} else {
				v.Gender = "女"
			}
		}
	}
	v.Tags = strings.Trim(strings.ReplaceAll(strings.ReplaceAll(v.Tags, "，", ","), " ", ""), ",")
	return nil
}

// ---------------- 网格巡查 ----------------

var PatrolRes = &Resource[model.Patrol]{Title: "巡查记录", Search: []string{"route", "content", "problems", "user_name"},
	Filters: map[string]string{"grid_id": "grid_id", "user_id": "user_id"}, DateCol: "patrol_date", GridScope: true,
	Order: "patrol_date DESC, id DESC",
	Prepare: func(c *gin.Context, p *model.Patrol, old *model.Patrol) error {
		u := mw.CurrentUser(c)
		if old == nil {
			p.UserID, p.UserName = u.ID, u.RealName
		} else if u.Role == model.RoleGrid && old.UserID != u.ID {
			return errors.New("只能修改自己的巡查记录")
		}
		if p.GridID == 0 {
			return errors.New("请选择巡查网格")
		}
		if p.PatrolDate == "" {
			p.PatrolDate = today()
		}
		if strings.TrimSpace(p.Content) == "" {
			return errors.New("请填写巡查情况")
		}
		return nil
	},
}

func init() {
	PatrolRes.Export = []Col[model.Patrol]{
		{"巡查日期", func(p *model.Patrol) string { return p.PatrolDate }},
		{"时间", func(p *model.Patrol) string { return p.StartTime + "-" + p.EndTime }},
		{"网格", func(p *model.Patrol) string { return gridNameCached(p.GridID) }},
		{"巡查人", func(p *model.Patrol) string { return p.UserName }},
		{"巡查路线", func(p *model.Patrol) string { return p.Route }},
		{"巡查情况", func(p *model.Patrol) string { return p.Content }},
		{"发现问题", func(p *model.Patrol) string { return p.Problems }},
	}
}

var gnCache = struct {
	m  map[uint]string
	at time.Time
}{}

func gridNameCached(id uint) string {
	if time.Since(gnCache.at) > 10*time.Second {
		gnCache.m, gnCache.at = gridNames(), time.Now()
	}
	return gnCache.m[id]
}

// ---------------- 任务督办 ----------------

var TaskRes = &Resource[model.Task]{Title: "督办任务", Search: []string{"title", "content", "assignee_name"},
	Filters: map[string]string{"status": "status", "priority": "priority", "assignee_id": "assignee_id", "category": "category"},
	OwnerCol: "assignee_id", WriteRoles: adminCadre, Order: "CASE WHEN status = '已完成' THEN 1 ELSE 0 END, deadline, id DESC",
	ExtraWhere: func(c *gin.Context, q *gorm.DB) *gorm.DB {
		if c.Query("overdue") == "1" {
			q = q.Where("deadline < ? AND status <> ?", today(), "已完成")
		}
		if c.Query("mine") == "1" {
			q = q.Where("assignee_id = ?", mw.CurrentUser(c).ID)
		}
		return q
	},
	Prepare: func(c *gin.Context, t *model.Task, old *model.Task) error {
		if strings.TrimSpace(t.Title) == "" {
			return errors.New("请填写任务标题")
		}
		if t.AssigneeID == 0 {
			return errors.New("请选择责任人")
		}
		if t.Deadline == "" {
			return errors.New("请选择截止日期")
		}
		t.AssigneeName = userName(t.AssigneeID)
		if old == nil {
			u := mw.CurrentUser(c)
			t.CreatorID, t.CreatorName = u.ID, u.RealName
		}
		if t.Priority == "" {
			t.Priority = "中"
		}
		normalizeTask(t)
		return nil
	},
	AfterList: func(c *gin.Context, items []model.Task) {
		for i := range items {
			markOverdue(&items[i])
		}
	},
	AfterGet: func(c *gin.Context, t *model.Task) any {
		markOverdue(t)
		ups := []model.TaskUpdate{}
		db.DB.Where("task_id = ?", t.ID).Order("id DESC").Find(&ups)
		return gin.H{"task": t, "updates": ups}
	},
	AfterDelete: func(c *gin.Context, t *model.Task) { db.DB.Where("task_id = ?", t.ID).Delete(&model.TaskUpdate{}) },
	Export: []Col[model.Task]{
		{"任务标题", func(t *model.Task) string { return t.Title }},
		{"类别", func(t *model.Task) string { return t.Category }},
		{"优先级", func(t *model.Task) string { return t.Priority }},
		{"责任人", func(t *model.Task) string { return t.AssigneeName }},
		{"交办人", func(t *model.Task) string { return t.CreatorName }},
		{"截止日期", func(t *model.Task) string { return t.Deadline }},
		{"状态", func(t *model.Task) string { return t.Status }},
		{"进度(%)", func(t *model.Task) string { return I(t.Progress) }},
		{"是否逾期", func(t *model.Task) string { return B(t.Overdue) }},
	},
}

func normalizeTask(t *model.Task) {
	if t.Progress < 0 {
		t.Progress = 0
	}
	if t.Progress > 100 {
		t.Progress = 100
	}
	if t.Status == "" {
		t.Status = "待办"
	}
	if t.Progress == 100 {
		t.Status = "已完成"
	}
	if t.Status == "已完成" {
		t.Progress = 100
		if t.CompletedAt == nil {
			now := time.Now()
			t.CompletedAt = &now
		}
	} else {
		t.CompletedAt = nil
		if t.Progress > 0 && t.Status == "待办" {
			t.Status = "进行中"
		}
	}
}

func markOverdue(t *model.Task) {
	t.Overdue = t.Status != "已完成" && t.Deadline != "" && t.Deadline < today()
}

// ---------------- 工作日志 / 民情日记 ----------------

var WorkLogRes = &Resource[model.WorkLog]{Title: "工作日志", Search: []string{"title", "content", "household_name", "user_name"},
	Filters: map[string]string{"type": "type", "user_id": "user_id", "household_id": "household_id", "grid_id": "grid_id"},
	DateCol: "log_date", OwnerCol: "user_id", Order: "log_date DESC, id DESC",
	ExtraWhere: func(c *gin.Context, q *gorm.DB) *gorm.DB {
		if c.Query("mine") == "1" {
			q = q.Where("user_id = ?", mw.CurrentUser(c).ID)
		}
		return q
	},
	Prepare: func(c *gin.Context, w *model.WorkLog, old *model.WorkLog) error {
		u := mw.CurrentUser(c)
		if old != nil && old.UserID != u.ID && u.Role != model.RoleAdmin {
			return errors.New("只能修改本人的日志")
		}
		if old == nil {
			w.UserID, w.UserName = u.ID, u.RealName
		} else {
			w.UserID, w.UserName = old.UserID, old.UserName
		}
		if w.Type == "" {
			w.Type = "工作日志"
		}
		if w.LogDate == "" {
			w.LogDate = today()
		}
		if strings.TrimSpace(w.Content) == "" {
			return errors.New("请填写日志内容")
		}
		w.HouseholdName = ""
		if w.HouseholdID > 0 {
			var h model.Household
			if db.DB.First(&h, w.HouseholdID).Error != nil {
				return errors.New("走访户不存在")
			}
			if ids := mw.MyGridIDs(c); ids != nil && !mw.InGrids(ids, h.GridID) {
				return errors.New("只能走访本网格内的家庭户")
			}
			w.HouseholdName, w.GridID = h.HeadName, h.GridID
		} else if w.Type == "走访记录" {
			return errors.New("走访记录请选择走访户")
		}
		return nil
	},
	Export: []Col[model.WorkLog]{
		{"日期", func(w *model.WorkLog) string { return w.LogDate }},
		{"类型", func(w *model.WorkLog) string { return w.Type }},
		{"记录人", func(w *model.WorkLog) string { return w.UserName }},
		{"标题", func(w *model.WorkLog) string { return w.Title }},
		{"内容", func(w *model.WorkLog) string { return w.Content }},
		{"走访户", func(w *model.WorkLog) string { return w.HouseholdName }},
		{"群众诉求", func(w *model.WorkLog) string { return w.Demand }},
		{"跟进情况", func(w *model.WorkLog) string { return w.Followup }},
	},
}

// ---------------- 通知公告 / 村务公开 ----------------

var NoticeRes = &Resource[model.Notice]{Title: "通知公告", Search: []string{"title", "content"},
	Filters: map[string]string{"category": "category", "status": "status"}, DateCol: "publish_date",
	WriteRoles: adminCadre, Order: "top DESC, publish_date DESC, id DESC",
	ExtraWhere: func(c *gin.Context, q *gorm.DB) *gorm.DB {
		if mw.IsGrid(c) {
			q = q.Where("status = ?", "已发布")
		}
		return q
	},
	Prepare: func(c *gin.Context, n *model.Notice, old *model.Notice) error {
		if strings.TrimSpace(n.Title) == "" {
			return errors.New("请填写标题")
		}
		if n.Category == "" {
			n.Category = "通知公告"
		}
		if n.Status == "" {
			n.Status = "已发布"
		}
		if n.PublishDate == "" {
			n.PublishDate = today()
		}
		if n.Publisher == "" {
			n.Publisher = mw.CurrentUser(c).RealName
		}
		return nil
	},
	AfterList: func(c *gin.Context, items []model.Notice) {
		if c.Query("brief") == "1" {
			for i := range items {
				r := []rune(stripTags(items[i].Content))
				if len(r) > 80 {
					r = r[:80]
				}
				items[i].Content = string(r)
			}
		}
	},
	AfterGet: func(c *gin.Context, n *model.Notice) any {
		if mw.IsGrid(c) && n.Status != "已发布" {
			return nil
		}
		db.DB.Model(n).UpdateColumn("views", gorm.Expr("views + 1"))
		n.Views++
		return n
	},
	Export: []Col[model.Notice]{
		{"标题", func(n *model.Notice) string { return n.Title }},
		{"栏目", func(n *model.Notice) string { return n.Category }},
		{"发布人", func(n *model.Notice) string { return n.Publisher }},
		{"发布日期", func(n *model.Notice) string { return n.PublishDate }},
		{"状态", func(n *model.Notice) string { return n.Status }},
		{"浏览量", func(n *model.Notice) string { return I(n.Views) }},
	},
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

func stripTags(s string) string { return tagRe.ReplaceAllString(s, "") }

// ---------------- 产业项目 / 资金 / 资产 ----------------

var ProjectRes = &Resource[model.Project]{Title: "产业项目", Search: []string{"name", "leader", "location"},
	Filters: map[string]string{"type": "type", "status": "status"}, ReadRoles: adminCadre, WriteRoles: adminCadre,
	Prepare: func(c *gin.Context, p *model.Project, old *model.Project) error {
		if strings.TrimSpace(p.Name) == "" {
			return errors.New("请填写项目名称")
		}
		if p.Progress < 0 || p.Progress > 100 {
			return errors.New("进度应在 0-100 之间")
		}
		if p.Status == "" {
			p.Status = "规划中"
		}
		return nil
	},
	AfterGet: func(c *gin.Context, p *model.Project) any {
		funds := []model.Fund{}
		db.DB.Where("project_id = ?", p.ID).Order("date DESC").Find(&funds)
		return gin.H{"project": p, "funds": funds}
	},
	Export: []Col[model.Project]{
		{"项目名称", func(p *model.Project) string { return p.Name }},
		{"类型", func(p *model.Project) string { return p.Type }},
		{"负责人", func(p *model.Project) string { return p.Leader }},
		{"地点", func(p *model.Project) string { return p.Location }},
		{"预算(万元)", func(p *model.Project) string { return F(p.Budget) }},
		{"已投入(万元)", func(p *model.Project) string { return F(p.Invested) }},
		{"进度(%)", func(p *model.Project) string { return I(p.Progress) }},
		{"状态", func(p *model.Project) string { return p.Status }},
		{"开始日期", func(p *model.Project) string { return p.StartDate }},
		{"计划完成", func(p *model.Project) string { return p.EndDate }},
		{"带动效益", func(p *model.Project) string { return p.Benefit }},
	},
	Import: []ImportCol{{"项目名称", "name"}, {"类型", "type"}, {"负责人", "leader"}, {"地点", "location"}, {"预算(万元)", "budget"},
		{"已投入(万元)", "invested"}, {"进度(%)", "progress"}, {"状态", "status"}, {"开始日期", "start_date"}, {"计划完成", "end_date"}, {"带动效益", "benefit"}},
}

var FundRes = &Resource[model.Fund]{Title: "资金台账", Search: []string{"summary", "voucher_no", "handler", "category"},
	Filters: map[string]string{"type": "type", "category": "category", "project_id": "project_id"}, DateCol: "date",
	ReadRoles: adminCadre, WriteRoles: adminCadre, Order: "date DESC, id DESC",
	Prepare: func(c *gin.Context, f *model.Fund, old *model.Fund) error {
		if f.Type != "收入" && f.Type != "支出" {
			return errors.New("请选择收支类型")
		}
		if f.Amount <= 0 {
			return errors.New("金额必须大于 0")
		}
		if f.Date == "" {
			f.Date = today()
		}
		return nil
	},
	Export: []Col[model.Fund]{
		{"日期", func(f *model.Fund) string { return f.Date }},
		{"凭证号", func(f *model.Fund) string { return f.VoucherNo }},
		{"收支", func(f *model.Fund) string { return f.Type }},
		{"科目", func(f *model.Fund) string { return f.Category }},
		{"金额(元)", func(f *model.Fund) string { return F(f.Amount) }},
		{"摘要", func(f *model.Fund) string { return f.Summary }},
		{"经办人", func(f *model.Fund) string { return f.Handler }},
	},
	Import: []ImportCol{{"日期", "date"}, {"凭证号", "voucher_no"}, {"收支", "type"}, {"科目", "category"}, {"金额(元)", "amount"}, {"摘要", "summary"}, {"经办人", "handler"}},
}

// FundSummary 资金收支汇总（随筛选条件变化）
func FundSummary(c *gin.Context) {
	type row struct {
		Type  string
		Total float64
	}
	var rows []row
	FundRes.query(c).Select("type, COALESCE(SUM(amount),0) AS total").Group("type").Scan(&rows)
	var in, out float64
	for _, r := range rows {
		if r.Type == "收入" {
			in = r.Total
		} else {
			out = r.Total
		}
	}
	util.OK(c, gin.H{"income": in, "expense": out, "balance": in - out})
}

var AssetRes = &Resource[model.Asset]{Title: "集体资产", Search: []string{"code", "name", "location", "keeper"},
	Filters: map[string]string{"category": "category", "status": "status"}, ReadRoles: adminCadre, WriteRoles: adminCadre,
	Prepare: func(c *gin.Context, a *model.Asset, old *model.Asset) error {
		if strings.TrimSpace(a.Name) == "" {
			return errors.New("请填写资产名称")
		}
		if a.Code == "" {
			a.Code = fmt.Sprintf("ZC%s%03d", time.Now().Format("060102"), time.Now().UnixNano()%1000)
		}
		return nil
	},
	Export: []Col[model.Asset]{
		{"资产编号", func(a *model.Asset) string { return a.Code }},
		{"资产名称", func(a *model.Asset) string { return a.Name }},
		{"类别", func(a *model.Asset) string { return a.Category }},
		{"数量", func(a *model.Asset) string { return fmt.Sprintf("%g", a.Quantity) }},
		{"单位", func(a *model.Asset) string { return a.Unit }},
		{"价值(万元)", func(a *model.Asset) string { return F(a.Value) }},
		{"坐落位置", func(a *model.Asset) string { return a.Location }},
		{"状态", func(a *model.Asset) string { return a.Status }},
		{"管理人", func(a *model.Asset) string { return a.Keeper }},
		{"取得日期", func(a *model.Asset) string { return a.AcquiredAt }},
	},
	Import: []ImportCol{{"资产编号", "code"}, {"资产名称", "name"}, {"类别", "category"}, {"数量", "quantity"}, {"单位", "unit"}, {"价值(万元)", "value"},
		{"坐落位置", "location"}, {"状态", "status"}, {"管理人", "keeper"}, {"取得日期", "acquired_at"}},
}

// ---------------- 会议活动 ----------------

var MeetingRes = &Resource[model.Meeting]{Title: "会议活动", Search: []string{"title", "topics", "host", "attendees"},
	Filters: map[string]string{"type": "type"}, DateCol: "meet_date", WriteRoles: adminCadre, Order: "meet_date DESC, id DESC",
	Prepare: func(c *gin.Context, m *model.Meeting, old *model.Meeting) error {
		if strings.TrimSpace(m.Title) == "" {
			return errors.New("请填写会议名称")
		}
		if m.MeetDate == "" {
			return errors.New("请选择会议日期")
		}
		m.Attendees = strings.Trim(strings.ReplaceAll(m.Attendees, "，", ","), ", ")
		if m.Attendees != "" {
			m.AttendCount = len(strings.Split(m.Attendees, ","))
		}
		return nil
	},
	Export: []Col[model.Meeting]{
		{"会议名称", func(m *model.Meeting) string { return m.Title }},
		{"类型", func(m *model.Meeting) string { return m.Type }},
		{"日期", func(m *model.Meeting) string { return m.MeetDate }},
		{"地点", func(m *model.Meeting) string { return m.Location }},
		{"主持人", func(m *model.Meeting) string { return m.Host }},
		{"记录人", func(m *model.Meeting) string { return m.Recorder }},
		{"应到", func(m *model.Meeting) string { return I(m.ShouldCount) }},
		{"实到", func(m *model.Meeting) string { return I(m.AttendCount) }},
		{"参会人员", func(m *model.Meeting) string { return m.Attendees }},
		{"议题", func(m *model.Meeting) string { return m.Topics }},
		{"决议", func(m *model.Meeting) string { return m.Resolution }},
	},
}

// ---------------- 操作日志 ----------------

var OpLogRes = &Resource[model.OpLog]{Title: "操作日志", Search: []string{"username", "real_name", "path", "module", "ip"},
	Filters: map[string]string{"module": "module", "action": "action", "user_id": "user_id"}, DateCol: "created_at",
	ReadRoles: adminOnly, WriteRoles: []string{"__none__"},
	Export: []Col[model.OpLog]{
		{"时间", func(o *model.OpLog) string { return TS(o.CreatedAt) }},
		{"账号", func(o *model.OpLog) string { return o.Username }},
		{"姓名", func(o *model.OpLog) string { return o.RealName }},
		{"模块", func(o *model.OpLog) string { return o.Module }},
		{"操作", func(o *model.OpLog) string { return o.Action }},
		{"请求", func(o *model.OpLog) string { return o.Method + " " + o.Path }},
		{"IP", func(o *model.OpLog) string { return o.IP }},
		{"状态码", func(o *model.OpLog) string { return I(o.Status) }},
	},
}
