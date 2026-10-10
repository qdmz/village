package handler

// 党建引领、民生保障、乡村治理、致富带头人、门户内容、数据驾驶舱、备份恢复

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qdmz/village-mgmt/internal/db"
	mw "github.com/qdmz/village-mgmt/internal/middleware"
	"github.com/qdmz/village-mgmt/internal/model"
	"github.com/qdmz/village-mgmt/internal/util"
	"gorm.io/gorm"
)

// ---------------- 党建引领 ----------------

var PartyOrgRes = &Resource[model.PartyOrg]{Title: "党组织", Search: []string{"name", "secretary"}, WriteRoles: adminCadre, Order: "sort, id",
	Prepare: func(c *gin.Context, o *model.PartyOrg, old *model.PartyOrg) error {
		if strings.TrimSpace(o.Name) == "" {
			return errors.New("请填写组织名称")
		}
		if old != nil && o.ParentID == old.ID {
			return errors.New("上级组织不能是自己")
		}
		return nil
	},
	AfterSave: func(c *gin.Context, o *model.PartyOrg, created bool) {
		db.DB.Model(&model.PartyMember{}).Where("org_id = ?", o.ID).Update("org_name", o.Name)
	},
}

func PartyOrgTree(c *gin.Context) {
	var orgs []model.PartyOrg
	db.DB.Order("sort, id").Find(&orgs)
	type cnt struct {
		OrgID uint
		N     int
	}
	var cs []cnt
	db.DB.Model(&model.PartyMember{}).Select("org_id, COUNT(*) AS n").Group("org_id").Scan(&cs)
	cm := map[uint]int{}
	for _, x := range cs {
		cm[x.OrgID] = x.N
	}
	var build func(pid uint) []gin.H
	build = func(pid uint) []gin.H {
		out := []gin.H{}
		for _, o := range orgs {
			if o.ParentID == pid {
				out = append(out, gin.H{"id": o.ID, "label": o.Name, "type": o.Type, "secretary": o.Secretary, "members": o.Members,
					"duty": o.Duty, "count": cm[o.ID], "data": o, "children": build(o.ID)})
			}
		}
		return out
	}
	util.OK(c, build(0))
}

var PartyMemberRes = &Resource[model.PartyMember]{Title: "党员花名册", Search: []string{"name", "phone", "post", "work_place"},
	Filters: map[string]string{"org_id": "org_id", "type": "type", "gender": "gender", "mobile": "mobile"}, WriteRoles: adminCadre, Order: "join_date, id",
	Prepare: func(c *gin.Context, p *model.PartyMember, old *model.PartyMember) error {
		if strings.TrimSpace(p.Name) == "" {
			return errors.New("请填写姓名")
		}
		if p.Type == "" {
			p.Type = "正式党员"
		}
		if p.Nation == "" {
			p.Nation = "汉族"
		}
		p.OrgName = ""
		if p.OrgID > 0 {
			var o model.PartyOrg
			if db.DB.First(&o, p.OrgID).Error == nil {
				p.OrgName = o.Name
			}
		}
		return nil
	},
	Export: []Col[model.PartyMember]{
		{"姓名", func(p *model.PartyMember) string { return p.Name }}, {"性别", func(p *model.PartyMember) string { return p.Gender }},
		{"出生日期", func(p *model.PartyMember) string { return p.Birth }}, {"民族", func(p *model.PartyMember) string { return p.Nation }},
		{"文化程度", func(p *model.PartyMember) string { return p.Education }}, {"入党日期", func(p *model.PartyMember) string { return p.JoinDate }},
		{"转正日期", func(p *model.PartyMember) string { return p.FormalDate }}, {"党员类型", func(p *model.PartyMember) string { return p.Type }},
		{"流动党员", func(p *model.PartyMember) string { return B(p.Mobile) }}, {"所属组织", func(p *model.PartyMember) string { return p.OrgName }},
		{"党内职务", func(p *model.PartyMember) string { return p.Post }}, {"联系电话", func(p *model.PartyMember) string { return p.Phone }},
		{"工作单位/务工地", func(p *model.PartyMember) string { return p.WorkPlace }},
	},
	Import: []ImportCol{{"姓名", "name"}, {"性别", "gender"}, {"出生日期", "birth"}, {"民族", "nation"}, {"文化程度", "education"},
		{"入党日期", "join_date"}, {"转正日期", "formal_date"}, {"党员类型", "type"}, {"流动党员", "mobile"}, {"党内职务", "post"},
		{"联系电话", "phone"}, {"工作单位/务工地", "work_place"}},
}

