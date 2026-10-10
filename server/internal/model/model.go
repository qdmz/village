// Package model 定义全部数据表。
package model

import "time"

type Base struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---------- 系统 ----------

const (
	RoleAdmin = "admin" // 管理员
	RoleCadre = "cadre" // 村干部
	RoleGrid  = "grid"  // 网格员
)

type User struct {
	Base
	Username  string     `gorm:"size:50;uniqueIndex" json:"username"`
	Password  string     `gorm:"size:100" json:"-"`
	RealName  string     `gorm:"size:50" json:"real_name"`
	Role      string     `gorm:"size:20;index" json:"role"`
	Phone     string     `gorm:"size:20" json:"phone"`
	Position  string     `gorm:"size:50" json:"position"` // 职务
	Status    int        `gorm:"default:1" json:"status"` // 1 启用 0 停用
	LastLogin *time.Time `json:"last_login"`
}

type OpLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Username  string    `gorm:"size:50" json:"username"`
	RealName  string    `gorm:"size:50" json:"real_name"`
	Method    string    `gorm:"size:10" json:"method"`
	Path      string    `gorm:"size:200" json:"path"`
	Module    string    `gorm:"size:50" json:"module"`
	Action    string    `gorm:"size:50" json:"action"`
	IP        string    `gorm:"size:64" json:"ip"`
	Status    int       `json:"status"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

type Setting struct {
	Key   string `gorm:"primaryKey;size:64" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

// ---------- 村情与网格 ----------

type Village struct {
	Base
	Name      string  `gorm:"size:100" json:"name"`
	Code      string  `gorm:"size:50" json:"code"`
	Town      string  `gorm:"size:100" json:"town"` // 所属乡镇
	Address   string  `gorm:"size:200" json:"address"`
	Area      float64 `json:"area"`       // 面积(平方公里)
	Farmland  float64 `json:"farmland"`   // 耕地(亩)
	Secretary string  `gorm:"size:50" json:"secretary"` // 党支部书记
	Director  string  `gorm:"size:50" json:"director"`  // 村委会主任
	Phone     string  `gorm:"size:30" json:"phone"`
	Intro     string  `gorm:"type:text" json:"intro"`
}

type Group struct { // 村民小组
	Base
	VillageID uint   `gorm:"index" json:"village_id"`
	Name      string `gorm:"size:50" json:"name"`
	Leader    string `gorm:"size:50" json:"leader"`
	Phone     string `gorm:"size:30" json:"phone"`
	Sort      int    `json:"sort"`
}

type Grid struct {
	Base
	VillageID  uint   `gorm:"index" json:"village_id"`
	GroupID    uint   `gorm:"index" json:"group_id"`
	Name       string `gorm:"size:50" json:"name"`
	Code       string `gorm:"size:30" json:"code"`
	Scope      string `gorm:"size:255" json:"scope"` // 网格范围
	WorkerID   uint   `gorm:"index" json:"worker_id"` // 网格员（用户）
	WorkerName string `gorm:"size:50" json:"worker_name"`
	Sort       int    `json:"sort"`
}

// ---------- 户籍与人口 ----------

type Household struct {
	Base
	HouseholdNo string  `gorm:"size:30;index" json:"household_no"` // 户号
	HeadName    string  `gorm:"size:50;index" json:"head_name"`   // 户主
	Address     string  `gorm:"size:200" json:"address"`
	GroupID     uint    `gorm:"index" json:"group_id"`
	GridID      uint    `gorm:"index" json:"grid_id"`
	Type        string  `gorm:"size:20" json:"type"` // 户类型
	Phone       string  `gorm:"size:20" json:"phone"`
	MemberCount int     `json:"member_count"`
	Income      float64 `json:"income"` // 家庭年收入(元)
	Housing     string  `gorm:"size:50" json:"housing"` // 住房情况
	Remark      string  `gorm:"size:500" json:"remark"`
}

