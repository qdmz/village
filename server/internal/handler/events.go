package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qdmz/village-mgmt/internal/db"
	mw "github.com/qdmz/village-mgmt/internal/middleware"
	"github.com/qdmz/village-mgmt/internal/model"
	"github.com/qdmz/village-mgmt/internal/util"
	"gorm.io/gorm"
)

// 事件流转：上报 → 受理 → 处理中 → 办结；上报/受理 可 退回；退回 可 重新上报

var EventRes = &Resource[model.Event]{Title: "事件台账", Search: []string{"event_no", "title", "location", "description", "reporter"},
	Filters: map[string]string{"status": "status", "category": "category", "grid_id": "grid_id", "urgency": "urgency", "handler_id": "handler_id"},
	GridScope: true, DeleteRoles: adminOnly,
	Order: "CASE status WHEN '上报' THEN 0 WHEN '受理' THEN 1 WHEN '处理中' THEN 2 WHEN '退回' THEN 3 ELSE 4 END, id DESC",
	ExtraWhere: func(c *gin.Context, q *gorm.DB) *gorm.DB {
		if v := c.Query("start"); v != "" {
			q = q.Where("created_at >= ?", v)
		}
		if v := c.Query("end"); v != "" {
			q = q.Where("created_at <= ?", v+" 23:59:59")
		}
		if c.Query("mine") == "1" {
			uid := mw.CurrentUser(c).ID
			q = q.Where("reporter_id = ? OR handler_id = ?", uid, uid)
		}
		return q
	},
	Prepare: func(c *gin.Context, e *model.Event, old *model.Event) error {
		u := mw.CurrentUser(c)
		if strings.TrimSpace(e.Title) == "" {
			return errors.New("请填写事件标题")
		}
		if e.Category == "" {
			return errors.New("请选择事件类别")
		}
		if e.GridID == 0 {
			return errors.New("请选择所属网格")
		}
		if e.Urgency == "" {
			e.Urgency = "一般"
		}
		if old == nil {
			e.Status = model.EvReported
			e.ReporterID, e.Reporter = u.ID, u.RealName
			e.HandlerID, e.Handler, e.Result, e.FinishedAt = 0, "", "", nil
			e.EventNo = fmt.Sprintf("SJ%s%04d", time.Now().Format("20060102"), time.Now().UnixNano()/1000%10000)
			return nil
		}
		if u.Role == model.RoleGrid && old.ReporterID != u.ID {
			return errors.New("只能修改自己上报的事件")
		}
		if old.Status != model.EvReported && old.Status != model.EvReturned && u.Role == model.RoleGrid {
			return errors.New("事件已受理，不能再修改")
		}
		// 状态与流转字段只能通过流转接口修改
		e.Status, e.ReporterID, e.Reporter = old.Status, old.ReporterID, old.Reporter
		e.HandlerID, e.Handler, e.Result, e.FinishedAt, e.EventNo = old.HandlerID, old.Handler, old.Result, old.FinishedAt, old.EventNo
		return nil
	},
	AfterSave: func(c *gin.Context, e *model.Event, created bool) {
		if created {
			u := mw.CurrentUser(c)
			db.DB.Create(&model.EventFlow{EventID: e.ID, Action: "上报", ToStatus: model.EvReported, OperatorID: u.ID, Operator: u.RealName, Comment: "网格事件上报"})
		}
	},
	AfterDelete: func(c *gin.Context, e *model.Event) { db.DB.Where("event_id = ?", e.ID).Delete(&model.EventFlow{}) },
	AfterGet: func(c *gin.Context, e *model.Event) any {
		flows := []model.EventFlow{}
		db.DB.Where("event_id = ?", e.ID).Order("id").Find(&flows)
		var g model.Grid
		db.DB.First(&g, e.GridID)
		return gin.H{"event": e, "flows": flows, "grid_name": g.Name, "actions": allowedActions(c, e)}
	},
	Export: []Col[model.Event]{
		{"事件编号", func(e *model.Event) string { return e.EventNo }},
		{"标题", func(e *model.Event) string { return e.Title }},
		{"类别", func(e *model.Event) string { return e.Category }},
		{"紧急程度", func(e *model.Event) string { return e.Urgency }},
		{"网格", func(e *model.Event) string { return gridNameCached(e.GridID) }},
		{"位置", func(e *model.Event) string { return e.Location }},
		{"描述", func(e *model.Event) string { return e.Description }},
		{"状态", func(e *model.Event) string { return e.Status }},
		{"上报人", func(e *model.Event) string { return e.Reporter }},
		{"上报时间", func(e *model.Event) string { return TS(e.CreatedAt) }},
		{"处理人", func(e *model.Event) string { return e.Handler }},
		{"处理结果", func(e *model.Event) string { return e.Result }},
	},
}

