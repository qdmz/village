// Package seed 首次启动时写入管理员账号与演示数据（虚构的“青山村”）。
package seed

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/qdmz/village-mgmt/internal/config"
	"github.com/qdmz/village-mgmt/internal/db"
	"github.com/qdmz/village-mgmt/internal/model"
	"github.com/qdmz/village-mgmt/internal/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func hash(p string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b)
}

func Run() error {
	var n int64
	db.DB.Model(&model.User{}).Count(&n)
	if n > 0 {
		return nil
	}
	pw := config.C.Seed.AdminPassword
	if pw == "" {
		pw = util.RandPassword(14)
		log.Printf("【重要】已创建初始管理员 admin，随机密码：%s （请登录后立即修改）", pw)
	} else {
		log.Printf("已创建初始管理员 admin（密码来自配置文件 seed.admin_password）")
	}
	admin := model.User{Username: "admin", Password: hash(pw), RealName: "系统管理员", Role: model.RoleAdmin, Position: "系统管理员", Status: 1}
	if err := db.DB.Create(&admin).Error; err != nil {
		return err
	}
	if !config.C.Seed.Demo {
		return db.DB.Create(&model.Village{Name: "示例村"}).Error
	}
	log.Printf("首次启动，正在写入演示数据……")
	return db.DB.Transaction(func(tx *gorm.DB) error { return demo(tx) })
}

type sd struct {
	tx  *gorm.DB
	r   *rand.Rand
	now time.Time
	err error
}

func (s *sd) create(v any) {
	if s.err == nil {
		s.err = s.tx.Create(v).Error
	}
}

func (s *sd) pick(list ...string) string { return list[s.r.Intn(len(list))] }
func (s *sd) day(offset int) string      { return s.now.AddDate(0, 0, offset).Format("2006-01-02") }
func (s *sd) at(offset int, hour int) time.Time {
	d := s.now.AddDate(0, 0, offset)
	t := time.Date(d.Year(), d.Month(), d.Day(), hour, s.r.Intn(60), 0, 0, d.Location())
	if t.After(s.now) {
		t = s.now.Add(-time.Duration(10+s.r.Intn(120)) * time.Minute)
	}
	return t
}

func (s *sd) phone() string {
	return fmt.Sprintf("1%s%08d", s.pick("39", "58", "37", "86", "52", "77", "35", "89"), s.r.Intn(100000000))
}

// idCard 生成带合法校验位的虚构身份证号
func idCard(birth string, male bool, r *rand.Rand) string {
	seq := r.Intn(50)*2 + 1
	if !male {
		seq--
	}
	body := fmt.Sprintf("431124%s%02d%d", strings.ReplaceAll(birth, "-", ""), r.Intn(100), seq%10)
	w := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	sum := 0
	for i, ch := range body {
		sum += int(ch-'0') * w[i]
	}
	return body + string("10X98765432"[sum%11])
}