// ---------------- 民生保障 ----------------

var WelfareRes = &Resource[model.Welfare]{Title: "民生保障台账", Search: []string{"name", "guardian", "remark"},
	Filters: map[string]string{"type": "type", "status": "status", "grid_id": "grid_id"}, GridScope: true, DeleteRoles: adminCadre,
	Order: "type, id",
	Prepare: func(c *gin.Context, w *model.Welfare, old *model.Welfare) error {
		if w.VillagerID > 0 {
			var v model.Villager
			if db.DB.First(&v, w.VillagerID).Error != nil {
				return errors.New("所选村民不存在")
			}
			w.Name, w.IDCard, w.GridID = v.Name, v.IDCard, v.GridID
			if w.Phone == "" {
				w.Phone = v.Phone
			}
		}
		if strings.TrimSpace(w.Name) == "" {
			return errors.New("请选择或填写保障对象")
		}
		if w.Type == "" {
			return errors.New("请选择保障类型")
		}
		if w.GridID == 0 {
			return errors.New("请选择所属网格")
		}
		if w.Status == "" {
			w.Status = "在保"
		}
		if strings.Contains(w.IDCard, "*") && old != nil {
			w.IDCard = old.IDCard
		}
		return nil
	},
	AfterList: func(c *gin.Context, items []model.Welfare) {
		for i := range items {
			items[i].IDCard = util.MaskIDCard(items[i].IDCard)
		}
	},
	Export: []Col[model.Welfare]{
		{"姓名", func(w *model.Welfare) string { return w.Name }}, {"身份证号(脱敏)", func(w *model.Welfare) string { return w.IDCard }},
		{"网格", func(w *model.Welfare) string { return gridNameCached(w.GridID) }}, {"保障类型", func(w *model.Welfare) string { return w.Type }},
		{"类别/等级", func(w *model.Welfare) string { return w.Level }}, {"月保障金额(元)", func(w *model.Welfare) string { return F(w.Amount) }},
		{"纳入日期", func(w *model.Welfare) string { return w.StartDate }}, {"状态", func(w *model.Welfare) string { return w.Status }},
		{"帮扶责任人", func(w *model.Welfare) string { return w.Guardian }}, {"联系电话", func(w *model.Welfare) string { return w.Phone }},
		{"备注", func(w *model.Welfare) string { return w.Remark }},
	},
	Import: []ImportCol{{"姓名", "name"}, {"网格", "grid_name"}, {"保障类型", "type"}, {"类别/等级", "level"}, {"月保障金额(元)", "amount"},
		{"纳入日期", "start_date"}, {"状态", "status"}, {"帮扶责任人", "guardian"}, {"联系电话", "phone"}, {"备注", "remark"}},
}

func WelfareStats(c *gin.Context) {
	type row struct {
		Type string  `json:"type"`
		N    int     `json:"n"`
		Sum  float64 `json:"sum"`
	}
	var rows []row
	WelfareRes.query(c).Where("status = ?", "在保").Select("type, COUNT(*) AS n, COALESCE(SUM(amount),0) AS sum").Group("type").Scan(&rows)
	util.OK(c, rows)
}

// ---------------- 乡村治理 ----------------

