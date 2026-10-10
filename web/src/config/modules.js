// 各业务模块的列表/表单/筛选配置。新增模块只需在此声明并在路由中引用。
import dayjs from 'dayjs'
import { store } from '@/store'
import { ageOf, fmtTime, money } from '@/utils/format'

const AC = ['admin', 'cadre']
const today = () => dayjs().format('YYYY-MM-DD')

export const OPT = {
  householdType: ['一般户', '低保户', '五保户', '脱贫户', '监测户'],
  gender: ['男', '女'],
  political: ['中共党员', '中共预备党员', '共青团员', '民主党派', '群众'],
  education: ['学龄前', '小学在读', '初中在读', '高中在读', '中专在读', '文盲或半文盲', '小学', '初中', '高中', '中专', '大专', '本科', '研究生'],
  employment: ['在家务农', '本地务工', '外出务工', '个体经营', '在校学生', '学龄前', '退休/养老', '无业'],
  relation: ['户主', '配偶', '长子', '次子', '长女', '次女', '儿媳', '女婿', '孙子', '孙女', '父亲', '母亲', '兄弟姐妹', '其他'],
  tags: ['低保', '五保', '残疾', '脱贫监测户', '留守老人', '留守儿童', '留守妇女', '退役军人'],
  eventCategory: ['矛盾纠纷', '安全隐患', '环境卫生', '基础设施', '民生诉求', '森林防火', '防汛抗旱', '社会治安', '其他'],
  eventStatus: ['上报', '受理', '处理中', '办结', '退回'],
  urgency: ['一般', '紧急', '特急'],
  taskStatus: ['待办', '进行中', '已完成'],
  priority: ['高', '中', '低'],
  taskCategory: ['民生保障', '森林防火', '防返贫监测', '人居环境', '安全生产', '集体经济', '党建工作', '优抚安置', '关爱服务', '其他'],
  logType: ['工作日志', '民情日记', '走访记录'],
  noticeCategory: ['通知公告', '党务公开', '村务公开', '财务公开', '村规民约'],
  projectType: ['种植业', '养殖业', '乡村旅游', '新能源', '基础设施', '农产品加工', '电商', '其他'],
  projectStatus: ['规划中', '建设中', '已完工', '运营中'],
  fundCategory: ['补助收入', '经营收入', '发包及上交收入', '投资收益', '其他收入', '管理费用', '公益设施支出', '项目支出', '其他支出'],
  assetCategory: ['资源性', '经营性', '非经营性'],
  assetStatus: ['在用', '闲置', '出租', '承包', '报废'],
  meetingType: ['支部党员大会', '支委会', '党小组会', '党课', '主题党日', '党建活动', '村民代表大会', '村民会议', '工作会议', '其他'],
  welfareType: ['低保', '特困供养', '优抚对象', '残疾人', '留守老人', '留守儿童', '留守妇女', '临时救助'],
  disputeCategory: ['邻里纠纷', '婚姻家庭', '土地山林', '债务纠纷', '其他'],
  disputeStatus: ['待调解', '调解中', '调解成功', '调解失败'],
  dimensions: ['爱党爱国', '邻里和睦', '人居环境', '勤劳致富', '移风易俗', '遵纪守法']
}

const householdRemote = (extra = {}) => ({ type: 'remote', url: '/households', labelKey: 'head_name', subKey: 'address', nameProp: 'household_name', ...extra })

