package seed

import (
	"fmt"
	"strings"

	"github.com/qdmz/village-mgmt/internal/model"
)

// extra 党建、民生保障、乡村治理、致富带头人、门户内容等演示数据
func (s *sd) extra(hs []*model.Household, vs []*model.Villager, grids []*model.Grid, us []*model.User) {
	secretary, liuxm, chenzq := us[0], us[1], us[2]
	// 党组织架构
	branch := &model.PartyOrg{Name: "中共青山村支部委员会", Type: "党支部", Secretary: secretary.RealName,
		Members: "张建国（书记）、刘秀梅（组织委员）、陈志强（宣传委员）", Duty: "全面领导村级各类组织和各项工作", Sort: 1}
	s.create(branch)
	var groups []*model.PartyOrg
	for i, n := range []string{"第一党小组（上屋场·塘湾）", "第二党小组（枫树坪·茶园冲）", "第三党小组（河背）", "流动党员党小组"} {
		g := &model.PartyOrg{ParentID: branch.ID, Name: n, Type: "党小组", Secretary: []string{"黄德胜", "吴建平", "张国强", "罗志军"}[i], Sort: i + 1,
			Duty: []string{"联系第一、二网格群众", "联系第三、四网格群众，服务茶叶产业", "联系第五网格群众", "外出务工党员线上学习与管理"}[i]}
		s.create(g)
		groups = append(groups, g)
	}
	s.create(&model.PartyOrg{ParentID: branch.ID, Name: "村务监督委员会", Type: "村级组织", Secretary: "黄立新", Sort: 9, Duty: "监督村务公开、财务管理和重大事项决策"})
	s.create(&model.PartyOrg{ParentID: branch.ID, Name: "青山村股份经济合作社", Type: "村级组织", Secretary: secretary.RealName, Sort: 10, Duty: "管理集体资产，发展集体经济"})

	// 党员花名册
	partyNames := []struct {
		name, post string
		male       bool
	}{{"张建国", "党支部书记", true}, {"刘秀梅", "组织委员", false}, {"陈志强", "宣传委员", true}, {"黄德胜", "党小组长", true},
		{"吴建平", "党小组长", true}, {"张国强", "党小组长", true}, {"王丽华", "", false}, {"李春生", "", true}, {"周德明", "", true},
		{"杨文杰", "", true}, {"罗志军", "流动党小组长", true}, {"谢长青", "", true}, {"黄立新", "监委会主任", true}, {"李桂兰", "", false},
		{"杨春华", "", false}, {"罗海林", "", true}, {"吴秀英", "", false}, {"刘永红", "", true}, {"张卫东", "", true}, {"陈美玲", "", false},
		{"赵和平", "", true}, {"周长青", "", true}, {"黄玉梅", "", false}, {"杨立新", "", true}, {"谢海燕", "", false}, {"罗胜利", "", true}}
	for i, p := range partyNames {
		age := 28 + s.r.Intn(45)
		g := "女"
		if p.male {
			g = "男"
		}
		join := fmt.Sprintf("%04d-%02d-01", s.now.Year()-age+22+s.r.Intn(8), 1+s.r.Intn(12))
		if join > s.day(0) {
			join = s.day(-400)
		}
		typ := "正式党员"
		formal := ""
		if i == 9 || i == 24 {
			typ = "预备党员"
			join = s.day(-120)
		} else {
			formal = join[:4] + "-12-01"
		}
		org := groups[i%3]
		mobile := i >= 18 && i%2 == 0 || i == 10 || i == 11
		wp := ""
		if mobile {
			org = groups[3]
			wp = s.pick("广东省东莞市", "浙江省义乌市", "广东省深圳市")
		}
		s.create(&model.PartyMember{Name: p.name, Gender: g, Birth: fmt.Sprintf("%04d-%02d-%02d", s.now.Year()-age, 1+s.r.Intn(12), 1+s.r.Intn(28)),
			Nation: "汉族", Education: s.pick("高中", "大专", "初中", "本科", "中专"), JoinDate: join, FormalDate: formal, Type: typ, Mobile: mobile,
			OrgID: org.ID, OrgName: org.Name, Post: p.post, Phone: s.phone(), WorkPlace: wp, GridID: grids[i%5].ID})
	}

	// 党建活动（三会一课以外）
	acts := []model.Meeting{
		{Title: "“红色村庄我守护”主题党日", Type: "主题党日", MeetDate: s.day(-3), Location: "村部及上屋场", Host: "张建国", Recorder: "陈志强",
			Attendees: "张建国,刘秀梅,陈志强,黄德胜,吴建平,王丽华,李春生,杨文杰", ShouldCount: 12, Topics: "党员志愿服务、环境整治",
			Content: "党员佩戴党徽开展公共环境提升行动，清理卫生死角 12 处。"},
		{Title: "“我为茶叶找销路”党员志愿服务", Type: "党建活动", MeetDate: s.day(-21), Location: "茶园冲茶叶基地", Host: "吴建平", Recorder: "王丽华",
			Attendees: "吴建平,张国强,周德明,杨春华,黄德胜", ShouldCount: 5, Topics: "直播助农", Content: "党员带头参与直播带货，销售秋茶 300 余斤。"},
	}
	for i := range acts {
		acts[i].AttendCount = len(strings.Split(acts[i].Attendees, ","))
		s.create(&acts[i])
	}

	// 民生保障台账：依据人口标签生成
	tagType := map[string]string{"低保": "低保", "五保": "特困供养", "残疾": "残疾人", "留守老人": "留守老人", "留守儿童": "留守儿童", "退役军人": "优抚对象"}
	for _, v := range vs {
		for _, t := range strings.Split(v.Tags, ",") {
			typ := tagType[t]
			if typ == "" {
				continue
			}
			w := &model.Welfare{VillagerID: v.ID, Name: v.Name, IDCard: v.IDCard, GridID: v.GridID, Type: typ, Status: "在保",
				StartDate: s.day(-200 - s.r.Intn(900)), Phone: v.Phone, Guardian: s.pick("王丽华", "李春生", "赵小燕", "周德明", "刘秀梅")}
			switch typ {
			case "低保":
				w.Level, w.Amount = s.pick("A类", "B类", "C类"), float64(300+s.r.Intn(4)*50)
			case "特困供养":
				w.Level, w.Amount = "分散供养", 780
			case "残疾人":
				w.Level = s.pick("肢体三级", "视力二级", "智力四级", "肢体四级")
				w.Amount = 90
			case "优抚对象":
				w.Level = "退役士兵"
			case "留守儿童":
				w.Level = "父母均外出"
			}
			s.create(w)
		}
	}
	// 留守妇女
	cnt := 0
	for _, v := range vs {
		if cnt >= 3 {
			break
		}
		if v.Gender == "女" && v.Relation == "配偶" && v.Employment == "在家务农" && v.Birth > fmt.Sprintf("%d", s.now.Year()-55) {
			s.create(&model.Welfare{VillagerID: v.ID, Name: v.Name, IDCard: v.IDCard, GridID: v.GridID, Type: "留守妇女", Level: "丈夫外出务工",
				Status: "在保", StartDate: s.day(-300), Guardian: "刘秀梅", Phone: v.Phone})
			cnt++
		}
	}
	s.create(&model.Welfare{VillagerID: vs[5].ID, Name: vs[5].Name, IDCard: vs[5].IDCard, GridID: vs[5].GridID, Type: "临时救助", Level: "因病",
		Amount: 2000, Status: "已退出", StartDate: s.day(-60), Guardian: "刘秀梅", Remark: "一次性救助 2000 元，已发放"})

	// 矛盾纠纷
	disputes := []model.Dispute{
		{Title: "宅基地围墙边界纠纷", Category: "邻里纠纷", Parties: "张德华、李国平", Status: "调解成功", Mediator: "张建国", Content: "两户因围墙修建位置发生争执。", Result: "实地测量后重新确定边界，双方签订调解协议。"},
		{Title: "土地流转租金拖欠", Category: "土地山林", Parties: "枫树坪组3户村民、某种植合作社", Status: "调解中", Mediator: "刘秀梅", Content: "合作社拖欠 2 个季度租金。"},
		{Title: "赡养老人分歧", Category: "婚姻家庭", Parties: "周氏兄弟三人", Status: "调解成功", Mediator: "刘秀梅", Content: "兄弟三人就母亲赡养费分担意见不一。", Result: "约定按月轮流照顾、赡养费平摊。"},
		{Title: "山林界址争议", Category: "土地山林", Parties: "河背组、茶园冲组", Status: "待调解", Mediator: "张国强", Content: "两组对一处山林界址存在争议。"},
		{Title: "务工工资纠纷", Category: "债务纠纷", Parties: "罗某与本地包工头", Status: "调解成功", Mediator: "陈志强", Content: "拖欠工资 6800 元。", Result: "包工头当场支付 4000 元，余款一个月内付清。"},
		{Title: "养狗扰邻", Category: "邻里纠纷", Parties: "杨某、黄某", Status: "调解失败", Mediator: "王丽华", Content: "犬只夜间吠叫影响休息。", Result: "已引导当事人通过司法途径解决，并持续跟踪。"},
	}
	for i := range disputes {
		d := &disputes[i]
		d.CaseNo = fmt.Sprintf("TJ%s%03d", s.now.Format("2006"), i+1)
		d.GridID = grids[i%5].ID
		d.AcceptDate = s.day(-5 - i*9)
		s.create(d)
	}

	// 治理积分
	reasons := map[string][]string{
		"爱党爱国": {"参加主题党日志愿服务", "国庆节悬挂国旗"},
		"邻里和睦": {"主动化解邻里矛盾", "帮助独居老人收稻谷"},
		"人居环境": {"“美丽屋场”月度评比优秀", "庭院整洁、垃圾分类到位"},
		"勤劳致富": {"发展茶叶种植带动邻里", "参加技能培训并就业"},
		"移风易俗": {"婚事新办、不收高额彩礼", "丧事简办"},
		"遵纪守法": {"劝阻电信诈骗", "参与森林防火巡查"},
	}
	dims := []string{"爱党爱国", "邻里和睦", "人居环境", "勤劳致富", "移风易俗", "遵纪守法"}
	for i := 0; i < 70; i++ {
		h := hs[s.r.Intn(len(hs))]
		d := dims[s.r.Intn(len(dims))]
		pts := 5 + s.r.Intn(4)*5
		reason := reasons[d][s.r.Intn(2)]
		if i%17 == 16 {
			pts, d, reason = -10, "人居环境", "房前屋后乱堆乱放，经提醒未整改"
		}
		s.create(&model.PointRecord{HouseholdID: h.ID, HouseholdName: h.HeadName, GridID: h.GridID, Dimension: d, Points: pts, Reason: reason,
			RecordDate: s.day(-s.r.Intn(120)), Recorder: s.pick("王丽华", "李春生", "赵小燕", "周德明", "刘秀梅")})
	}

	// 红黑榜
	period := s.now.Format("2006年01月")
	for i, r := range []struct{ board, cat, reason string }{
		{"红榜", "美丽庭院", "庭院整洁美观，种花种菜，评为本月“美丽屋场”示范户。"},
		{"红榜", "孝老爱亲", "数十年如一日照顾瘫痪在床的母亲。"},
		{"红榜", "移风易俗", "儿子婚礼简办，节省开支捐给村老年协会。"},
		{"红榜", "热心公益", "主动参与森林防火巡查和防汛值守。"},
		{"黑榜", "人居环境", "房前屋后杂物乱堆乱放，经多次提醒未整改。"},
		{"黑榜", "违规用火", "在林区边缘焚烧秸秆，已批评教育。"},
	} {
		h := hs[(i*7+3)%len(hs)]
		s.create(&model.Honor{Board: r.board, Target: h.HeadName + "户", HouseholdID: h.ID, GridID: h.GridID, Category: r.cat, Reason: r.reason,
			Period: period, Published: true})
	}

	// 致富带头人
	for _, l := range []model.RichLeader{
		{Name: "吴建平", Gender: "男", Industry: "茶叶", Entity: "青山云雾茶叶专业合作社", Scale: "茶园 460 亩，年加工干茶 12 吨", Drive: 46, Income: 320, IsParty: true,
			Story: "返乡创业 10 年，带领乡亲把荒山变茶园，推行“合作社+农户”保底收购，带动 46 户增收。"},
		{Name: "杨春华", Gender: "女", Industry: "稻虾共作", Entity: "塘湾家庭农场", Scale: "稻虾田 180 亩", Drive: 21, Income: 145, IsParty: true,
			Story: "从外出务工到返乡养虾，探索出“一水两用、一田双收”的稻虾共作模式。"},
		{Name: "罗海林", Gender: "男", Industry: "乡村旅游", Entity: "青山人家农家乐", Scale: "餐位 120 个、客房 12 间", Drive: 15, Income: 86, IsParty: false,
			Story: "依托青山风光开办农家乐，优先收购本村农产品、聘用本村妇女。"},
		{Name: "谢美玲", Gender: "女", Industry: "电商", Entity: "青山好物电商服务站", Scale: "年销售农产品 80 万元", Drive: 60, Income: 80, IsParty: false,
			Story: "通过直播带货把茶叶、腊肉、土蜂蜜卖向全国。"},
		{Name: "张国强", Gender: "男", Industry: "养殖", Entity: "国强生态养牛场", Scale: "存栏肉牛 85 头", Drive: 9, Income: 120, IsParty: true,
			Story: "发展生态养牛，免费为周边农户提供技术指导。"},
	} {
		x := l
		x.Phone = s.phone()
		s.create(&x)
	}

	// 门户内容
	portal := []model.PortalItem{
		{Kind: "banner", Title: "青山村 · 党建引领乡村振兴", Summary: "党建+网格 · 茶旅融合 · 强村富民", Color: "linear-gradient(120deg,#0f766e,#14b8a6 60%,#a3e635)", Sort: 1, Published: true},
		{Kind: "banner", Title: "高山云雾茶 · 一片叶子富一方百姓", Summary: "高山生态茶叶基地扩建进行中，带动 46 户农户增收", Color: "linear-gradient(120deg,#065f46,#059669 55%,#fbbf24)", Sort: 2, Published: true},
		{Kind: "banner", Title: "美丽屋场 · 共建共治共享", Summary: "治理积分制让好人有好报、文明有回报", Color: "linear-gradient(120deg,#1e3a8a,#0891b2 60%,#34d399)", Sort: 3, Published: true},
		{Kind: "news", Title: "青山村开展“红色村庄我守护”主题党日活动", Summary: "全体在家党员佩戴党徽开展公共环境提升行动，清理卫生死角 12 处。", Date: s.day(-3), Published: true},
		{Kind: "news", Title: "秋茶采摘进入尾声，合作社保底收购稳增收", Summary: "今年秋茶收购量同比增长 18%，农户户均增收 2400 元。", Date: s.day(-9), Published: true},
		{Kind: "news", Title: "村民代表大会表决通过道路硬化项目", Summary: "16 名代表全票同意实施村部至枫树坪 1.8 公里道路硬化。", Date: s.day(-17), Published: true},
		{Kind: "news", Title: "网格员上门协助老人缴纳医保", Summary: "5 名网格员分片入户，为行动不便老人提供代缴服务。", Date: s.day(-2), Published: true},
		{Kind: "story", Title: "荒山变茶园：吴建平的返乡创业路", Summary: "十年间，他把 400 多亩荒山变成了致富茶园。",
			Content: "2015 年，在外打拼多年的吴建平回到青山村，看着村后大片撂荒的山地，他决定试种高山茶。\n\n起初村民并不看好，他就带头流转土地、请来农技专家，第一年先种下 60 亩。三年后茶园投产，合作社实行保底收购，越来越多的农户加入进来。\n\n如今，合作社茶园已发展到 460 亩，带动 46 户农户户均年增收 6000 余元。", Date: s.day(-40), Sort: 1, Published: true},
		{Kind: "story", Title: "一块积分卡，换来村里新风尚", Summary: "治理积分制让“小积分”撬动“大治理”。",
			Content: "去年起，青山村推行治理积分制，从爱党爱国、邻里和睦、人居环境、勤劳致富、移风易俗、遵纪守法六个维度为农户记分。\n\n积分可以在村积分超市兑换生活用品，还与评先评优挂钩。村里乱堆乱放少了，主动参与志愿服务的人多了。", Date: s.day(-60), Sort: 2, Published: true},
		{Kind: "story", Title: "稻田里养出“致富虾”", Summary: "稻虾共作让低洼田亩均增收 1800 元。",
			Content: "塘湾组的低洼田过去只能种一季稻，收益很低。2024 年，村里引进稻虾共作模式，“一水两用、一田双收”。\n\n现在示范片面积已达 420 亩，村集体每年还能获得 8 万元分红。", Date: s.day(-90), Sort: 3, Published: true},
	}
	for i := range portal {
		if portal[i].Date == "" {
			portal[i].Date = s.day(-1)
		}
		s.create(&portal[i])
	}

	// 村规民约（作为一类公开内容）
	s.create(&model.Notice{Title: "青山村村规民约（2026 年修订）", Category: "村规民约", Status: "已发布", Publisher: chenzq.RealName, PublishDate: s.day(-70), Top: false, Views: 312,
		Content: "第一条　爱党爱国，遵纪守法，自觉维护社会稳定。\n第二条　邻里互助，团结友爱，有矛盾纠纷通过协商或调解解决。\n第三条　爱护环境，房前屋后不乱堆乱放，生活垃圾分类投放。\n第四条　移风易俗，红白喜事简办，不搞攀比，不收高额彩礼。\n第五条　孝老爱亲，赡养老人，关爱留守儿童。\n第六条　森林防火期内不在林区野外用火。\n第七条　爱护公共设施，不私自占用集体土地和道路。\n第八条　积极参与治理积分评比，积分与评先评优挂钩。\n\n本村规民约经村民代表会议讨论通过，自公布之日起施行。"})
	_ = liuxm
}