type Villager struct {
	Base
	HouseholdID uint   `gorm:"index" json:"household_id"`
	GridID      uint   `gorm:"index" json:"grid_id"`
	Name        string `gorm:"size:50;index" json:"name"`
	IDCard      string `gorm:"size:18" json:"id_card"`
	Gender      string `gorm:"size:4" json:"gender"`
	Birth       string `gorm:"size:10" json:"birth"`
	Relation    string `gorm:"size:20" json:"relation"` // 与户主关系
	Nation      string `gorm:"size:20" json:"nation"`
	Political   string `gorm:"size:20" json:"political"` // 政治面貌
	Education   string `gorm:"size:20" json:"education"`
	Employment  string `gorm:"size:20" json:"employment"` // 就业状态
	WorkPlace   string `gorm:"size:100" json:"work_place"`
	Phone       string `gorm:"size:20" json:"phone"`
	Tags        string `gorm:"size:200" json:"tags"` // 特殊人群标签，逗号分隔
	Remark      string `gorm:"size:500" json:"remark"`
}

// ---------- 网格巡查与事件 ----------

type Patrol struct {
	Base
	GridID     uint   `gorm:"index" json:"grid_id"`
	UserID     uint   `gorm:"index" json:"user_id"`
	UserName   string `gorm:"size:50" json:"user_name"`
	PatrolDate string `gorm:"size:10;index" json:"patrol_date"`
	StartTime  string `gorm:"size:5" json:"start_time"`
	EndTime    string `gorm:"size:5" json:"end_time"`
	Route      string `gorm:"size:255" json:"route"`
	Content    string `gorm:"type:text" json:"content"`
	Problems   string `gorm:"type:text" json:"problems"`
	Images     string `gorm:"type:text" json:"images"` // JSON 数组
}

const (
	EvReported   = "上报"
	EvAccepted   = "受理"
	EvProcessing = "处理中"
	EvDone       = "办结"
	EvReturned   = "退回"
)

type Event struct {
	Base
	EventNo     string     `gorm:"size:30;index" json:"event_no"`
	Title       string     `gorm:"size:100" json:"title"`
	Category    string     `gorm:"size:30;index" json:"category"`
	Urgency     string     `gorm:"size:10" json:"urgency"` // 一般/紧急/特急
	GridID      uint       `gorm:"index" json:"grid_id"`
	Location    string     `gorm:"size:200" json:"location"`
	Description string     `gorm:"type:text" json:"description"`
	Images      string     `gorm:"type:text" json:"images"`
	Status      string     `gorm:"size:10;index" json:"status"`
	ReporterID  uint       `gorm:"index" json:"reporter_id"`
	Reporter    string     `gorm:"size:50" json:"reporter"`
	HandlerID   uint       `gorm:"index" json:"handler_id"`
	Handler     string     `gorm:"size:50" json:"handler"`
	Result      string     `gorm:"type:text" json:"result"`
	FinishedAt  *time.Time `json:"finished_at"`
}