var DisputeRes = &Resource[model.Dispute]{Title: "矛盾纠纷调解", Search: []string{"case_no", "title", "parties", "mediator"},
	Filters: map[string]string{"status": "status", "category": "category", "grid_id": "grid_id"}, DateCol: "accept_date",
	GridScope: true, DeleteRoles: adminCadre, Order: "accept_date DESC, id DESC",
	Prepare: func(c *gin.Context, d *model.Dispute, old *model.Dispute) error {
		if strings.TrimSpace(d.Title) == "" {
			return errors.New("请填写纠纷事由")
		}
		if d.GridID == 0 {
			return errors.New("请选择所属网格")
		}
		if d.Status == "" {
			d.Status = "待调解"
		}
		if d.AcceptDate == "" {
			d.AcceptDate = today()
		}
		if old == nil {
			d.CaseNo = fmt.Sprintf("TJ%s%03d", time.Now().Format("20060102"), time.Now().UnixNano()/1000%1000)
		} else {
			d.CaseNo = old.CaseNo
		}
		return nil
	},
	Export: []Col[model.Dispute]{
		{"案件编号", func(d *model.Dispute) string { return d.CaseNo }}, {"纠纷事由", func(d *model.Dispute) string { return d.Title }},
		{"类别", func(d *model.Dispute) string { return d.Category }}, {"当事人", func(d *model.Dispute) string { return d.Parties }},
		{"网格", func(d *model.Dispute) string { return gridNameCached(d.GridID) }}, {"受理日期", func(d *model.Dispute) string { return d.AcceptDate }},
		{"调解员", func(d *model.Dispute) string { return d.Mediator }}, {"状态", func(d *model.Dispute) string { return d.Status }},
		{"纠纷简况", func(d *model.Dispute) string { return d.Content }}, {"调解结果", func(d *model.Dispute) string { return d.Result }},
	},
	Import: []ImportCol{{"纠纷事由", "title"}, {"类别", "category"}, {"当事人", "parties"}, {"网格", "grid_name"}, {"受理日期", "accept_date"},
		{"调解员", "mediator"}, {"状态", "status"}, {"纠纷简况", "content"}, {"调解结果", "result"}},
}

var HonorRes = &Resource[model.Honor]{Title: "红黑榜", Search: []string{"target", "reason", "category"},
	Filters: map[string]string{"board": "board", "period": "period"}, WriteRoles: adminCadre, Order: "id DESC",
	ExtraWhere: func(c *gin.Context, q *gorm.DB) *gorm.DB {
		if mw.IsGrid(c) {
			q = q.Where("published = ?", true)
		}
		return q
	},
	Prepare: func(c *gin.Context, h *model.Honor, old *model.Honor) error {
		if h.Board != "红榜" && h.Board != "黑榜" {
			return errors.New("请选择红榜或黑榜")
		}
		if h.HouseholdID > 0 {
			var hh model.Household
			if db.DB.First(&hh, h.HouseholdID).Error == nil {
				if h.Target == "" {
					h.Target = hh.HeadName + "户"
				}
				h.GridID = hh.GridID
			}
		}
		if strings.TrimSpace(h.Target) == "" {
			return errors.New("请填写上榜对象")
		}
		if h.Period == "" {
			h.Period = time.Now().Format("2006年01月")
		}
		return nil
	},
	Export: []Col[model.Honor]{{"榜单", func(h *model.Honor) string { return h.Board }}, {"对象", func(h *model.Honor) string { return h.Target }},
		{"类别", func(h *model.Honor) string { return h.Category }}, {"事由", func(h *model.Honor) string { return h.Reason }},
		{"期次", func(h *model.Honor) string { return h.Period }}, {"已公示", func(h *model.Honor) string { return B(h.Published) }}},
}

var Dimensions = []string{"爱党爱国", "邻里和睦", "人居环境", "勤劳致富", "移风易俗", "遵纪守法"}