export const M = {
  groups: {
    title: '村民小组', api: '/groups', writeRoles: AC, exportable: true,
    fields: [
      { prop: 'name', label: '小组名称', required: true },
      { prop: 'leader', label: '组长', mobile: 'meta' },
      { prop: 'phone', label: '联系电话', mobile: 'meta' },
      { prop: 'sort', label: '排序', type: 'number', list: false, default: 0 }
    ]
  },
  grids: {
    title: '网格', api: '/grids', writeRoles: AC,
    fields: [
      { prop: 'name', label: '网格名称', required: true, minWidth: 160 },
      { prop: 'code', label: '网格编码', mobile: 'meta' },
      { prop: 'group_id', label: '所属小组', type: 'select', options: () => store._groups || [], fmt: r => (store._groups || []).find(g => g.value === r.group_id)?.label || '' },
      { prop: 'worker_id', label: '网格员', type: 'user', userRole: ['grid', 'cadre'], nameProp: 'worker_name', mobile: 'meta' },
      { prop: 'scope', label: '网格范围', type: 'textarea', minWidth: 220 },
      { prop: 'sort', label: '排序', type: 'number', list: false, default: 0 }
    ]
  },

  households: {
    title: '户籍台账', itemName: '家庭户', api: '/households', deleteRoles: AC, searchPlaceholder: '户号 / 户主 / 地址 / 电话',
    mobileTitle: r => `${r.head_name} 户`,
    filters: [{ prop: 'grid_id', label: '网格', type: 'grid' }, { prop: 'type', label: '户类型', type: 'select', options: OPT.householdType }],
    importable: { headers: ['户号', '户主', '户类型', '网格', '家庭住址', '联系电话', '家庭年收入', '住房情况', '备注'], sample: ['QS01999', '张三', '一般户', '第一网格（上屋场）', '青山村上屋场1号', '13800000000', '50000', '砖混结构自建房', ''] },
    fields: [
      { prop: 'head_name', label: '户主', required: true, width: 100 },
      { prop: 'household_no', label: '户号', width: 110, mobile: 'meta', placeholder: '留空自动生成' },
      { prop: 'type', label: '户类型', type: 'select', options: OPT.householdType, tag: true, width: 100, mobile: 'tag', default: '一般户' },
      { prop: 'grid_id', label: '所属网格', type: 'grid', required: true, minWidth: 150, mobile: 'meta' },
      { prop: 'address', label: '家庭住址', minWidth: 200, mobile: 'meta' },
      { prop: 'phone', label: '联系电话', width: 125 },
      { prop: 'member_count', label: '人口数', type: 'number', width: 80, form: false, mobile: 'meta' },
      { prop: 'income', label: '家庭年收入(元)', type: 'money', width: 130, step: 1000, min: 0 },
      { prop: 'housing', label: '住房情况', type: 'select', allowCreate: true, options: ['砖混结构自建房', '框架结构楼房', '土木结构老屋（已鉴定C级）', '危房（D级）', '易地搬迁安置房'], list: false },
      { prop: 'remark', label: '备注', type: 'textarea', list: false }
    ]
  },

  villagers: {
    title: '人口台账', itemName: '村民', api: '/villagers', deleteRoles: AC, searchPlaceholder: '姓名 / 身份证号 / 电话',
    fetchDetail: true, actionWidth: 170,
    loadForEdit: d => d,
    filters: [
      { prop: 'grid_id', label: '网格', type: 'grid' },
      { prop: 'gender', label: '性别', type: 'select', options: OPT.gender },
      { prop: 'tag', label: '特殊人群', type: 'select', options: OPT.tags },
      { prop: 'political', label: '政治面貌', type: 'select', options: OPT.political },
      { prop: 'employment', label: '就业状态', type: 'select', options: OPT.employment }
    ],
    importable: {
      headers: ['户号', '姓名', '身份证号', '性别', '出生日期', '与户主关系', '民族', '政治面貌', '文化程度', '就业状态', '务工地点', '联系电话', '特殊人群标签'],
      sample: ['QS01001', '张小明', '', '男', '2015-03-02', '孙子', '汉族', '群众', '小学在读', '在校学生', '', '', '留守儿童']
    },
    fields: [
      { prop: 'name', label: '姓名', required: true, width: 90 },
      { prop: 'household_id', label: '所属家庭', ...householdRemote({ nameProp: '_hh' }), required: true, list: false, detail: false },
      { prop: 'id_card', label: '身份证号', width: 175, placeholder: '18 位，自动识别性别与出生日期', maxlength: 18 },
      { prop: 'gender', label: '性别', type: 'select', options: OPT.gender, width: 64, mobile: 'meta' },
      { prop: 'birth', label: '出生日期', type: 'date', width: 110 },
      { prop: '_age', label: '年龄', type: 'computed', fmt: r => ageOf(r.birth), width: 64, form: false, mobile: 'meta' },
      { prop: 'relation', label: '与户主关系', type: 'select', options: OPT.relation, allowCreate: true, width: 100 },
      { prop: 'nation', label: '民族', default: '汉族', list: false },
      { prop: 'political', label: '政治面貌', type: 'select', options: OPT.political, width: 100, default: '群众' },
      { prop: 'education', label: '文化程度', type: 'select', options: OPT.education, width: 100 },
      { prop: 'employment', label: '就业状态', type: 'select', options: OPT.employment, width: 100, mobile: 'meta' },
      { prop: 'work_place', label: '务工地点', list: false },
      { prop: 'phone', label: '联系电话', width: 120 },
      { prop: 'grid_id', label: '网格', type: 'grid', form: false, minWidth: 140 },
      { prop: 'tags', label: '特殊人群标签', type: 'tags', options: OPT.tags, minWidth: 180, span: 24 },
      { prop: 'remark', label: '备注', type: 'textarea', list: false }
    ]
  },

  patrols: {
    title: '网格巡查', itemName: '巡查记录', api: '/patrols', searchPlaceholder: '路线 / 内容 / 问题 / 巡查人', createText: '记录巡查',
    mobileTitle: r => `${r.patrol_date} · ${r.user_name}`,
    filters: [{ prop: 'grid_id', label: '网格', type: 'grid' }, { prop: '_range', label: '日期', type: 'daterange' }],
    fields: [
      { prop: 'patrol_date', label: '巡查日期', type: 'date', required: true, default: today, width: 110 },
      { prop: 'grid_id', label: '巡查网格', type: 'grid', required: true, minWidth: 150, mobile: 'meta' },
      { prop: 'user_name', label: '巡查人', form: false, width: 90 },
      { prop: 'start_time', label: '开始时间', type: 'time', list: false, default: () => dayjs().subtract(1, 'hour').format('HH:mm') },
      { prop: 'end_time', label: '结束时间', type: 'time', list: false, default: () => dayjs().format('HH:mm') },
      { prop: 'route', label: '巡查路线', minWidth: 160, mobile: 'meta' },
      { prop: 'content', label: '巡查情况', type: 'textarea', required: true, minWidth: 220, mobile: 'meta' },
      { prop: 'problems', label: '发现问题', type: 'textarea', minWidth: 160, placeholder: '无问题可不填；需处置的问题请同时“上报事件”' },
      { prop: 'images', label: '现场照片', type: 'images', width: 120, default: '[]' }
    ]
  },

  events: {
    title: '事件上报', itemName: '事件', api: '/events', deleteRoles: ['admin'], searchPlaceholder: '编号 / 标题 / 位置 / 上报人', createText: '上报事件',
    canEdit: r => ['上报', '退回'].includes(r.status) && (store.user.role !== 'grid' || r.reporter_id === store.user.id),
    rowClass: ({ row }) => (row.urgency === '特急' && row.status !== '办结' ? 'row-hot' : ''),
    filters: [
      { prop: 'status', label: '状态', type: 'select', options: OPT.eventStatus },
      { prop: 'category', label: '类别', type: 'select', options: OPT.eventCategory },
      { prop: 'grid_id', label: '网格', type: 'grid' },
      { prop: '_range', label: '日期', type: 'daterange' },
      { prop: 'mine', label: '与我相关', type: 'check' }
    ],
    fields: [
      { prop: 'title', label: '事件标题', required: true, minWidth: 180 },
      { prop: 'event_no', label: '事件编号', form: false, width: 150, list: false },
      { prop: 'category', label: '事件类别', type: 'select', options: OPT.eventCategory, required: true, width: 100, mobile: 'meta' },
      { prop: 'urgency', label: '紧急程度', type: 'select', options: OPT.urgency, tag: true, default: '一般', width: 90, mobile: 'tag' },
      { prop: 'grid_id', label: '所属网格', type: 'grid', required: true, minWidth: 140 },
      { prop: 'location', label: '发生位置', minWidth: 160, mobile: 'meta' },
      { prop: 'description', label: '事件描述', type: 'textarea', required: true, list: false },
      { prop: 'images', label: '现场图片', type: 'images', default: '[]', list: false },
      { prop: 'status', label: '状态', tag: true, form: false, width: 80, mobile: 'tag' },
      { prop: 'reporter', label: '上报人', form: false, width: 80 },
      { prop: 'handler', label: '处理人', form: false, width: 80 },
      { prop: 'created_at', label: '上报时间', form: false, fmt: r => fmtTime(r.created_at), width: 140, mobile: 'meta' }
    ]
  },

  tasks: {
    title: '任务督办', itemName: '任务', api: '/tasks', writeRoles: AC, searchPlaceholder: '任务 / 内容 / 责任人', createText: '交办任务',
    rowClass: ({ row }) => (row.overdue ? 'row-overdue' : ''),
    filters: [
      { prop: 'status', label: '状态', type: 'select', options: OPT.taskStatus },
      { prop: 'priority', label: '优先级', type: 'select', options: OPT.priority },
      { prop: 'assignee_id', label: '责任人', type: 'user', roles: AC },
      { prop: 'overdue', label: '仅看逾期', type: 'check' },
      { prop: 'mine', label: '我负责的', type: 'check', roles: AC }
    ],
    fields: [
      { prop: 'title', label: '任务标题', required: true, minWidth: 200, span: 24 },
      { prop: 'category', label: '任务类别', type: 'select', options: OPT.taskCategory, allowCreate: true, width: 100 },
      { prop: 'priority', label: '优先级', type: 'select', options: OPT.priority, tag: true, default: '中', width: 76, mobile: 'tag' },
      { prop: 'assignee_id', label: '责任人', type: 'user', nameProp: 'assignee_name', required: true, width: 90, mobile: 'meta' },
      { prop: 'deadline', label: '截止日期', type: 'date', required: true, width: 110, mobile: 'meta', fmt: r => r.deadline + (r.overdue ? '（逾期）' : '') },
      { prop: 'status', label: '状态', type: 'select', options: OPT.taskStatus, tag: true, default: '待办', width: 86, mobile: 'tag' },
      { prop: 'progress', label: '进度', type: 'progress', width: 150, default: 0, mobile: 'meta' },
      { prop: 'content', label: '任务要求', type: 'textarea', list: false },
      { prop: 'creator_name', label: '交办人', form: false, width: 80 }
    ]
  },

  worklogs: {
    title: '工作日志 / 民情日记', itemName: '日志', api: '/worklogs', searchPlaceholder: '标题 / 内容 / 走访户', createText: '写日志',
    canEdit: r => store.user.role === 'admin' || r.user_id === store.user.id,
    canDelete: r => store.user.role === 'admin' || r.user_id === store.user.id,
    mobileTitle: r => r.title || r.type,
    filters: [
      { prop: 'type', label: '类型', type: 'select', options: OPT.logType },
      { prop: 'user_id', label: '记录人', type: 'user', roles: AC },
      { prop: '_range', label: '日期', type: 'daterange' },
      { prop: 'mine', label: '只看我的', type: 'check', roles: AC }
    ],
    fields: [
      { prop: 'log_date', label: '日期', type: 'date', required: true, default: today, width: 110, mobile: 'meta' },
      { prop: 'type', label: '类型', type: 'select', options: OPT.logType, required: true, default: '工作日志', tag: true, width: 96, mobile: 'tag' },
      { prop: 'title', label: '标题', minWidth: 180 },
      { prop: 'household_id', label: '走访户', ...householdRemote(), showIf: f => f.type === '走访记录' || f.household_id, list: false, detail: false },
      { prop: 'household_name', label: '走访户', form: false, width: 90, mobile: 'meta' },
      { prop: 'content', label: '内容', type: 'textarea', required: true, minWidth: 260, rows: 6 },
      { prop: 'demand', label: '群众诉求', type: 'textarea', rows: 2, list: false },
      { prop: 'followup', label: '跟进情况', type: 'textarea', rows: 2, list: false },
      { prop: 'user_name', label: '记录人', form: false, width: 80, mobile: 'meta' }
    ]
  },

  notices: {
    title: '通知公告 / 村务公开', itemName: '公告', api: '/notices', writeRoles: AC, searchPlaceholder: '标题 / 内容', formWidth: '860px', detailWidth: '720px',
    filters: [
      { prop: 'category', label: '栏目', type: 'select', options: OPT.noticeCategory },
      { prop: 'status', label: '状态', type: 'select', options: ['已发布', '草稿'], roles: AC },
      { prop: '_range', label: '日期', type: 'daterange' }
    ],
    fields: [
      { prop: 'title', label: '标题', required: true, minWidth: 260, span: 24 },
      { prop: 'category', label: '栏目', type: 'select', options: OPT.noticeCategory, required: true, default: '通知公告', tag: true, width: 96, mobile: 'tag' },
      { prop: 'publish_date', label: '发布日期', type: 'date', default: today, width: 110, mobile: 'meta' },
      { prop: 'publisher', label: '发布人', width: 90, placeholder: '默认当前用户' },
      { prop: 'status', label: '状态', type: 'select', options: ['已发布', '草稿'], default: '已发布', tag: true, width: 80, roles: AC },
      { prop: 'top', label: '置顶', type: 'switch', width: 64 },
      { prop: 'views', label: '浏览', form: false, width: 64, mobile: 'meta' },
      { prop: 'content', label: '正文', type: 'textarea', rows: 12, list: false }
    ]
  },

  projects: {
    title: '产业项目', itemName: '项目', api: '/projects', writeRoles: AC, searchPlaceholder: '项目名称 / 负责人 / 地点', fetchDetail: false,
    importable: { headers: ['项目名称', '类型', '负责人', '地点', '预算(万元)', '已投入(万元)', '进度(%)', '状态', '开始日期', '计划完成', '带动效益'] },
    filters: [{ prop: 'type', label: '类型', type: 'select', options: OPT.projectType }, { prop: 'status', label: '状态', type: 'select', options: OPT.projectStatus }],
    fields: [
      { prop: 'name', label: '项目名称', required: true, minWidth: 180, span: 24 },
      { prop: 'type', label: '类型', type: 'select', options: OPT.projectType, width: 90, mobile: 'meta' },
      { prop: 'status', label: '状态', type: 'select', options: OPT.projectStatus, tag: true, default: '规划中', width: 86, mobile: 'tag' },
      { prop: 'leader', label: '负责人', width: 80 },
      { prop: 'location', label: '地点', minWidth: 120, list: false },
      { prop: 'budget', label: '预算(万元)', type: 'money', step: 1, min: 0, width: 100, mobile: 'meta' },
      { prop: 'invested', label: '已投入(万元)', type: 'money', step: 1, min: 0, width: 110 },
      { prop: 'progress', label: '进度', type: 'progress', default: 0, width: 150, mobile: 'meta' },
      { prop: 'start_date', label: '开始日期', type: 'date', width: 110 },
      { prop: 'end_date', label: '计划完成', type: 'date', width: 110, list: false },
      { prop: 'benefit', label: '带动效益', type: 'textarea', rows: 2, minWidth: 200 },
      { prop: 'description', label: '项目简介', type: 'textarea', list: false }
    ]
  },

  funds: {
    title: '资金台账', itemName: '收支记录', api: '/funds', writeRoles: AC, searchPlaceholder: '摘要 / 凭证号 / 经办人 / 科目',
    importable: { headers: ['日期', '凭证号', '收支', '科目', '金额(元)', '摘要', '经办人'], sample: ['2026-10-01', '记-2610-001', '收入', '经营收入', '6500', '光伏电站发电收益', '陈志强'] },
    filters: [
      { prop: 'type', label: '收支', type: 'select', options: ['收入', '支出'] },
      { prop: 'category', label: '科目', type: 'select', options: OPT.fundCategory },
      { prop: '_range', label: '日期', type: 'daterange' }
    ],
    mobileTitle: r => r.summary,
    fields: [
      { prop: 'date', label: '日期', type: 'date', required: true, default: today, width: 110, mobile: 'meta' },
      { prop: 'voucher_no', label: '凭证号', width: 120 },
      { prop: 'type', label: '收支', type: 'select', options: ['收入', '支出'], required: true, tag: true, width: 70, mobile: 'tag' },
      { prop: 'category', label: '科目', type: 'select', options: OPT.fundCategory, allowCreate: true, width: 120, mobile: 'meta' },
      { prop: 'amount', label: '金额(元)', type: 'money', required: true, min: 0, step: 100, precision: 2, width: 120, mobile: 'meta' },
      { prop: 'summary', label: '摘要', required: true, minWidth: 220, span: 24 },
      { prop: 'handler', label: '经办人', width: 80 },
      { prop: 'project_id', label: '关联项目', type: 'select', options: () => store._projects || [], list: false, fmt: r => (store._projects || []).find(p => p.value === r.project_id)?.label || '' }
    ]
  },

  assets: {
    title: '集体资产', itemName: '资产', api: '/assets', writeRoles: AC, searchPlaceholder: '编号 / 名称 / 位置 / 管理人',
    importable: { headers: ['资产编号', '资产名称', '类别', '数量', '单位', '价值(万元)', '坐落位置', '状态', '管理人', '取得日期'] },
    filters: [{ prop: 'category', label: '类别', type: 'select', options: OPT.assetCategory }, { prop: 'status', label: '状态', type: 'select', options: OPT.assetStatus }],
    fields: [
      { prop: 'name', label: '资产名称', required: true, minWidth: 160 },
      { prop: 'code', label: '资产编号', width: 110, placeholder: '留空自动生成' },
      { prop: 'category', label: '类别', type: 'select', options: OPT.assetCategory, width: 90, mobile: 'meta' },
      { prop: 'quantity', label: '数量', type: 'number', min: 0, width: 80, fmt: r => `${r.quantity}${r.unit || ''}`, mobile: 'meta' },
      { prop: 'unit', label: '单位', list: false },
      { prop: 'value', label: '价值(万元)', type: 'money', min: 0, step: 1, width: 100, mobile: 'meta' },
      { prop: 'location', label: '坐落位置', minWidth: 140 },
      { prop: 'status', label: '状态', type: 'select', options: OPT.assetStatus, tag: true, default: '在用', width: 76, mobile: 'tag' },
      { prop: 'keeper', label: '管理人', width: 80 },
      { prop: 'acquired_at', label: '取得日期', type: 'date', width: 110 },
      { prop: 'remark', label: '备注', type: 'textarea', list: false }
    ]
  },

  leaders: {
    title: '致富带头人', itemName: '带头人', api: '/leaders', writeRoles: AC, searchPlaceholder: '姓名 / 产业 / 经营主体',
    importable: { headers: ['姓名', '性别', '产业类型', '经营主体', '经营规模', '带动户数', '年产值(万元)', '党员', '联系电话'] },
    fields: [
      { prop: 'name', label: '姓名', required: true, width: 90 },
      { prop: 'gender', label: '性别', type: 'select', options: OPT.gender, width: 64 },
      { prop: 'industry', label: '产业类型', width: 90, tag: true, mobile: 'tag' },
      { prop: 'entity', label: '经营主体', minWidth: 170, mobile: 'meta' },
      { prop: 'scale', label: '经营规模', minWidth: 160 },
      { prop: 'drive', label: '带动户数', type: 'number', min: 0, width: 90, mobile: 'meta' },
      { prop: 'income', label: '年产值(万元)', type: 'money', min: 0, width: 110 },
      { prop: 'is_party', label: '党员', type: 'switch', width: 64 },
      { prop: 'phone', label: '联系电话', width: 120, list: false },
      { prop: 'story', label: '创业事迹', type: 'textarea', list: false }
    ]
  },

  meetings: {
    title: '三会一课 / 会议活动', itemName: '会议', api: '/meetings', writeRoles: AC, searchPlaceholder: '会议名称 / 议题 / 主持人 / 参会人', formWidth: '860px',
    filters: [{ prop: 'type', label: '类型', type: 'select', options: OPT.meetingType }, { prop: '_range', label: '日期', type: 'daterange' }],
    fields: [
      { prop: 'title', label: '会议名称', required: true, minWidth: 220, span: 24 },
      { prop: 'type', label: '类型', type: 'select', options: OPT.meetingType, required: true, tag: true, width: 110, mobile: 'tag' },
      { prop: 'meet_date', label: '日期', type: 'date', required: true, default: today, width: 110, mobile: 'meta' },
      { prop: 'location', label: '地点', width: 140 },
      { prop: 'host', label: '主持人', width: 80, mobile: 'meta' },
      { prop: 'recorder', label: '记录人', list: false },
      { prop: 'should_count', label: '应到人数', type: 'number', min: 0, list: false },
      { prop: 'attend_count', label: '实到', form: false, width: 64, fmt: r => `${r.attend_count}/${r.should_count || '-'}`, mobile: 'meta' },
      { prop: 'attendees', label: '参会人员', type: 'textarea', rows: 2, list: false, placeholder: '姓名之间用逗号分隔，系统自动统计实到人数' },
      { prop: 'absent', label: '缺席情况', list: false },
      { prop: 'topics', label: '议题', type: 'textarea', rows: 2, minWidth: 200 },
      { prop: 'content', label: '会议内容', type: 'textarea', list: false },
      { prop: 'resolution', label: '会议决议', type: 'textarea', rows: 2, list: false }
    ]
  },

  partyMembers: {
    title: '党员花名册', itemName: '党员', api: '/party/members', writeRoles: AC, searchPlaceholder: '姓名 / 电话 / 职务',
    importable: { headers: ['姓名', '性别', '出生日期', '民族', '文化程度', '入党日期', '转正日期', '党员类型', '流动党员', '党内职务', '联系电话', '工作单位/务工地'] },
    filters: [
      { prop: 'org_id', label: '所属组织', type: 'select', options: () => store._orgs || [] },
      { prop: 'type', label: '类型', type: 'select', options: ['正式党员', '预备党员'] },
      { prop: 'mobile', label: '流动党员', type: 'select', options: [{ label: '流动党员', value: 'true' }, { label: '在村党员', value: 'false' }] }
    ],
    fields: [
      { prop: 'name', label: '姓名', required: true, width: 90 },
      { prop: 'gender', label: '性别', type: 'select', options: OPT.gender, width: 64 },
      { prop: 'birth', label: '出生日期', type: 'date', width: 110, list: false },
      { prop: '_age', label: '年龄', type: 'computed', form: false, fmt: r => ageOf(r.birth), width: 64, mobile: 'meta' },
      { prop: 'nation', label: '民族', default: '汉族', list: false },
      { prop: 'education', label: '文化程度', type: 'select', options: OPT.education, width: 90 },
      { prop: 'join_date', label: '入党日期', type: 'date', width: 110, mobile: 'meta' },
      { prop: 'formal_date', label: '转正日期', type: 'date', list: false },
      { prop: 'type', label: '党员类型', type: 'select', options: ['正式党员', '预备党员'], default: '正式党员', tag: true, width: 96, mobile: 'tag' },
      { prop: 'mobile', label: '流动党员', type: 'switch', width: 86 },
      { prop: 'org_id', label: '所属组织', type: 'select', options: () => store._orgs || [], nameProp: 'org_name', fmt: r => r.org_name, minWidth: 170 },
      { prop: 'post', label: '党内职务', width: 100, mobile: 'meta' },
      { prop: 'phone', label: '联系电话', width: 120 },
      { prop: 'work_place', label: '工作单位/务工地', list: false },
      { prop: 'remark', label: '备注', type: 'textarea', list: false }
    ]
  },

  welfare: {
    title: '民生保障台账', itemName: '保障对象', api: '/welfare', deleteRoles: AC, searchPlaceholder: '姓名 / 帮扶责任人',
    importable: { headers: ['姓名', '网格', '保障类型', '类别/等级', '月保障金额(元)', '纳入日期', '状态', '帮扶责任人', '联系电话', '备注'] },
    filters: [
      { prop: 'type', label: '保障类型', type: 'select', options: OPT.welfareType },
      { prop: 'status', label: '状态', type: 'select', options: ['在保', '已退出'] },
      { prop: 'grid_id', label: '网格', type: 'grid' }
    ],
    fields: [
      { prop: 'villager_id', label: '保障对象', type: 'remote', url: '/villagers', labelKey: 'name', subKey: 'birth', nameProp: 'name', list: false, detail: false, onPick: (o, f) => { if (o) f.grid_id = o.grid_id } },
      { prop: 'name', label: '姓名', required: true, width: 90, placeholder: '可从上方选择村民自动带出' },
      { prop: 'id_card', label: '身份证号', form: false, width: 170 },
      { prop: 'type', label: '保障类型', type: 'select', options: OPT.welfareType, required: true, tag: true, width: 96, mobile: 'tag' },
      { prop: 'level', label: '类别/等级', width: 100, mobile: 'meta' },
      { prop: 'amount', label: '月保障金(元)', type: 'money', min: 0, step: 10, width: 110, mobile: 'meta' },
      { prop: 'grid_id', label: '网格', type: 'grid', required: true, minWidth: 140 },
      { prop: 'start_date', label: '纳入日期', type: 'date', width: 110 },
      { prop: 'status', label: '状态', type: 'select', options: ['在保', '已退出'], default: '在保', tag: true, width: 76 },
      { prop: 'guardian', label: '帮扶责任人', width: 96, mobile: 'meta' },
      { prop: 'phone', label: '联系电话', list: false },
      { prop: 'remark', label: '备注', type: 'textarea', list: false }
    ]
  },

  disputes: {
    title: '矛盾纠纷调解', itemName: '纠纷', api: '/disputes', deleteRoles: AC, searchPlaceholder: '编号 / 事由 / 当事人 / 调解员',
    importable: { headers: ['纠纷事由', '类别', '当事人', '网格', '受理日期', '调解员', '状态', '纠纷简况', '调解结果'] },
    filters: [
      { prop: 'status', label: '状态', type: 'select', options: OPT.disputeStatus },
      { prop: 'category', label: '类别', type: 'select', options: OPT.disputeCategory },
      { prop: 'grid_id', label: '网格', type: 'grid' }
    ],
    fields: [
      { prop: 'title', label: '纠纷事由', required: true, minWidth: 180 },
      { prop: 'case_no', label: '编号', form: false, width: 120 },
      { prop: 'category', label: '类别', type: 'select', options: OPT.disputeCategory, width: 90, mobile: 'meta' },
      { prop: 'parties', label: '当事人', minWidth: 140, mobile: 'meta' },
      { prop: 'grid_id', label: '网格', type: 'grid', required: true, minWidth: 140 },
      { prop: 'accept_date', label: '受理日期', type: 'date', default: today, width: 110 },
      { prop: 'mediator', label: '调解员', width: 80 },
      { prop: 'status', label: '状态', type: 'select', options: OPT.disputeStatus, default: '待调解', tag: true, width: 90, mobile: 'tag' },
      { prop: 'content', label: '纠纷简况', type: 'textarea', list: false },
      { prop: 'result', label: '调解结果', type: 'textarea', list: false }
    ]
  },

  rules: {
    title: '村规民约', itemName: '村规民约', api: '/notices', writeRoles: AC, exportable: false, formWidth: '860px', detailWidth: '720px',
    defaultQuery: { category: '村规民约' },
    fields: [
      { prop: 'title', label: '标题', required: true, minWidth: 260, span: 24 },
      { prop: 'category', label: '栏目', default: '村规民约', list: false, form: false },
      { prop: 'publish_date', label: '公布日期', type: 'date', default: today, width: 110, mobile: 'meta' },
      { prop: 'status', label: '状态', type: 'select', options: ['已发布', '草稿'], default: '已发布', tag: true, width: 80 },
      { prop: 'views', label: '浏览', form: false, width: 70 },
      { prop: 'content', label: '条款内容', type: 'textarea', rows: 14, list: false }
    ],
    beforeSave: d => ({ ...d, category: '村规民约' })
  },

  honors: {
    title: '红黑榜', itemName: '上榜记录', api: '/honors', writeRoles: AC, searchPlaceholder: '对象 / 事由',
    filters: [{ prop: 'board', label: '榜单', type: 'select', options: ['红榜', '黑榜'] }],
    fields: [
      { prop: 'board', label: '榜单', type: 'select', options: ['红榜', '黑榜'], required: true, tag: true, default: '红榜', width: 76, mobile: 'tag' },
      { prop: 'household_id', label: '关联农户', ...householdRemote({ nameProp: 'target' }), list: false, detail: false },
      { prop: 'target', label: '上榜对象', required: true, width: 120 },
      { prop: 'category', label: '类别', width: 100, mobile: 'meta' },
      { prop: 'reason', label: '事由', type: 'textarea', rows: 3, minWidth: 260, mobile: 'meta' },
      { prop: 'period', label: '期次', width: 110, placeholder: '如 2026年10月，留空为当月' },
      { prop: 'published', label: '公示', type: 'switch', width: 64, default: true }
    ]
  },

  points: {
    title: '治理积分记录', itemName: '积分记录', api: '/points', deleteRoles: AC, searchPlaceholder: '户主 / 事由 / 记录人', createText: '记录积分',
    mobileTitle: r => `${r.household_name} 户 ${r.points > 0 ? '+' : ''}${r.points}`,
    filters: [{ prop: 'dimension', label: '维度', type: 'select', options: OPT.dimensions }, { prop: 'grid_id', label: '网格', type: 'grid' }, { prop: '_range', label: '日期', type: 'daterange' }],
    fields: [
      { prop: 'household_id', label: '家庭户', ...householdRemote(), required: true, list: false, detail: false },
      { prop: 'household_name', label: '户主', form: false, width: 90 },
      { prop: 'grid_id', label: '网格', type: 'grid', form: false, minWidth: 140 },
      { prop: 'dimension', label: '积分维度', type: 'select', options: OPT.dimensions, required: true, width: 100, tag: true, mobile: 'tag' },
      { prop: 'points', label: '积分', type: 'number', required: true, min: -100, max: 100, step: 5, default: 10, width: 70, fmt: r => (r.points > 0 ? '+' : '') + r.points },
      { prop: 'reason', label: '事由', minWidth: 200, mobile: 'meta', span: 24 },
      { prop: 'record_date', label: '日期', type: 'date', default: today, width: 110, mobile: 'meta' },
      { prop: 'recorder', label: '记录人', form: false, width: 80 }
    ]
  },

  portal: {
    title: '门户内容管理', itemName: '门户内容', api: '/portal', writeRoles: AC, exportable: false, formWidth: '860px',
    filters: [{ prop: 'kind', label: '类型', type: 'select', options: [{ label: '首页横幅', value: 'banner' }, { label: '工作快讯', value: 'news' }, { label: '振兴故事', value: 'story' }] }],
    fields: [
      { prop: 'kind', label: '类型', type: 'select', required: true, default: 'news', options: [{ label: '首页横幅', value: 'banner' }, { label: '工作快讯', value: 'news' }, { label: '振兴故事', value: 'story' }], width: 96, mobile: 'tag' },
      { prop: 'title', label: '标题', required: true, minWidth: 240, span: 24 },
      { prop: 'summary', label: '摘要/副标题', type: 'textarea', rows: 2, minWidth: 200 },
      { prop: 'content', label: '正文', type: 'textarea', rows: 8, list: false, showIf: f => f.kind !== 'banner' },
      { prop: 'image', label: '图片地址', list: false, placeholder: '可选，填写图片 URL 或上传后粘贴' },
      { prop: 'color', label: '横幅背景', list: false, showIf: f => f.kind === 'banner', placeholder: 'CSS 渐变，如 linear-gradient(120deg,#0f766e,#14b8a6)' },
      { prop: 'link', label: '跳转链接', list: false, placeholder: 'http(s):// 或 #/ 开头' },
      { prop: 'date', label: '日期', type: 'date', default: today, width: 110, mobile: 'meta' },
      { prop: 'sort', label: '排序', type: 'number', default: 0, width: 70 },
      { prop: 'published', label: '发布', type: 'switch', default: true, width: 70 }
    ]
  },

  users: {
    title: '用户管理', itemName: '用户', api: '/users', writeRoles: ['admin'], searchPlaceholder: '账号 / 姓名 / 电话',
    canDelete: r => r.id !== store.user.id,
    filters: [{ prop: 'role', label: '角色', type: 'select', options: [{ label: '管理员', value: 'admin' }, { label: '村干部', value: 'cadre' }, { label: '网格员', value: 'grid' }] }],
    fields: [
      { prop: 'username', label: '登录账号', required: true, width: 110, disabled: f => !!f.id },
      { prop: 'real_name', label: '姓名', required: true, width: 90 },
      { prop: 'password', label: '初始密码', list: false, detail: false, showIf: f => !f.id, required: true, placeholder: '至少 8 位，含字母和数字' },
      { prop: 'role', label: '角色', type: 'select', required: true, tag: true, width: 90, mobile: 'tag', default: 'grid',
        options: [{ label: '管理员', value: 'admin' }, { label: '村干部', value: 'cadre' }, { label: '网格员', value: 'grid' }] },
      { prop: 'position', label: '职务', minWidth: 160, mobile: 'meta' },
      { prop: 'phone', label: '电话', width: 120 },
      { prop: 'status', label: '状态', type: 'select', options: [{ label: '启用', value: 1 }, { label: '停用', value: 0 }], default: 1, width: 70 },
      { prop: 'last_login', label: '最后登录', form: false, fmt: r => fmtTime(r.last_login), width: 140 }
    ]
  },

  oplogs: {
    title: '操作日志', api: '/oplogs', writeRoles: ['__none__'], searchPlaceholder: '账号 / 姓名 / 路径 / 模块 / IP', creatable: false, editable: false, deletable: false,
    mobileTitle: r => `${r.real_name} · ${r.module} · ${r.action}`,
    filters: [{ prop: '_range', label: '日期', type: 'daterange' }],
    fields: [
      { prop: 'created_at', label: '时间', fmt: r => fmtTime(r.created_at), width: 140, mobile: 'meta' },
      { prop: 'real_name', label: '操作人', width: 100 },
      { prop: 'username', label: '账号', width: 100 },
      { prop: 'module', label: '模块', width: 100 },
      { prop: 'action', label: '操作', width: 90, tag: true },
      { prop: 'path', label: '请求', minWidth: 220, fmt: r => `${r.method} ${r.path}` },
      { prop: 'ip', label: 'IP', width: 120, mobile: 'meta' },
      { prop: 'status', label: '状态码', width: 76 }
    ]
  }
}

export { money }