func demo(tx *gorm.DB) error {
	s := &sd{tx: tx, r: rand.New(rand.NewSource(20261010)), now: time.Now()}

	// ---------- 用户 ----------
	mk := func(u, name, role, pos string) *model.User {
		x := &model.User{Username: u, Password: hash(util.RandPassword(16)), RealName: name, Role: role, Position: pos, Phone: s.phone(), Status: 1}
		s.create(x)
		return x
	}
	secretary := mk("zhangjg", "张建国", model.RoleCadre, "党支部书记、村委会主任")
	liuxm := mk("liuxm", "刘秀梅", model.RoleCadre, "村委委员、妇女主任")
	chenzq := mk("chenzq", "陈志强", model.RoleCadre, "村报账员、文书")
	wanglh := mk("wanglh", "王丽华", model.RoleGrid, "专职网格员")
	lics := mk("lics", "李春生", model.RoleGrid, "兼职网格员")
	zhaoxy := mk("zhaoxy", "赵小燕", model.RoleGrid, "兼职网格员")
	zhoudm := mk("zhoudm", "周德明", model.RoleGrid, "兼职网格员")
	cadres := []*model.User{secretary, liuxm, chenzq}

	// ---------- 村情 ----------
	v := &model.Village{Name: "青山村", Code: "431124108201", Town: "云溪镇", Address: "云溪镇青山村村民委员会（青山大道18号）",
		Area: 12.6, Farmland: 3260, Secretary: "张建国", Director: "张建国", Phone: "0746-7821XXX",
		Intro: "青山村位于云溪镇东北部，距镇政府 6 公里，下辖 5 个村民小组，以茶叶、稻虾共作和乡村旅游为主导产业。" +
			"近年来依托“党建+网格”治理模式，推进人居环境整治和集体经济发展，先后获评县级美丽乡村示范村、市级文明村。"}
	s.create(v)
	groupNames := []string{"第一村民小组", "第二村民小组", "第三村民小组", "第四村民小组", "第五村民小组"}
	places := []string{"上屋场", "塘湾", "枫树坪", "茶园冲", "河背"}
	leaders := []string{"黄德胜", "杨春华", "吴建平", "张国强", "李桂兰"}
	workers := []*model.User{wanglh, wanglh, lics, zhaoxy, zhoudm}
	var grids []*model.Grid
	for i, gn := range groupNames {
		g := &model.Group{VillageID: v.ID, Name: gn, Leader: leaders[i], Phone: s.phone(), Sort: i + 1}
		s.create(g)
		grid := &model.Grid{VillageID: v.ID, GroupID: g.ID, Name: fmt.Sprintf("第%s网格（%s）", []string{"一", "二", "三", "四", "五"}[i], places[i]),
			Code: fmt.Sprintf("QS-WG%02d", i+1), Scope: fmt.Sprintf("%s片区，东至%s，西至%s", places[i], s.pick("青山大道", "村部广场", "茶厂", "老石桥"), s.pick("云溪河", "后山林地", "县道X021", "水库坝")),
			WorkerID: workers[i].ID, WorkerName: workers[i].RealName, Sort: i + 1}
		s.create(grid)
		grids = append(grids, grid)
	}

	// ---------- 户籍与人口 ----------
	surnames := []string{"张", "张", "李", "王", "刘", "陈", "杨", "黄", "周", "吴", "赵", "罗", "谢"}
	maleGiven := []string{"建国", "国平", "德华", "志军", "春生", "海林", "立新", "永红", "文杰", "明辉", "长青", "胜利", "伟", "强", "军", "斌", "浩然", "子轩", "宇航", "俊杰", "铁柱", "满仓", "和平", "卫东"}
	femaleGiven := []string{"秀英", "桂兰", "玉梅", "淑芬", "春花", "丽", "芳", "红霞", "小燕", "美玲", "雨涵", "梓涵", "欣怡", "思琪", "翠萍", "冬梅", "秋菊", "海燕"}
	workPlaces := []string{"广东省东莞市", "广东省深圳市", "浙江省义乌市", "湖南省长沙市", "广东省佛山市", "江苏省苏州市", "本县工业园", "云溪镇"}
	var households []*model.Household
	var villagers []*model.Villager
	types := map[int]string{3: "低保户", 9: "脱贫户", 14: "低保户", 17: "五保户", 21: "脱贫户", 26: "监测户", 30: "脱贫户", 33: "低保户", 36: "监测户", 39: "脱贫户"}
	year := s.now.Year()
	birth := func(age int) string {
		return fmt.Sprintf("%04d-%02d-%02d", year-age, s.r.Intn(12)+1, s.r.Intn(28)+1)
	}
	education := func(age int) string {
		switch {
		case age < 7:
			return "学龄前"
		case age < 13:
			return "小学在读"
		case age < 16:
			return "初中在读"
		case age < 19:
			return s.pick("高中在读", "中专在读")
		case age >= 70:
			return s.pick("文盲或半文盲", "小学", "小学")
		case age >= 50:
			return s.pick("小学", "初中", "初中")
		case age >= 30:
			return s.pick("初中", "高中", "中专", "初中")
		default:
			return s.pick("高中", "大专", "本科", "中专")
		}
	}
	for i := 0; i < 40; i++ {
		grid := grids[i%5]
		htype := types[i]
		if htype == "" {
			htype = "一般户"
		}
		sur := surnames[s.r.Intn(len(surnames))]
		headAge := 38 + s.r.Intn(40)
		headMale := s.r.Intn(10) < 8
		if htype == "五保户" {
			headAge = 78
			headMale = true
		}
		headName := sur + maleGiven[s.r.Intn(14)]
		if !headMale {
			headName = sur + femaleGiven[s.r.Intn(10)]
		}
		h := &model.Household{HouseholdNo: fmt.Sprintf("QS%02d%03d", i%5+1, i+1), HeadName: headName,
			Address: fmt.Sprintf("青山村%s%s%d号", groupNames[i%5], places[i%5], 1+s.r.Intn(60)),
			GridID: grid.ID, GroupID: grid.GroupID, Type: htype, Phone: s.phone(),
			Income: float64(18000+s.r.Intn(90000)) / 1, Housing: s.pick("砖混结构自建房", "砖混结构自建房", "框架结构楼房", "土木结构老屋（已鉴定C级）")}
		if htype == "低保户" || htype == "五保户" || htype == "监测户" {
			h.Income = float64(6000 + s.r.Intn(9000))
		}
		s.create(h)
		households = append(households, h)

		type mem struct {
			name, rel string
			age       int
			male      bool
		}
		members := []mem{{headName, "户主", headAge, headMale}}
		if htype != "五保户" {
			spouseMale := !headMale
			sn := surnames[s.r.Intn(len(surnames))] + femaleGiven[s.r.Intn(10)]
			if spouseMale {
				sn = surnames[s.r.Intn(len(surnames))] + maleGiven[s.r.Intn(14)]
			}
			if s.r.Intn(10) < 8 {
				members = append(members, mem{sn, "配偶", headAge - 2 + s.r.Intn(5), spouseMale})
			}
			if headAge < 70 && s.r.Intn(10) < 8 {
				ca := headAge - 24 - s.r.Intn(6)
				if ca > 0 {
					cm := s.r.Intn(2) == 0
					cn := sur + maleGiven[14+s.r.Intn(10)]
					if !cm {
						cn = sur + femaleGiven[10+s.r.Intn(8)]
					}
					rel := "长子"
					if !cm {
						rel = "长女"
					}
					members = append(members, mem{cn, rel, ca, cm})
					if ca > 25 && s.r.Intn(10) < 6 {
						gm := s.r.Intn(2) == 0
						gn := sur + maleGiven[18+s.r.Intn(6)]
						rel := "孙子"
						if !gm {
							gn, rel = sur+femaleGiven[10+s.r.Intn(8)], "孙女"
						}
						members = append(members, mem{gn, rel, 3 + s.r.Intn(12), gm})
					}
				}
			}
			if headAge < 55 && s.r.Intn(10) < 3 {
				members = append(members, mem{sur + maleGiven[s.r.Intn(14)], "父亲", headAge + 26 + s.r.Intn(5), true})
			}
		}
		parentsAway := false
		var hm []*model.Villager
		for _, m := range members {
			b := birth(m.age)
			g := "女"
			if m.male {
				g = "男"
			}
			x := &model.Villager{HouseholdID: h.ID, GridID: grid.ID, Name: m.name, IDCard: idCard(b, m.male, s.r), Gender: g, Birth: b,
				Relation: m.rel, Nation: "汉族", Education: education(m.age), Political: "群众"}
			if s.r.Intn(30) == 0 {
				x.Nation = s.pick("土家族", "苗族", "瑶族")
			}
			switch {
			case m.age < 7:
				x.Employment = "学龄前"
			case m.age < 19:
				x.Employment = "在校学生"
				if m.age >= 14 && s.r.Intn(2) == 0 {
					x.Political = "共青团员"
				}
			case m.age < 60:
				x.Employment = s.pick("外出务工", "外出务工", "外出务工", "本地务工", "在家务农", "个体经营")
				if x.Employment == "外出务工" {
					x.WorkPlace = workPlaces[s.r.Intn(6)]
					if m.rel == "长子" || m.rel == "长女" {
						parentsAway = true
					}
				} else if x.Employment == "本地务工" {
					x.WorkPlace = workPlaces[6+s.r.Intn(2)]
				}
				if m.age < 29 && s.r.Intn(3) == 0 {
					x.Political = "共青团员"
				}
				x.Phone = s.phone()
			default:
				x.Employment = s.pick("在家务农", "退休/养老", "在家务农")
				if m.rel == "户主" {
					x.Phone = s.phone()
				}
			}
			if m.age >= 30 && s.r.Intn(9) == 0 {
				x.Political = "中共党员"
			}
			tags := []string{}
			switch htype {
			case "低保户":
				tags = append(tags, "低保")
			case "五保户":
				tags = append(tags, "五保")
			case "监测户":
				tags = append(tags, "脱贫监测户")
			}
			if s.r.Intn(18) == 0 && m.age > 20 {
				tags = append(tags, "残疾")
			}
			if m.male && m.age >= 35 && m.age <= 70 && s.r.Intn(12) == 0 {
				tags = append(tags, "退役军人")
				if x.Political == "群众" {
					x.Political = "中共党员"
				}
			}
			x.Tags = strings.Join(tags, ",")
			if m.rel == "户主" && h.Phone != "" {
				x.Phone = h.Phone
			}
			hm = append(hm, x)
		}
		for _, x := range hm {
			age := year - atoi(x.Birth[:4])
			if parentsAway && (x.Relation == "户主" || x.Relation == "配偶" || x.Relation == "父亲") && age >= 60 {
				x.Tags = addTag(x.Tags, "留守老人")
			}
			if parentsAway && (x.Relation == "孙子" || x.Relation == "孙女") && age < 16 {
				x.Tags = addTag(x.Tags, "留守儿童")
			}
			if htype == "五保户" {
				x.Tags = addTag(x.Tags, "留守老人")
			}
			s.create(x)
			villagers = append(villagers, x)
		}
		h.MemberCount = len(hm)
		if s.err == nil {
			s.err = tx.Model(h).Update("member_count", h.MemberCount).Error
		}
	}

	// ---------- 网格巡查 ----------
	patrolContents := []string{
		"沿主干道及房前屋后巡查，环境卫生整体良好，垃圾收集点已清运。",
		"走访独居老人 3 户，身体状况稳定；检查村道护栏及路灯，运行正常。",
		"巡查山塘水库及河道，水位正常，警示牌完好，未发现野泳人员。",
		"巡查林区周边，开展森林防火宣传，劝导村民不在林区焚烧秸秆。",
		"检查出租房及经营户消防设施，灭火器均在有效期内。",
		"入户宣传医保缴费和防范电信诈骗，发放宣传资料 20 余份。",
	}
	problems := []string{"", "", "", "塘湾路段 2 盏路灯不亮，已登记上报。", "发现 1 处生活垃圾乱堆放，已督促清理。", "河背组一处排水沟堵塞。"}
	for i := 0; i < 26; i++ {
		g := grids[i%5]
		uid, uname := g.WorkerID, g.WorkerName
		p := &model.Patrol{GridID: g.ID, UserID: uid, UserName: uname, PatrolDate: s.day(-i * 30 / 26), StartTime: "08:30", EndTime: s.pick("10:30", "11:00", "11:30"),
			Route: fmt.Sprintf("%s→村道→%s", strings.Split(strings.Split(g.Name, "（")[1], "）")[0], s.pick("村部", "茶园", "河堤", "学校路口")),
			Content: patrolContents[s.r.Intn(len(patrolContents))], Problems: problems[s.r.Intn(len(problems))], Images: "[]"}
		s.create(p)
	}

	// ---------- 事件 ----------
	type ev struct {
		title, cat, urg, loc, desc, status string
		grid, days                         int
	}
	evs := []ev{
		{"塘湾组路灯损坏", "基础设施", "一般", "塘湾组主干道 3 号电杆附近", "连续两晚有 2 盏路灯不亮，村民夜间出行不便。", model.EvDone, 1, -24},
		{"邻里宅基地边界纠纷", "矛盾纠纷", "紧急", "上屋场 23 号与 25 号之间", "两户因围墙修建位置发生争执，情绪较激动，需调解。", model.EvDone, 0, -20},
		{"河背组排水沟堵塞", "环境卫生", "一般", "河背组 8 号屋后排水沟", "排水沟被落叶和生活垃圾堵塞，雨天积水。", model.EvDone, 4, -18},
		{"独居老人房屋漏雨", "民生诉求", "紧急", "茶园冲 12 号", "五保户老人房屋屋顶漏雨，需修缮。", model.EvProcessing, 3, -9},
		{"林区周边焚烧秸秆", "森林防火", "特急", "枫树坪后山林区边缘", "发现有村民在林区 50 米内焚烧秸秆，已当场劝离并扑灭。", model.EvDone, 2, -15},
		{"村道边坡出现裂缝", "安全隐患", "紧急", "县道X021 青山段 K3+200", "连续降雨后边坡出现约 5 米裂缝，存在滑坡风险。", model.EvProcessing, 2, -6},
		{"垃圾收集点未及时清运", "环境卫生", "一般", "上屋场垃圾收集点", "垃圾桶已满溢，保洁公司 3 天未清运。", model.EvAccepted, 0, -3},
		{"申请临时救助", "民生诉求", "一般", "塘湾组 31 号", "村民因病住院产生大额医疗费用，申请临时救助。", model.EvAccepted, 1, -4},
		{"山塘护栏缺失", "安全隐患", "紧急", "河背组小山塘", "山塘东侧护栏缺失约 10 米，附近有儿童玩耍。", model.EvReported, 4, -1},
		{"自来水管爆裂", "基础设施", "紧急", "茶园冲 5 号门前", "主水管爆裂，周边 6 户停水。", model.EvReported, 3, 0},
		{"流浪犬伤人隐患", "安全隐患", "一般", "村小学门口", "学校周边有流浪犬出没，家长反映存在安全隐患。", model.EvReported, 0, 0},
		{"土地流转租金纠纷", "矛盾纠纷", "一般", "枫树坪组", "3 户村民反映流转租金未按时支付。", model.EvReturned, 2, -5},
		{"广场健身器材损坏", "基础设施", "一般", "村部文化广场", "2 台健身器材螺丝松动，存在安全隐患。", model.EvDone, 0, -28},
		{"汛期河道水位上涨", "防汛抗旱", "特急", "云溪河青山段", "上游降雨导致河道水位快速上涨，需转移低洼处群众。", model.EvDone, 4, -40},
		{"疑似电信诈骗求助", "社会治安", "紧急", "塘湾组 15 号", "老人接到冒充公检法电话，要求转账，网格员及时劝阻。", model.EvDone, 1, -12},
		{"违规搭建棚房", "其他", "一般", "茶园冲路口", "有村民在路口私自搭建棚房，影响通行。", model.EvProcessing, 3, -8},
	}
	for i, e := range evs {
		g := grids[e.grid]
		created := s.at(e.days, 9+s.r.Intn(8))
		x := &model.Event{EventNo: fmt.Sprintf("SJ%s%04d", created.Format("20060102"), 101+i), Title: e.title, Category: e.cat, Urgency: e.urg,
			GridID: g.ID, Location: e.loc, Description: e.desc, Images: "[]", Status: e.status, ReporterID: g.WorkerID, Reporter: g.WorkerName}
		x.CreatedAt, x.UpdatedAt = created, created
		handler := cadres[s.r.Intn(len(cadres))]
		if e.status == model.EvProcessing || e.status == model.EvDone {
			x.HandlerID, x.Handler = handler.ID, handler.RealName
		}
		if e.status == model.EvDone {
			ft := created.Add(time.Duration(20+s.r.Intn(60)) * time.Hour)
			x.FinishedAt = &ft
			x.Result = s.pick("已联系相关单位处理完毕，经回访群众满意。", "经村调委会调解，双方达成一致并签订协议。", "已组织人员现场处置，隐患消除。")
		}
		s.create(x)
		flows := []model.EventFlow{{EventID: x.ID, Action: "上报", ToStatus: model.EvReported, OperatorID: g.WorkerID, Operator: g.WorkerName, Comment: "网格事件上报", CreatedAt: created}}
		t := created
		next := func(action, from, to string, op *model.User, cm string) {
			t = t.Add(time.Duration(1+s.r.Intn(6)) * time.Hour)
			flows = append(flows, model.EventFlow{EventID: x.ID, Action: action, FromStatus: from, ToStatus: to, OperatorID: op.ID, Operator: op.RealName, Comment: cm, CreatedAt: t})
		}
		switch e.status {
		case model.EvAccepted:
			next("受理", model.EvReported, model.EvAccepted, secretary, "已受理，安排处理")
		case model.EvProcessing:
			next("受理", model.EvReported, model.EvAccepted, secretary, "已受理")
			next("交办处理", model.EvAccepted, model.EvProcessing, secretary, "交由 "+handler.RealName+" 处理")
		case model.EvDone:
			next("受理", model.EvReported, model.EvAccepted, secretary, "已受理")
			next("交办处理", model.EvAccepted, model.EvProcessing, secretary, "交由 "+handler.RealName+" 处理")
			flows = append(flows, model.EventFlow{EventID: x.ID, Action: "办结", FromStatus: model.EvProcessing, ToStatus: model.EvDone, OperatorID: handler.ID, Operator: handler.RealName, Comment: x.Result, CreatedAt: *x.FinishedAt})
		case model.EvReturned:
			next("退回", model.EvReported, model.EvReturned, secretary, "请补充涉及农户名单及流转合同信息后重新上报")
		}
		for i := range flows {
			s.create(&flows[i])
		}
	}

	// ---------- 任务督办 ----------
	type tk struct {
		title, cat, pri, content string
		who                      *model.User
		dl, prog                 int
		status                   string
	}
	tks := []tk{
		{"2027 年度城乡居民医保征缴", "民生保障", "高", "完成本网格居民医保征缴，确保参保率不低于 98%。", wanglh, 20, 65, "进行中"},
		{"秋冬季森林防火宣传入户", "森林防火", "高", "每户发放防火告知书，重点人员签订承诺书。", lics, 7, 40, "进行中"},
		{"防返贫监测户季度走访", "防返贫监测", "高", "对监测户、脱贫户开展季度走访，更新收入情况。", zhaoxy, -3, 70, "进行中"},
		{"人居环境整治“月评比”", "人居环境", "中", "组织各组开展庭院“美丽屋场”评比并公示。", liuxm, 12, 20, "进行中"},
		{"村道边坡隐患整改跟踪", "安全生产", "高", "协调交通站对 K3+200 边坡进行加固，每日巡查。", secretary, 5, 30, "进行中"},
		{"老年人高龄补贴资格核查", "民生保障", "中", "核查 80 周岁以上老人名单，补齐资料。", zhoudm, -6, 50, "进行中"},
		{"村集体经济年度台账整理", "集体经济", "中", "整理 1-9 月资金收支凭证并装订。", chenzq, 15, 0, "待办"},
		{"党员冬训动员", "党建工作", "中", "通知全体党员参加冬训，统计流动党员参训方式。", secretary, 30, 0, "待办"},
		{"夏季防溺水排查", "安全生产", "高", "排查全村山塘水库警示牌、救生设施。", wanglh, -40, 100, "已完成"},
		{"退役军人信息采集更新", "优抚安置", "低", "核实退役军人联系方式与家庭情况。", zhoudm, -10, 100, "已完成"},
		{"留守儿童关爱活动", "关爱服务", "中", "联合学校开展暑期留守儿童关爱活动。", liuxm, -50, 100, "已完成"},
		{"农房安全隐患排查", "安全生产", "高", "对 C/D 级农房逐户排查并建立台账。", lics, -2, 85, "进行中"},
	}
	for _, t := range tks {
		x := &model.Task{Title: t.title, Content: t.content, Category: t.cat, Priority: t.pri, AssigneeID: t.who.ID, AssigneeName: t.who.RealName,
			CreatorID: secretary.ID, CreatorName: secretary.RealName, Deadline: s.day(t.dl), Status: t.status, Progress: t.prog}
		x.CreatedAt = s.at(t.dl-25, 9)
		if t.status == "已完成" {
			ct := s.at(t.dl-2, 17)
			x.CompletedAt = &ct
		}
		s.create(x)
		if t.prog > 0 {
			steps := []int{t.prog}
			if t.prog > 40 {
				steps = []int{t.prog / 2, t.prog}
			}
			for k, p := range steps {
				u := &model.TaskUpdate{TaskID: x.ID, Progress: p, Content: s.pick("已完成部分入户，剩余农户本周内完成。", "已召开小组会议部署，正在逐户推进。", "已完成台账整理，待上报审核。", "按计划推进中。"),
					OperatorID: t.who.ID, Operator: t.who.RealName, CreatedAt: s.at(t.dl-20+k*8, 16)}
				if p == 100 {
					u.Content = "任务已全部完成，资料已归档。"
				}
				s.create(u)
			}
		}
	}

	// ---------- 工作日志 / 民情日记 ----------
	visitDemands := []string{"希望协助办理慢性病门诊卡。", "反映屋后山体有落石，希望加装防护网。", "咨询小额信贷政策，想扩大养殖规模。", "希望村里帮忙联系外出务工岗位。", "无特别诉求，身体状况良好。"}
	for i := 0; i < 22; i++ {
		var u *model.User
		var x *model.WorkLog
		if i%3 == 0 {
			u = cadres[s.r.Intn(3)]
			x = &model.WorkLog{UserID: u.ID, UserName: u.RealName, LogDate: s.day(-i), Type: "工作日志",
				Title: s.pick("村两委例会及工作部署", "接待来访群众", "参加镇里乡村振兴推进会", "整理防返贫监测资料"),
				Content: s.pick("上午召开村两委例会，部署本周重点工作；下午到茶叶基地查看项目进度。", "接待来访群众 4 人次，涉及低保咨询、宅基地审批等事项，已逐一答复。", "参加镇乡村振兴推进会，传达会议精神并落实到各组。")}
		} else {
			h := households[s.r.Intn(len(households))]
			var g *model.Grid
			for _, gg := range grids {
				if gg.ID == h.GridID {
					g = gg
				}
			}
			uid := g.WorkerID
			u = map[uint]*model.User{wanglh.ID: wanglh, lics.ID: lics, zhaoxy.ID: zhaoxy, zhoudm.ID: zhoudm}[uid]
			typ := "走访记录"
			if i%4 == 1 {
				typ = "民情日记"
			}
			x = &model.WorkLog{UserID: u.ID, UserName: u.RealName, LogDate: s.day(-i), Type: typ, Title: "走访" + h.HeadName + "家",
				Content: fmt.Sprintf("入户走访%s家（%s），了解家庭生产生活情况，宣传医保征缴及防诈骗知识。", h.HeadName, h.Type),
				HouseholdID: h.ID, HouseholdName: h.HeadName, GridID: h.GridID, Demand: visitDemands[s.r.Intn(len(visitDemands))],
				Followup: s.pick("已记录并向村委反映。", "已当场答复政策。", "下周回访。")}
		}
		s.create(x)
	}

	// ---------- 通知公告 / 村务公开 ----------
	notices := []model.Notice{
		{Title: "关于开展 2027 年度城乡居民医保集中征缴的通知", Category: "通知公告", Top: true, PublishDate: s.day(-2),
			Content: "各村民小组、全体村民：\n\n2027 年度城乡居民基本医疗保险集中征缴工作已经开始，个人缴费标准按县医保局最新公布执行。\n\n缴费方式：\n1. 通过“湘税社保”等线上渠道自助缴费；\n2. 到村便民服务中心现场缴费；\n3. 由本组网格员上门协助缴费。\n\n特困人员、低保对象等按政策享受资助参保。请大家相互转告，按时缴费。\n\n青山村村民委员会"},
		{Title: "秋冬季森林防火期禁止野外用火的通告", Category: "通知公告", Top: true, PublishDate: s.day(-5),
			Content: "即日起进入秋冬季森林防火期。严禁在林区及林缘 100 米范围内烧田埂草、焚烧秸秆、燃放烟花爆竹、野外吸烟等一切野外用火行为。\n\n发现火情请立即拨打 12119 或联系村委会值班电话。"},
		{Title: "第三季度村级财务收支公开", Category: "财务公开", PublishDate: s.day(-8),
			Content: "一、本季度收入合计 38.62 万元：上级转移支付 18.00 万元、集体经营收入（茶叶基地分红、门面租赁）14.12 万元、其他收入 6.50 万元。\n\n二、本季度支出合计 29.47 万元：基础设施维护 12.30 万元、人居环境整治 6.85 万元、村干部报酬 7.20 万元、办公费用 3.12 万元。\n\n三、截至本季度末，村集体账面资金余额详见资金台账。\n\n如有疑问，请向村务监督委员会反映。"},
		{Title: "关于发展预备党员的公示", Category: "党务公开", PublishDate: s.day(-12),
			Content: "经支部委员会研究，拟将杨文杰同志列为预备党员发展对象，现予以公示。公示期 5 个工作日。\n\n如有意见，请以书面或口头形式向村党支部或镇党委组织办反映。"},
		{Title: "“四议两公开”：村部至枫树坪道路硬化项目决议公开", Category: "村务公开", PublishDate: s.day(-16),
			Content: "经村党支部提议、村两委会商议、党员大会审议、村民代表会议决议，同意实施村部至枫树坪 1.8 公里道路硬化项目，预算 96 万元。\n\n施工单位通过公开招标确定，接受全体村民监督。"},
		{Title: "第三季度党费收缴情况公开", Category: "党务公开", PublishDate: s.day(-20),
			Content: "本季度应缴党费党员 26 名，实缴 26 名，收缴党费合计 1,560 元，已全额上缴镇党委。"},
		{Title: "低保对象动态调整结果公示", Category: "村务公开", PublishDate: s.day(-26),
			Content: "根据县民政局要求，经入户调查、民主评议，本季度新增低保对象 1 户 2 人，退出 1 户 1 人。名单详见村务公开栏。\n\n公示期 7 天，监督电话：村务监督委员会。"},
		{Title: "关于开展“美丽屋场”评比活动的通知", Category: "通知公告", PublishDate: s.day(-30),
			Content: "为持续推进人居环境整治，村委会决定在全村开展“美丽屋场”月度评比，评比结果在村部公示并给予奖励。"},
		{Title: "8 月份财务收支明细公开", Category: "财务公开", PublishDate: s.day(-45),
			Content: "8 月份收入 12.40 万元，支出 9.86 万元，明细已张贴于村务公开栏，欢迎村民监督。"},
		{Title: "冬季用电安全提醒（草稿）", Category: "通知公告", Status: "草稿", PublishDate: s.day(0),
			Content: "入冬后取暖用电增多，请村民注意电热毯、取暖器使用安全……"},
	}
	for i := range notices {
		n := &notices[i]
		n.Publisher = s.pick("张建国", "刘秀梅", "陈志强")
		if n.Status == "" {
			n.Status = "已发布"
		}
		n.Views = 20 + s.r.Intn(260)
		s.create(n)
	}

	// ---------- 产业项目 ----------
	projects := []*model.Project{
		{Name: "高山生态茶叶基地扩建", Type: "种植业", Leader: "张建国", Location: "茶园冲后山", Budget: 180, Invested: 132, Progress: 72, Status: "建设中",
			StartDate: s.day(-300), EndDate: s.day(120), Benefit: "新增茶园 260 亩，带动 46 户农户务工增收，户均年增收约 6000 元。",
			Description: "采取“村集体+合作社+农户”模式，新建标准化茶园并配套建设加工车间。"},
		{Name: "稻虾共作综合种养示范片", Type: "养殖业", Leader: "吴建平", Location: "塘湾组、河背组", Budget: 85, Invested: 85, Progress: 100, Status: "运营中",
			StartDate: s.day(-620), EndDate: s.day(-250), Benefit: "示范面积 420 亩，亩均增收 1800 元，村集体年分红 8 万元。", Description: "流转低洼稻田发展稻虾共作。"},
		{Name: "青山民宿与乡村旅游接待中心", Type: "乡村旅游", Leader: "刘秀梅", Location: "上屋场老屋群", Budget: 260, Invested: 64, Progress: 25, Status: "建设中",
			StartDate: s.day(-90), EndDate: s.day(300), Benefit: "预计年接待游客 2 万人次，提供 30 个就业岗位。", Description: "盘活闲置老屋 8 栋，改造为特色民宿。"},
		{Name: "村级光伏电站", Type: "新能源", Leader: "陈志强", Location: "村部及小学屋顶", Budget: 60, Invested: 60, Progress: 100, Status: "运营中",
			StartDate: s.day(-900), EndDate: s.day(-780), Benefit: "装机 200kW，年发电收益约 9 万元，纳入村集体收入。"},
		{Name: "村部至枫树坪道路硬化", Type: "基础设施", Leader: "张国强", Location: "村部—枫树坪", Budget: 96, Invested: 0, Progress: 0, Status: "规划中",
			StartDate: s.day(30), EndDate: s.day(150), Benefit: "解决 3 个组 180 余人出行难问题。"},
	}
	for _, p := range projects {
		s.create(p)
	}

	// ---------- 资金台账 ----------
	voucher := 1
	for m := 11; m >= 0; m-- {
		base := time.Date(s.now.Year(), s.now.Month()-time.Month(m), 1, 0, 0, 0, 0, s.now.Location())
		date := func(d int) string {
			t := base.AddDate(0, 0, d)
			if t.After(s.now) {
				t = s.now
			}
			return t.Format("2006-01-02")
		}
		add := func(typ, cat string, amt float64, sum string, d int, pid uint) {
			s.create(&model.Fund{Date: date(d), Type: typ, Category: cat, Amount: amt, Summary: sum, VoucherNo: fmt.Sprintf("记-%s-%03d", base.Format("0601"), voucher),
				Handler: "陈志强", ProjectID: pid})
			voucher++
		}
		add("收入", "经营收入", float64(6000+s.r.Intn(4000)), "光伏电站发电收益", 5, projects[3].ID)
		add("支出", "管理费用", float64(5800+s.r.Intn(800)), "村干部报酬及办公费用", 10, 0)
		if m%3 == 0 {
			add("收入", "补助收入", float64(50000+s.r.Intn(30000)), "上级转移支付（村级运转经费）", 3, 0)
			add("收入", "发包及上交收入", float64(12000+s.r.Intn(6000)), "门面及鱼塘租赁收入", 15, 0)
		}
		if m%2 == 0 {
			add("支出", "公益设施支出", float64(8000+s.r.Intn(20000)), s.pick("村道维护", "路灯维修更换", "垃圾清运服务费", "山塘护栏修缮"), 18, 0)
		}
		if m < 10 && m%4 == 1 {
			amt := float64(100000 + s.r.Intn(200000))
			add("收入", "补助收入", amt+30000, "上级衔接资金（茶叶基地项目）", 8, projects[0].ID)
			add("支出", "项目支出", amt, "高山生态茶叶基地建设款", 20, projects[0].ID)
		}
		if m == 6 {
			add("收入", "投资收益", 80000, "稻虾共作示范片年度分红", 25, projects[1].ID)
		}
		if m == 2 {
			add("收入", "补助收入", 300000, "乡村旅游项目上级奖补资金", 5, projects[2].ID)
			add("支出", "项目支出", 240000, "民宿改造首期工程款", 12, projects[2].ID)
		}
	}

	// ---------- 集体资产 ----------
	assets := []model.Asset{
		{Name: "村部综合服务楼", Category: "非经营性", Quantity: 1, Unit: "栋", Value: 165, Location: "青山大道 18 号", Status: "在用", Keeper: "陈志强", AcquiredAt: "2016-08-20"},
		{Name: "村级光伏电站", Category: "经营性", Quantity: 200, Unit: "kW", Value: 60, Location: "村部及小学屋顶", Status: "在用", Keeper: "陈志强", AcquiredAt: s.day(-780)},
		{Name: "临街门面", Category: "经营性", Quantity: 6, Unit: "间", Value: 72, Location: "青山大道 20-30 号", Status: "出租", Keeper: "张建国", AcquiredAt: "2019-03-15"},
		{Name: "集体山林", Category: "资源性", Quantity: 3850, Unit: "亩", Value: 0, Location: "后山、枫树坪", Status: "在用", Keeper: "张国强", AcquiredAt: "1983-01-01", Remark: "其中生态公益林 2100 亩"},
		{Name: "集体鱼塘", Category: "资源性", Quantity: 46, Unit: "亩", Value: 0, Location: "塘湾组", Status: "承包", Keeper: "杨春华", AcquiredAt: "1983-01-01", Remark: "承包期至 2030 年"},
		{Name: "茶叶加工车间及设备", Category: "经营性", Quantity: 1, Unit: "套", Value: 48, Location: "茶园冲", Status: "在用", Keeper: "吴建平", AcquiredAt: s.day(-200)},
		{Name: "文化广场及健身器材", Category: "非经营性", Quantity: 1, Unit: "处", Value: 26, Location: "村部前", Status: "在用", Keeper: "刘秀梅", AcquiredAt: "2020-10-01"},
		{Name: "小型挖掘机", Category: "经营性", Quantity: 1, Unit: "台", Value: 18, Location: "村部车库", Status: "闲置", Keeper: "张国强", AcquiredAt: "2021-05-12"},
		{Name: "原村小学旧校舍", Category: "非经营性", Quantity: 1, Unit: "栋", Value: 30, Location: "枫树坪", Status: "闲置", Keeper: "陈志强", AcquiredAt: "1995-09-01", Remark: "拟改造为老年活动中心"},
		{Name: "办公电脑及打印设备", Category: "非经营性", Quantity: 8, Unit: "台", Value: 4.6, Location: "村部", Status: "在用", Keeper: "陈志强", AcquiredAt: "2023-02-10"},
	}
	for i := range assets {
		assets[i].Code = fmt.Sprintf("QS-ZC-%03d", i+1)
		s.create(&assets[i])
	}

	// ---------- 会议活动 ----------
	party := "张建国,刘秀梅,陈志强,黄德胜,吴建平,杨文杰,王丽华,李春生,张国强,周德明,罗志军,谢长青"
	reps := "张建国,刘秀梅,陈志强,黄德胜,杨春华,吴建平,张国强,李桂兰,王丽华,李春生,赵小燕,周德明,罗海林,谢美玲,黄立新,刘永红"
	meetings := []model.Meeting{
		{Title: "10 月支部党员大会（主题党日）", Type: "支部党员大会", MeetDate: s.day(-6), Location: "村部党员活动室", Host: "张建国", Recorder: "陈志强",
			Attendees: party, ShouldCount: 26, Absent: "外出务工流动党员 14 人（线上参学）", Topics: "1. 学习上级会议精神；2. 讨论发展预备党员；3. 开展志愿服务",
			Content: "会议集中学习了有关文件精神，讨论了发展对象情况，并组织党员到茶园冲开展环境卫生志愿服务。", Resolution: "同意将杨文杰同志列为预备党员发展对象，报镇党委审批。"},
		{Title: "支委会：研究秋冬季重点工作", Type: "支委会", MeetDate: s.day(-10), Location: "村部会议室", Host: "张建国", Recorder: "陈志强",
			Attendees: "张建国,刘秀梅,陈志强", ShouldCount: 3, Topics: "医保征缴、森林防火、防返贫监测", Content: "逐项研究部署秋冬季重点工作。", Resolution: "明确各项工作责任人和完成时限，纳入任务督办。"},
		{Title: "第三季度村民代表大会", Type: "村民代表大会", MeetDate: s.day(-17), Location: "村部大会议室", Host: "张建国", Recorder: "刘秀梅",
			Attendees: reps, ShouldCount: 18, Absent: "罗志军、谢长青（请假）", Topics: "1. 村部至枫树坪道路硬化项目；2. 第三季度财务公开",
			Content: "代表们对道路硬化项目进行了充分讨论。", Resolution: "16 名代表全票同意实施道路硬化项目。"},
		{Title: "党小组会：网格党员联户", Type: "党小组会", MeetDate: s.day(-25), Location: "塘湾组活动点", Host: "黄德胜", Recorder: "王丽华",
			Attendees: "黄德胜,王丽华,李春生,杨文杰", ShouldCount: 5, Topics: "落实党员联户包片", Content: "明确每名党员联系 5-8 户群众。", Resolution: "建立党员联户台账。"},
		{Title: "党课：传承红色基因 建设美丽乡村", Type: "党课", MeetDate: s.day(-35), Location: "村部党员活动室", Host: "张建国", Recorder: "陈志强",
			Attendees: party, ShouldCount: 26, Topics: "书记讲党课", Content: "张建国同志结合本村发展历程讲授专题党课。"},
		{Title: "人居环境整治工作推进会", Type: "工作会议", MeetDate: s.day(-40), Location: "村部会议室", Host: "刘秀梅", Recorder: "赵小燕",
			Attendees: "张建国,刘秀梅,陈志强,王丽华,李春生,赵小燕,周德明,黄德胜,杨春华,吴建平,张国强,李桂兰", ShouldCount: 12, Topics: "美丽屋场评比方案", Content: "讨论评比细则。", Resolution: "每月评比一次，结果公示。"},
		{Title: "8 月支部党员大会", Type: "支部党员大会", MeetDate: s.day(-66), Location: "村部党员活动室", Host: "张建国", Recorder: "陈志强",
			Attendees: party, ShouldCount: 26, Topics: "防汛救灾总结", Content: "总结汛期工作，表扬先进党员。"},
		{Title: "第二季度村民代表大会", Type: "村民代表大会", MeetDate: s.day(-108), Location: "村部大会议室", Host: "张建国", Recorder: "刘秀梅",
			Attendees: reps, ShouldCount: 18, Topics: "民宿项目实施方案", Content: "审议民宿项目方案。", Resolution: "同意实施，资金使用接受村务监督委员会监督。"},
	}
	for i := range meetings {
		m := &meetings[i]
		m.AttendCount = len(strings.Split(m.Attendees, ","))
		s.create(m)
	}

	s.extra(households, villagers, grids, []*model.User{secretary, liuxm, chenzq, wanglh, lics, zhaoxy, zhoudm})

	if s.err != nil {
		return s.err
	}
	log.Printf("演示数据写入完成：%d 户、%d 人", len(households), len(villagers))
	return nil
}

func atoi(s string) int {
	n := 0
	fmt.Sscanf(s, "%d", &n)
	return n
}

func addTag(tags, t string) string {
	if tags == "" {
		return t
	}
	if strings.Contains(tags, t) {
		return tags
	}
	return tags + "," + t
}