var PointRes = &Resource[model.PointRecord]{Title: "治理积分记录", Search: []string{"household_name", "reason", "recorder"},
	Filters: map[string]string{"dimension": "dimension", "household_id": "household_id", "grid_id": "grid_id"}, DateCol: "record_date",
	GridScope: true, DeleteRoles: adminCadre, Order: "record_date DESC, id DESC",
	Prepare: func(c *gin.Context, p *model.PointRecord, old *model.PointRecord) error {
		var h model.Household
		if p.HouseholdID == 0 || db.DB.First(&h, p.HouseholdID).Error != nil {
			return errors.New("请选择家庭户")
		}
		p.HouseholdName, p.GridID = h.HeadName, h.GridID
		if p.Points == 0 {
			return errors.New("请填写积分（加分为正，扣分为负）")
		}
		if p.Dimension == "" {
			return errors.New("请选择积分维度")
		}
		if p.RecordDate == "" {
			p.RecordDate = today()
		}
		if old == nil {
			p.Recorder = mw.CurrentUser(c).RealName
		}
		return nil
	},
	Export: []Col[model.PointRecord]{{"日期", func(p *model.PointRecord) string { return p.RecordDate }}, {"户主", func(p *model.PointRecord) string { return p.HouseholdName }},
		{"网格", func(p *model.PointRecord) string { return gridNameCached(p.GridID) }}, {"维度", func(p *model.PointRecord) string { return p.Dimension }},
		{"积分", func(p *model.PointRecord) string { return I(p.Points) }}, {"事由", func(p *model.PointRecord) string { return p.Reason }},
		{"记录人", func(p *model.PointRecord) string { return p.Recorder }}},
}

type rankRow struct {
	HouseholdID   uint   `json:"household_id"`
	HouseholdName string `json:"household_name"`
	GridID        uint   `json:"grid_id"`
	GridName      string `json:"grid_name"`
	Total         int    `json:"total"`
	Times         int    `json:"times"`
}

func pointRank(limit int) []rankRow {
	var rows []rankRow
	db.DB.Model(&model.PointRecord{}).Select("household_id, MAX(household_name) AS household_name, MAX(grid_id) AS grid_id, SUM(points) AS total, COUNT(*) AS times").
		Group("household_id").Order("total DESC").Limit(limit).Scan(&rows)
	for i := range rows {
		rows[i].GridName = gridNameCached(rows[i].GridID)
	}
	return rows
}

func PointRank(c *gin.Context) {
	type dim struct {
		Dimension string `json:"dimension"`
		Total     int    `json:"total"`
	}
	var dims []dim
	db.DB.Model(&model.PointRecord{}).Select("dimension, SUM(points) AS total").Group("dimension").Scan(&dims)
	util.OK(c, gin.H{"rank": pointRank(50), "dimensions": dims, "dimension_list": Dimensions})
}

// ---------------- 乡村振兴 ----------------

var RichLeaderRes = &Resource[model.RichLeader]{Title: "致富带头人", Search: []string{"name", "industry", "entity"},
	Filters: map[string]string{"industry": "industry"}, WriteRoles: adminCadre, Order: "drive DESC, id",
	Prepare: func(c *gin.Context, r *model.RichLeader, old *model.RichLeader) error {
		if strings.TrimSpace(r.Name) == "" {
			return errors.New("请填写姓名")
		}
		return nil
	},
	Export: []Col[model.RichLeader]{{"姓名", func(r *model.RichLeader) string { return r.Name }}, {"性别", func(r *model.RichLeader) string { return r.Gender }},
		{"产业类型", func(r *model.RichLeader) string { return r.Industry }}, {"经营主体", func(r *model.RichLeader) string { return r.Entity }},
		{"经营规模", func(r *model.RichLeader) string { return r.Scale }}, {"带动户数", func(r *model.RichLeader) string { return I(r.Drive) }},
		{"年产值(万元)", func(r *model.RichLeader) string { return F(r.Income) }}, {"党员", func(r *model.RichLeader) string { return B(r.IsParty) }},
		{"联系电话", func(r *model.RichLeader) string { return r.Phone }}},
	Import: []ImportCol{{"姓名", "name"}, {"性别", "gender"}, {"产业类型", "industry"}, {"经营主体", "entity"}, {"经营规模", "scale"},
		{"带动户数", "drive"}, {"年产值(万元)", "income"}, {"党员", "is_party"}, {"联系电话", "phone"}},
}