type EventFlow struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	EventID    uint      `gorm:"index" json:"event_id"`
	Action     string    `gorm:"size:20" json:"action"`
	FromStatus string    `gorm:"size:10" json:"from_status"`
	ToStatus   string    `gorm:"size:10" json:"to_status"`
	OperatorID uint      `json:"operator_id"`
	Operator   string    `gorm:"size:50" json:"operator"`
	Comment    string    `gorm:"size:500" json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

// ---------- 任务督办 ----------

type Task struct {
	Base
	Title        string     `gorm:"size:100" json:"title"`
	Content      string     `gorm:"type:text" json:"content"`
	Category     string     `gorm:"size:30" json:"category"`
	Priority     string     `gorm:"size:4" json:"priority"` // 高/中/低
	AssigneeID   uint       `gorm:"index" json:"assignee_id"`
	AssigneeName string     `gorm:"size:50" json:"assignee_name"`
	CreatorID    uint       `json:"creator_id"`
	CreatorName  string     `gorm:"size:50" json:"creator_name"`
	Deadline     string     `gorm:"size:10;index" json:"deadline"`
	Status       string     `gorm:"size:10;index" json:"status"` // 待办/进行中/已完成
	Progress     int        `json:"progress"`
	CompletedAt  *time.Time `json:"completed_at"`
	Overdue      bool       `gorm:"-" json:"overdue"`
}

type TaskUpdate struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TaskID     uint      `gorm:"index" json:"task_id"`
	Progress   int       `json:"progress"`
	Content    string    `gorm:"size:500" json:"content"`
	OperatorID uint      `json:"operator_id"`
	Operator   string    `gorm:"size:50" json:"operator"`
	CreatedAt  time.Time `json:"created_at"`
}

// ---------- 工作日志 / 民情日记 ----------

type WorkLog struct {
	Base
	UserID        uint   `gorm:"index" json:"user_id"`
	UserName      string `gorm:"size:50" json:"user_name"`
	LogDate       string `gorm:"size:10;index" json:"log_date"`
	Type          string `gorm:"size:20;index" json:"type"` // 工作日志/民情日记/走访记录
	Title         string `gorm:"size:100" json:"title"`
	Content       string `gorm:"type:text" json:"content"`
	HouseholdID   uint   `gorm:"index" json:"household_id"`
	HouseholdName string `gorm:"size:50" json:"household_name"`
	GridID        uint   `gorm:"index" json:"grid_id"`
	Demand        string `gorm:"size:500" json:"demand"` // 群众诉求
	Followup      string `gorm:"size:500" json:"followup"`
}

// ---------- 通知公告 / 村务公开 ----------

type Notice struct {
	Base
	Title       string `gorm:"size:150" json:"title"`
	Category    string `gorm:"size:20;index" json:"category"` // 通知公告/党务公开/村务公开/财务公开
	Content     string `gorm:"type:text" json:"content"`
	Publisher   string `gorm:"size:50" json:"publisher"`
	PublishDate string `gorm:"size:10;index" json:"publish_date"`
	Status      string `gorm:"size:10" json:"status"` // 已发布/草稿
	Top         bool   `json:"top"`
	Views       int    `json:"views"`
}

// ---------- 产业项目 / 资金 / 资产 ----------

type Project struct {
	Base
	Name        string  `gorm:"size:100" json:"name"`
	Type        string  `gorm:"size:20" json:"type"`
	Leader      string  `gorm:"size:50" json:"leader"`
	Location    string  `gorm:"size:200" json:"location"`
	Budget      float64 `json:"budget"`   // 万元
	Invested    float64 `json:"invested"` // 万元
	Progress    int     `json:"progress"`
	Status      string  `gorm:"size:10" json:"status"`
	StartDate   string  `gorm:"size:10" json:"start_date"`
	EndDate     string  `gorm:"size:10" json:"end_date"`
	Benefit     string  `gorm:"size:500" json:"benefit"` // 带动效益
	Description string  `gorm:"type:text" json:"description"`
}

type Fund struct {
	Base
	Date      string  `gorm:"size:10;index" json:"date"`
	Type      string  `gorm:"size:4;index" json:"type"` // 收入/支出
	Category  string  `gorm:"size:30" json:"category"`
	Amount    float64 `json:"amount"` // 元
	Summary   string  `gorm:"size:255" json:"summary"`
	VoucherNo string  `gorm:"size:30" json:"voucher_no"`
	Handler   string  `gorm:"size:50" json:"handler"`
	ProjectID uint    `json:"project_id"`
}

type Asset struct {
	Base
	Code       string  `gorm:"size:30" json:"code"`
	Name       string  `gorm:"size:100" json:"name"`
	Category   string  `gorm:"size:20" json:"category"` // 资源性/经营性/非经营性
	Quantity   float64 `json:"quantity"`
	Unit       string  `gorm:"size:10" json:"unit"`
	Value      float64 `json:"value"` // 万元
	Location   string  `gorm:"size:200" json:"location"`
	Status     string  `gorm:"size:10" json:"status"` // 在用/闲置/出租/承包/报废
	Keeper     string  `gorm:"size:50" json:"keeper"`
	AcquiredAt string  `gorm:"size:10" json:"acquired_at"`
	Remark     string  `gorm:"size:500" json:"remark"`
}

// ---------- 会议活动 ----------

type Meeting struct {
	Base
	Title         string `gorm:"size:150" json:"title"`
	Type          string `gorm:"size:20;index" json:"type"`
	MeetDate      string `gorm:"size:10;index" json:"meet_date"`
	Location      string `gorm:"size:100" json:"location"`
	Host          string `gorm:"size:50" json:"host"`
	Recorder      string `gorm:"size:50" json:"recorder"`
	Attendees     string `gorm:"type:text" json:"attendees"` // 参会人员，逗号分隔
	AttendCount   int    `json:"attend_count"`
	ShouldCount   int    `json:"should_count"`
	Absent        string `gorm:"size:255" json:"absent"`
	Topics        string `gorm:"type:text" json:"topics"`
	Content       string `gorm:"type:text" json:"content"`
	Resolution    string `gorm:"type:text" json:"resolution"`
}

func All() []any {
	return []any{&PartyOrg{}, &PartyMember{}, &Welfare{}, &Dispute{}, &Honor{}, &PointRecord{}, &RichLeader{}, &PortalItem{}, &User{}, &OpLog{}, &Setting{}, &Village{}, &Group{}, &Grid{}, &Household{}, &Villager{},
		&Patrol{}, &Event{}, &EventFlow{}, &Task{}, &TaskUpdate{}, &WorkLog{}, &Notice{}, &Project{}, &Fund{}, &Asset{}, &Meeting{}}
}

func (Group) TableName() string { return "village_groups" }

// ---------- 党建引领 ----------

type PartyOrg struct { // 党组织架构
	Base
	ParentID  uint   `gorm:"index" json:"parent_id"`
	Name      string `gorm:"size:100" json:"name"`
	Type      string `gorm:"size:20" json:"type"` // 党委/党总支/党支部/党小组
	Secretary string `gorm:"size:50" json:"secretary"`
	Members   string `gorm:"size:500" json:"members"` // 委员/成员
	Duty      string `gorm:"size:500" json:"duty"`
	Sort      int    `json:"sort"`
}

type PartyMember struct { // 党员花名册
	Base
	Name       string `gorm:"size:50;index" json:"name"`
	Gender     string `gorm:"size:4" json:"gender"`
	Birth      string `gorm:"size:10" json:"birth"`
	Nation     string `gorm:"size:20" json:"nation"`
	Education  string `gorm:"size:20" json:"education"`
	JoinDate   string `gorm:"size:10" json:"join_date"`  // 入党日期
	FormalDate string `gorm:"size:10" json:"formal_date"` // 转正日期
	Type       string `gorm:"size:20" json:"type"`        // 正式党员/预备党员
	Mobile     bool   `json:"mobile"`                     // 流动党员
	OrgID      uint   `gorm:"index" json:"org_id"`
	OrgName    string `gorm:"size:100" json:"org_name"`
	Post       string `gorm:"size:50" json:"post"` // 党内职务
	Phone      string `gorm:"size:20" json:"phone"`
	WorkPlace  string `gorm:"size:100" json:"work_place"`
	GridID     uint   `gorm:"index" json:"grid_id"`
	Remark     string `gorm:"size:500" json:"remark"`
}

// ---------- 民生保障 ----------

type Welfare struct {
	Base
	VillagerID uint    `gorm:"index" json:"villager_id"`
	Name       string  `gorm:"size:50;index" json:"name"`
	IDCard     string  `gorm:"size:18" json:"id_card"`
	GridID     uint    `gorm:"index" json:"grid_id"`
	Type       string  `gorm:"size:20;index" json:"type"` // 低保/特困供养/优抚对象/残疾人/留守老人/留守儿童/留守妇女/临时救助
	Level      string  `gorm:"size:30" json:"level"`      // 类别/等级
	Amount     float64 `json:"amount"`                     // 月保障金额(元)
	StartDate  string  `gorm:"size:10" json:"start_date"`
	Status     string  `gorm:"size:10" json:"status"` // 在保/已退出
	Guardian   string  `gorm:"size:50" json:"guardian"` // 监护/帮扶责任人
	Phone      string  `gorm:"size:20" json:"phone"`
	Remark     string  `gorm:"size:500" json:"remark"`
}

// ---------- 乡村治理 ----------

type Dispute struct { // 矛盾纠纷调解
	Base
	CaseNo     string `gorm:"size:30" json:"case_no"`
	Title      string `gorm:"size:100" json:"title"`
	Category   string `gorm:"size:20" json:"category"` // 邻里/婚姻家庭/土地山林/债务/其他
	Parties    string `gorm:"size:200" json:"parties"`
	GridID     uint   `gorm:"index" json:"grid_id"`
	AcceptDate string `gorm:"size:10" json:"accept_date"`
	Mediator   string `gorm:"size:50" json:"mediator"`
	Status     string `gorm:"size:10" json:"status"` // 待调解/调解中/调解成功/调解失败
	Content    string `gorm:"type:text" json:"content"`
	Result     string `gorm:"type:text" json:"result"`
}

type Honor struct { // 红黑榜
	Base
	Board       string `gorm:"size:4;index" json:"board"` // 红榜/黑榜
	Target      string `gorm:"size:50" json:"target"`     // 对象（户/个人）
	HouseholdID uint   `json:"household_id"`
	GridID      uint   `json:"grid_id"`
	Category    string `gorm:"size:30" json:"category"`
	Reason      string `gorm:"size:500" json:"reason"`
	Period      string `gorm:"size:20" json:"period"` // 期次，如 2026年10月
	Published   bool   `json:"published"`
}

type PointRecord struct { // 治理积分
	Base
	HouseholdID   uint   `gorm:"index" json:"household_id"`
	HouseholdName string `gorm:"size:50" json:"household_name"`
	GridID        uint   `gorm:"index" json:"grid_id"`
	Dimension     string `gorm:"size:20" json:"dimension"` // 爱党爱国/邻里和睦/人居环境/勤劳致富/移风易俗/遵纪守法
	Points        int    `json:"points"`
	Reason        string `gorm:"size:300" json:"reason"`
	RecordDate    string `gorm:"size:10;index" json:"record_date"`
	Recorder      string `gorm:"size:50" json:"recorder"`
}

// ---------- 乡村振兴 ----------

type RichLeader struct { // 致富带头人
	Base
	Name     string  `gorm:"size:50" json:"name"`
	Gender   string  `gorm:"size:4" json:"gender"`
	Phone    string  `gorm:"size:20" json:"phone"`
	Industry string  `gorm:"size:30" json:"industry"`
	Entity   string  `gorm:"size:100" json:"entity"` // 经营主体
	Scale    string  `gorm:"size:100" json:"scale"`
	Drive    int     `json:"drive"`  // 带动户数
	Income   float64 `json:"income"` // 年产值(万元)
	IsParty  bool    `json:"is_party"`
	Story    string  `gorm:"type:text" json:"story"`
}

// ---------- 门户 ----------

type PortalItem struct {
	Base
	Kind      string `gorm:"size:10;index" json:"kind"` // banner/news/story
	Title     string `gorm:"size:150" json:"title"`
	Summary   string `gorm:"size:500" json:"summary"`
	Content   string `gorm:"type:text" json:"content"`
	Image     string `gorm:"size:255" json:"image"`
	Color     string `gorm:"size:100" json:"color"` // banner 渐变色
	Link      string `gorm:"size:255" json:"link"`
	Date      string `gorm:"size:10" json:"date"`
	Sort      int    `json:"sort"`
	Published bool   `json:"published"`
}
