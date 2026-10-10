package handler

import (
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

type kv struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// Dashboard 工作台统计（网格员仅统计本网格）
func Dashboard(c *gin.Context) {
	u := mw.CurrentUser(c)
	grids := mw.MyGridIDs(c)
	scope := func(q *gorm.DB) *gorm.DB {
		if grids != nil {
			return q.Where("grid_id IN ?", grids)
		}
		return q
	}
	now := time.Now()
	td := now.Format("2006-01-02")
	var households, villagers, gridN, eventsOpen, eventsTotal, patrolsMonth int64
	scope(db.DB.Model(&model.Household{})).Count(&households)
	scope(db.DB.Model(&model.Villager{})).Count(&villagers)
	if grids != nil {
		gridN = int64(len(grids))
		if len(grids) == 1 && grids[0] == 0 {
			gridN = 0
		}
	} else {
		db.DB.Model(&model.Grid{}).Count(&gridN)
	}
	scope(db.DB.Model(&model.Event{})).Count(&eventsTotal)
	scope(db.DB.Model(&model.Event{})).Where("status <> ?", model.EvDone).Count(&eventsOpen)
	scope(db.DB.Model(&model.Patrol{})).Where("patrol_date >= ?", now.Format("2006-01")+"-01").Count(&patrolsMonth)

	taskQ := func() *gorm.DB {
		q := db.DB.Model(&model.Task{})
		if u.Role == model.RoleGrid {
			q = q.Where("assignee_id = ?", u.ID)
		}
		return q
	}
	var tasksTotal, tasksDone, tasksOverdue int64
	taskQ().Count(&tasksTotal)
	taskQ().Where("status = ?", "已完成").Count(&tasksDone)
	taskQ().Where("status <> ? AND deadline < ?", "已完成", td).Count(&tasksOverdue)

	// 人口结构
	var vs []model.Villager
	scope(db.DB.Model(&model.Villager{})).Select("gender, birth, tags, political, employment").Find(&vs)
	ageBands := []string{"0-17岁", "18-35岁", "36-59岁", "60-79岁", "80岁以上"}
	ages := map[string]float64{}
	gender := map[string]float64{}
	tags := map[string]float64{}
	emp := map[string]float64{}
	party := 0.0
	for _, v := range vs {
		gender[v.Gender]++
		if len(v.Birth) >= 4 {
			y, _ := strconv.Atoi(v.Birth[:4])
			a := now.Year() - y
			switch {
			case a < 18:
				ages[ageBands[0]]++
			case a < 36:
				ages[ageBands[1]]++
			case a < 60:
				ages[ageBands[2]]++
			case a < 80:
				ages[ageBands[3]]++
			default:
				ages[ageBands[4]]++
			}
		}
		for _, t := range strings.Split(v.Tags, ",") {
			if t != "" {
				tags[t]++
			}
		}
		if v.Employment != "" {
			emp[v.Employment]++
		}
		if v.Political == "中共党员" || v.Political == "中共预备党员" {
			party++
		}
	}
	toKV := func(m map[string]float64, order []string) []kv {
		out := []kv{}
		for _, k := range order {
			out = append(out, kv{k, m[k]})
		}
		return out
	}
	tagOrder := []string{"低保", "五保", "残疾", "脱贫监测户", "留守老人", "留守儿童", "退役军人"}
	empOrder := []string{"在家务农", "本地务工", "外出务工", "个体经营", "在校学生", "学龄前", "退休/养老", "无业"}

	// 事件状态
	type sc struct {
		Status string
		N      float64
	}
	var esc []sc
	scope(db.DB.Model(&model.Event{})).Select("status, COUNT(*) AS n").Group("status").Scan(&esc)
	em := map[string]float64{}
	for _, x := range esc {
		em[x.Status] = x.N
	}
	var ecat []struct {
		Category string
		N        float64
	}
	scope(db.DB.Model(&model.Event{})).Select("category, COUNT(*) AS n").Group("category").Order("n DESC").Scan(&ecat)
	catKV := []kv{}
	for _, x := range ecat {
		catKV = append(catKV, kv{x.Category, x.N})
	}

	// 任务状态
	var tsc []sc
	taskQ().Select("status, COUNT(*) AS n").Group("status").Scan(&tsc)
	tm := map[string]float64{}
	for _, x := range tsc {
		tm[x.Status] = x.N
	}
	rate := 0.0
	if tasksTotal > 0 {
		rate = float64(tasksDone) * 100 / float64(tasksTotal)
	}

	var pm int64
	db.DB.Model(&model.PartyMember{}).Count(&pm)
	if pm > 0 && grids == nil {
		party = float64(pm)
	}
	out := gin.H{
		"cards": gin.H{"households": households, "villagers": villagers, "grids": gridN, "events_open": eventsOpen,
			"events_total": eventsTotal, "tasks_total": tasksTotal, "tasks_done": tasksDone, "tasks_overdue": tasksOverdue,
			"task_rate": rate, "patrols_month": patrolsMonth, "party_members": party},
		"population": gin.H{"gender": toKV(gender, []string{"男", "女"}), "age": toKV(ages, ageBands), "tags": toKV(tags, tagOrder),
			"employment": toKV(emp, empOrder)},
		"events":   gin.H{"status": toKV(em, []string{model.EvReported, model.EvAccepted, model.EvProcessing, model.EvDone, model.EvReturned}), "category": catKV},
		"tasks":    gin.H{"status": toKV(tm, []string{"待办", "进行中", "已完成"}), "overdue": tasksOverdue, "rate": rate},
		"role":     u.Role,
	}

	// 资金收支（近 12 个月，干部/管理员可见）
	if u.Role != model.RoleGrid {
		start := time.Date(now.Year(), now.Month()-11, 1, 0, 0, 0, 0, now.Location())
		var fs []model.Fund
		db.DB.Where("date >= ?", start.Format("2006-01-02")).Find(&fs)
		months := []string{}
		inM, outM := map[string]float64{}, map[string]float64{}
		for i := 0; i < 12; i++ {
			months = append(months, start.AddDate(0, i, 0).Format("2006-01"))
		}
		for _, f := range fs {
			if len(f.Date) < 7 {
				continue
			}
			if f.Type == "收入" {
				inM[f.Date[:7]] += f.Amount
			} else {
				outM[f.Date[:7]] += f.Amount
			}
		}
		ins, outs := []float64{}, []float64{}
		for _, m := range months {
			ins = append(ins, inM[m]/10000)
			outs = append(outs, outM[m]/10000)
		}
		var totIn, totOut float64
		db.DB.Model(&model.Fund{}).Where("type = ?", "收入").Select("COALESCE(SUM(amount),0)").Scan(&totIn)
		db.DB.Model(&model.Fund{}).Where("type = ?", "支出").Select("COALESCE(SUM(amount),0)").Scan(&totOut)
		var projN int64
		db.DB.Model(&model.Project{}).Count(&projN)
		out["funds"] = gin.H{"months": months, "income": ins, "expense": outs, "total_income": totIn, "total_expense": totOut, "balance": totIn - totOut}
		out["cards"].(gin.H)["projects"] = projN
	}

	// 待办
	todos := []gin.H{}
	var myTasks []model.Task
	db.DB.Where("assignee_id = ? AND status <> ?", u.ID, "已完成").Order("deadline").Limit(8).Find(&myTasks)
	for _, t := range myTasks {
		todos = append(todos, gin.H{"type": "task", "id": t.ID, "title": t.Title, "deadline": t.Deadline, "priority": t.Priority,
			"overdue": t.Deadline < td, "status": t.Status})
	}
	var evs []model.Event
	eq := db.DB.Model(&model.Event{})
	switch u.Role {
	case model.RoleGrid:
		eq = eq.Where("(handler_id = ? AND status = ?) OR (reporter_id = ? AND status = ?)", u.ID, model.EvProcessing, u.ID, model.EvReturned)
	default:
		eq = eq.Where("status IN ? OR (handler_id = ? AND status = ?)", []string{model.EvReported, model.EvAccepted}, u.ID, model.EvProcessing)
	}
	eq.Order("id DESC").Limit(8).Find(&evs)
	for _, e := range evs {
		todos = append(todos, gin.H{"type": "event", "id": e.ID, "title": e.Title, "status": e.Status, "urgency": e.Urgency,
			"created_at": e.CreatedAt})
	}
	out["todos"] = todos

	var notices []model.Notice
	db.DB.Select("id, title, category, publish_date, top").Where("status = ?", "已发布").Order("top DESC, publish_date DESC, id DESC").Limit(6).Find(&notices)
	out["notices"] = notices

	var recent []model.Event
	scope(db.DB.Model(&model.Event{})).Order("id DESC").Limit(5).Find(&recent)
	out["recent_events"] = recent
	util.OK(c, out)
}