// FundStats 集体经济收入统计：按年度、按科目
func FundStats(c *gin.Context) {
	var fs []model.Fund
	db.DB.Find(&fs)
	years := map[string][2]float64{}
	cats := map[string]float64{}
	thisYear := time.Now().Format("2006")
	for _, f := range fs {
		if len(f.Date) < 4 {
			continue
		}
		y := years[f.Date[:4]]
		if f.Type == "收入" {
			y[0] += f.Amount
			if f.Date[:4] == thisYear {
				cats[f.Category] += f.Amount
			}
		} else {
			y[1] += f.Amount
		}
		years[f.Date[:4]] = y
	}
	ys := []string{}
	for y := range years {
		ys = append(ys, y)
	}
	sort.Strings(ys)
	yearRows := []gin.H{}
	for _, y := range ys {
		yearRows = append(yearRows, gin.H{"year": y, "income": years[y][0], "expense": years[y][1]})
	}
	catRows := []kv{}
	for k, v := range cats {
		catRows = append(catRows, kv{k, v})
	}
	sort.Slice(catRows, func(i, j int) bool { return catRows[i].Value > catRows[j].Value })
	var assetVal float64
	db.DB.Model(&model.Asset{}).Select("COALESCE(SUM(value),0)").Scan(&assetVal)
	util.OK(c, gin.H{"years": yearRows, "categories": catRows, "asset_value": assetVal})
}

// ---------------- 门户 ----------------

var PortalRes = &Resource[model.PortalItem]{Title: "门户内容", Search: []string{"title", "summary"},
	Filters: map[string]string{"kind": "kind"}, ReadRoles: adminCadre, WriteRoles: adminCadre, Order: "kind, sort, date DESC, id DESC",
	Prepare: func(c *gin.Context, p *model.PortalItem, old *model.PortalItem) error {
		if p.Kind != "banner" && p.Kind != "news" && p.Kind != "story" {
			return errors.New("内容类型不正确")
		}
		if strings.TrimSpace(p.Title) == "" {
			return errors.New("请填写标题")
		}
		if p.Date == "" {
			p.Date = today()
		}
		if p.Link != "" && !strings.HasPrefix(p.Link, "http://") && !strings.HasPrefix(p.Link, "https://") && !strings.HasPrefix(p.Link, "#/") {
			return errors.New("链接需以 http(s):// 或 #/ 开头")
		}
		return nil
	},
}

// PublicPortal 门户首页数据（无需登录，仅公开内容与汇总数字）
func PublicPortal(c *gin.Context) {
	items := []model.PortalItem{}
	db.DB.Where("published = ?", true).Order("sort, date DESC, id DESC").Find(&items)
	banners, news, stories := []model.PortalItem{}, []model.PortalItem{}, []model.PortalItem{}
	for _, it := range items {
		switch it.Kind {
		case "banner":
			banners = append(banners, it)
		case "news":
			news = append(news, it)
		case "story":
			stories = append(stories, it)
		}
	}
	notices := []model.Notice{}
	db.DB.Where("status = ?", "已发布").Order("top DESC, publish_date DESC, id DESC").Limit(12).Find(&notices)
	var v model.Village
	db.DB.First(&v)
	var hh, pp, party, grids, proj int64
	db.DB.Model(&model.Household{}).Count(&hh)
	db.DB.Model(&model.Villager{}).Count(&pp)
	db.DB.Model(&model.PartyMember{}).Count(&party)
	db.DB.Model(&model.Grid{}).Count(&grids)
	db.DB.Model(&model.Project{}).Count(&proj)
	var income float64
	db.DB.Model(&model.Fund{}).Where("type = ? AND date >= ?", "收入", time.Now().Format("2006")+"-01-01").Select("COALESCE(SUM(amount),0)").Scan(&income)
	honors := []model.Honor{}
	db.DB.Where("published = ?", true).Order("id DESC").Limit(10).Find(&honors)
	leaders := []model.RichLeader{}
	db.DB.Select("id, name, industry, entity, drive, story").Order("drive DESC").Limit(6).Find(&leaders)
	util.OK(c, gin.H{"banners": banners, "news": news, "stories": stories, "notices": notices, "honors": honors, "leaders": leaders,
		"village": gin.H{"name": v.Name, "town": v.Town, "intro": v.Intro, "area": v.Area, "farmland": v.Farmland, "address": v.Address, "phone": v.Phone},
		"stats": gin.H{"households": hh, "people": pp, "party": party, "grids": grids, "projects": proj, "income": income},
		"rank":  pointRank(10)})
}