// allowedActions 当前用户对事件可执行的流转动作
func allowedActions(c *gin.Context, e *model.Event) []string {
	u := mw.CurrentUser(c)
	manager := u.Role == model.RoleAdmin || u.Role == model.RoleCadre
	acts := []string{}
	switch e.Status {
	case model.EvReported:
		if manager {
			acts = append(acts, "accept", "reject")
		}
	case model.EvAccepted:
		if manager {
			acts = append(acts, "process", "reject")
		}
	case model.EvProcessing:
		if manager || e.HandlerID == u.ID {
			acts = append(acts, "finish")
		}
		if manager {
			acts = append(acts, "process") // 改派
		}
	case model.EvReturned:
		if e.ReporterID == u.ID || manager {
			acts = append(acts, "resubmit")
		}
	}
	return acts
}

type flowReq struct {
	Action    string `json:"action"`
	Comment   string `json:"comment"`
	HandlerID uint   `json:"handler_id"`
	Result    string `json:"result"`
}

// EventFlowAction 事件流转
func EventFlowAction(c *gin.Context) {
	var e model.Event
	if err := db.DB.First(&e, util.ParamID(c)).Error; err != nil {
		util.NotFound(c)
		return
	}
	if !EventRes.canAccess(c, &e) && e.HandlerID != mw.CurrentUser(c).ID {
		util.Forbidden(c)
		return
	}
	var req flowReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, "参数格式错误")
		return
	}
	ok := false
	for _, a := range allowedActions(c, &e) {
		if a == req.Action {
			ok = true
		}
	}
	if !ok {
		util.BadRequest(c, "当前状态下不能执行该操作")
		return
	}
	u := mw.CurrentUser(c)
	from := e.Status
	label := ""
	switch req.Action {
	case "accept":
		e.Status, label = model.EvAccepted, "受理"
	case "reject":
		if strings.TrimSpace(req.Comment) == "" {
			util.BadRequest(c, "请填写退回原因")
			return
		}
		e.Status, label = model.EvReturned, "退回"
	case "process":
		if req.HandlerID == 0 {
			util.BadRequest(c, "请选择处理人")
			return
		}
		e.HandlerID, e.Handler = req.HandlerID, userName(req.HandlerID)
		e.Status, label = model.EvProcessing, "交办处理"
		if req.Comment == "" {
			req.Comment = "交由 " + e.Handler + " 处理"
		}
	case "finish":
		if strings.TrimSpace(req.Result) == "" {
			util.BadRequest(c, "请填写办理结果")
			return
		}
		now := time.Now()
		e.Status, e.Result, e.FinishedAt, label = model.EvDone, req.Result, &now, "办结"
		if req.Comment == "" {
			req.Comment = req.Result
		}
	case "resubmit":
		e.Status, label = model.EvReported, "重新上报"
	}
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&e).Error; err != nil {
			return err
		}
		return tx.Create(&model.EventFlow{EventID: e.ID, Action: label, FromStatus: from, ToStatus: e.Status,
			OperatorID: u.ID, Operator: u.RealName, Comment: req.Comment}).Error
	})
	if err != nil {
		util.ServerError(c, err)
		return
	}
	util.OK(c, e)
}

type progressReq struct {
	Progress int    `json:"progress"`
	Content  string `json:"content"`
}

// TaskProgress 责任人或干部更新任务进度
func TaskProgress(c *gin.Context) {
	var t model.Task
	if err := db.DB.First(&t, util.ParamID(c)).Error; err != nil {
		util.NotFound(c)
		return
	}
	u := mw.CurrentUser(c)
	if u.Role == model.RoleGrid && t.AssigneeID != u.ID {
		util.Forbidden(c)
		return
	}
	var req progressReq
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequest(c, "参数格式错误")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		util.BadRequest(c, "请填写进展情况")
		return
	}
	t.Progress = req.Progress
	if t.Progress < 100 && t.Status == "已完成" {
		t.Status = "进行中"
	}
	normalizeTask(&t)
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
		return tx.Create(&model.TaskUpdate{TaskID: t.ID, Progress: t.Progress, Content: req.Content, OperatorID: u.ID, Operator: u.RealName}).Error
	})
	if err != nil {
		util.ServerError(c, err)
		return
	}
	markOverdue(&t)
	util.OK(c, t)
}