func PublicNotice(c *gin.Context) {
	var n model.Notice
	if db.DB.Where("status = ?", "已发布").First(&n, util.ParamID(c)).Error != nil {
		util.NotFound(c)
		return
	}
	db.DB.Model(&n).UpdateColumn("views", gorm.Expr("views + 1"))
	util.OK(c, n)
}

func PublicItem(c *gin.Context) {
	var it model.PortalItem
	if db.DB.Where("published = ?", true).First(&it, util.ParamID(c)).Error != nil {
		util.NotFound(c)
		return
	}
	util.OK(c, it)
}

// ---------------- 数据驾驶舱 ----------------

func Screen(c *gin.Context) {
	out := gin.H{}
	var v model.Village
	db.DB.First(&v)
	out["village"] = v
	// 各网格户数/人数
	var grids []model.Grid
	db.DB.Order("sort, id").Find(&grids)
	type cnt struct {
		GridID uint
		N      float64
	}
	var hc, pc, ec []cnt
	db.DB.Model(&model.Household{}).Select("grid_id, COUNT(*) AS n").Group("grid_id").Scan(&hc)
	db.DB.Model(&model.Villager{}).Select("grid_id, COUNT(*) AS n").Group("grid_id").Scan(&pc)
	db.DB.Model(&model.Event{}).Select("grid_id, COUNT(*) AS n").Group("grid_id").Scan(&ec)
	m := func(cs []cnt) map[uint]float64 {
		r := map[uint]float64{}
		for _, x := range cs {
			r[x.GridID] = x.N
		}
		return r
	}
	hm, pm, em := m(hc), m(pc), m(ec)
	gridRows := []gin.H{}
	for _, g := range grids {
		gridRows = append(gridRows, gin.H{"name": g.Name, "worker": g.WorkerName, "households": hm[g.ID], "people": pm[g.ID], "events": em[g.ID]})
	}
	out["grids"] = gridRows
	// 近 30 天事件趋势
	start := time.Now().AddDate(0, 0, -29)
	var evs []model.Event
	db.DB.Select("created_at, status").Where("created_at >= ?", start.Format("2006-01-02")).Find(&evs)
	days, counts := []string{}, []int{}
	idx := map[string]int{}
	for i := 0; i < 30; i++ {
		d := start.AddDate(0, 0, i).Format("01-02")
		idx[d] = i
		days = append(days, d)
		counts = append(counts, 0)
	}
	for _, e := range evs {
		if i, ok := idx[e.CreatedAt.Format("01-02")]; ok {
			counts[i]++
		}
	}
	out["event_trend"] = gin.H{"days": days, "counts": counts}
	// 民生保障
	var ws []struct {
		Type string  `json:"name"`
		N    float64 `json:"value"`
	}
	db.DB.Model(&model.Welfare{}).Where("status = ?", "在保").Select("type, COUNT(*) AS n").Group("type").Scan(&ws)
	out["welfare"] = ws
	// 项目
	var ps []model.Project
	db.DB.Select("name, progress, budget, invested, status").Order("id").Find(&ps)
	out["projects"] = ps
	// 党员
	var pmN, pmMobile, pmPre int64
	db.DB.Model(&model.PartyMember{}).Count(&pmN)
	db.DB.Model(&model.PartyMember{}).Where("mobile = ?", true).Count(&pmMobile)
	db.DB.Model(&model.PartyMember{}).Where("type = ?", "预备党员").Count(&pmPre)
	out["party"] = gin.H{"total": pmN, "mobile": pmMobile, "probation": pmPre}
	var meetN int64
	db.DB.Model(&model.Meeting{}).Where("meet_date >= ?", time.Now().Format("2006")+"-01-01").Count(&meetN)
	out["meetings_year"] = meetN
	// 积分榜
	out["rank"] = pointRank(8)
	// 矛盾纠纷
	var ds []struct {
		Status string  `json:"name"`
		N      float64 `json:"value"`
	}
	db.DB.Model(&model.Dispute{}).Select("status, COUNT(*) AS n").Group("status").Scan(&ds)
	out["disputes"] = ds
	var latest []model.Event
	db.DB.Select("id, title, category, status, created_at, grid_id").Order("id DESC").Limit(10).Find(&latest)
	out["latest_events"] = latest
	var leaders int64
	db.DB.Model(&model.RichLeader{}).Count(&leaders)
	out["leaders"] = leaders
	util.OK(c, out)
}

// ---------------- 备份恢复 ----------------

type tabler interface{ TableName() string }

func tableNames() []string {
	names := []string{}
	for _, m := range model.All() {
		stmt := &gorm.Statement{DB: db.DB}
		_ = stmt.Parse(m)
		names = append(names, stmt.Schema.Table)
	}
	return names
}

func Backup(c *gin.Context) {
	data := map[string]any{"_meta": gin.H{"app": "village-mgmt", "version": "v0.1.0", "time": time.Now().Format(time.RFC3339)}}
	for _, t := range tableNames() {
		rows := []map[string]any{}
		if err := db.DB.Table(t).Find(&rows).Error; err != nil {
			util.ServerError(c, err)
			return
		}
		data[t] = rows
	}
	b, _ := json.Marshal(data)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"village-backup-%s.json\"", time.Now().Format("20060102150405")))
	c.Data(200, "application/json; charset=utf-8", b)
}

func Restore(c *gin.Context) {
	f, err := c.FormFile("file")
	if err != nil {
		util.BadRequest(c, "请选择备份文件")
		return
	}
	src, err := f.Open()
	if err != nil {
		util.ServerError(c, err)
		return
	}
	defer src.Close()
	b, _ := io.ReadAll(io.LimitReader(src, 200<<20))
	var data map[string][]map[string]any
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &raw); err != nil {
		util.BadRequest(c, "备份文件格式错误")
		return
	}
	if _, ok := raw["_meta"]; !ok {
		util.BadRequest(c, "不是本系统的备份文件")
		return
	}
	delete(raw, "_meta")
	data = map[string][]map[string]any{}
	for k, v := range raw {
		var rows []map[string]any
		if err := json.Unmarshal(v, &rows); err != nil {
			util.BadRequest(c, "备份数据表 "+k+" 格式错误")
			return
		}
		data[k] = rows
	}
	if len(data["users"]) == 0 {
		util.BadRequest(c, "备份中没有用户数据，已拒绝恢复")
		return
	}
	known := map[string]bool{}
	for _, t := range tableNames() {
		known[t] = true
	}
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for t := range known {
			rows, ok := data[t]
			if !ok {
				continue
			}
			if err := tx.Exec("DELETE FROM " + t).Error; err != nil {
				return err
			}
			for i := 0; i < len(rows); i += 200 {
				end := i + 200
				if end > len(rows) {
					end = len(rows)
				}
				batch := rows[i:end]
				for _, r := range batch {
					for k, v := range r {
						if s, ok := v.(string); ok && len(s) >= 20 && s[4] == '-' && s[10] == 'T' {
							if tm, err := time.Parse(time.RFC3339Nano, s); err == nil {
								r[k] = tm
							}
						}
					}
				}
				if err := tx.Table(t).Create(batch).Error; err != nil {
					return fmt.Errorf("恢复 %s 失败：%w", t, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		util.ServerError(c, err)
		return
	}
	util.OK(c, gin.H{"tables": len(data)})
}

var _ tabler = model.Group{}
