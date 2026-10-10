<template>
  <div class="portal">
    <header class="p-nav">
      <div class="p-wrap nav-in">
        <div class="p-brand"><span class="p-logo">{{ store.settings.logo_text }}</span><div><b>{{ store.settings.village_name }}</b><small>{{ d?.village.town || '数字乡村' }}</small></div></div>
        <nav class="p-links" v-if="!store.isMobile">
          <a href="javascript:;" @click="to('intro')">村情</a><a href="javascript:;" @click="to('news')">快讯</a>
          <a href="javascript:;" @click="to('notice')">公开</a><a href="javascript:;" @click="to('story')">振兴故事</a><a href="javascript:;" @click="to('honor')">红黑榜</a>
        </nav>
        <el-button type="primary" round @click="$router.push(store.token ? '/dashboard' : '/login')">{{ store.token ? '进入工作台' : '工作人员登录' }}</el-button>
      </div>
    </header>

    <section class="p-banner" v-if="d">
      <el-carousel :height="store.isMobile ? '220px' : '380px'" :interval="5000" arrow="hover" indicator-position="outside" v-if="d.banners.length">
        <el-carousel-item v-for="b in d.banners" :key="b.id">
          <div class="bn" :style="bnStyle(b)" @click="b.link && go(b.link)">
            <div class="p-wrap bn-in">
              <h1>{{ b.title }}</h1>
              <p>{{ b.summary }}</p>
            </div>
            <svg class="bn-hill" viewBox="0 0 1200 160" preserveAspectRatio="none"><path d="M0 110 C220 40 420 150 640 90 S960 40 1200 100 V160 H0Z" fill="rgba(255,255,255,.12)"/></svg>
          </div>
        </el-carousel-item>
      </el-carousel>
    </section>

    <div class="p-wrap" v-if="d">
      <section class="p-stats">
        <div><b>{{ d.stats.households }}</b><span>户</span><em>农户</em></div>
        <div><b>{{ d.stats.people }}</b><span>人</span><em>人口</em></div>
        <div><b>{{ d.stats.party }}</b><span>名</span><em>党员</em></div>
        <div><b>{{ d.stats.grids }}</b><span>个</span><em>治理网格</em></div>
        <div><b>{{ d.stats.projects }}</b><span>个</span><em>产业项目</em></div>
        <div><b>{{ wan(d.stats.income) }}</b><span>万元</span><em>本年集体收入</em></div>
      </section>

      <section id="intro" class="p-sec intro">
        <div class="sec-h"><h2>村情概况</h2><span>VILLAGE PROFILE</span></div>
        <div class="intro-box">
          <p class="pre">{{ d.village.intro }}</p>
          <ul>
            <li><em>面积</em>{{ d.village.area }} 平方公里</li><li><em>耕地</em>{{ d.village.farmland }} 亩</li>
            <li><em>地址</em>{{ d.village.address }}</li><li v-if="d.village.phone"><em>电话</em>{{ d.village.phone }}</li>
          </ul>
        </div>
      </section>

      <div class="two">
        <section id="news" class="p-sec">
          <div class="sec-h"><h2>工作快讯</h2><span>NEWS</span></div>
          <div class="news">
            <div v-for="n in d.news" :key="n.id" class="news-item" @click="$router.push('/p/item/' + n.id)">
              <div class="nd"><b>{{ n.date.slice(8) }}</b><small>{{ n.date.slice(0, 7) }}</small></div>
              <div class="nt"><h4>{{ n.title }}</h4><p>{{ n.summary }}</p></div>
            </div>
            <el-empty v-if="!d.news.length" description="暂无快讯" />
          </div>
        </section>
        <section id="notice" class="p-sec">
          <div class="sec-h"><h2>通知公告 · 村务公开</h2><span>PUBLIC AFFAIRS</span></div>
          <el-tabs v-model="cat" class="ntabs">
            <el-tab-pane v-for="c in cats" :key="c" :label="c" :name="c" />
          </el-tabs>
          <ul class="list-plain">
            <li v-for="n in notices" :key="n.id" @click="$router.push('/p/notice/' + n.id)">
              <el-tag v-if="n.top" size="small" type="danger" effect="dark">置顶</el-tag>
              <span class="t">{{ n.title }}</span><span class="d">{{ n.publish_date }}</span>
            </li>
            <el-empty v-if="!notices.length" description="暂无内容" :image-size="60" />
          </ul>
        </section>
      </div>

      <section id="story" class="p-sec">
        <div class="sec-h"><h2>振兴故事</h2><span>STORIES</span></div>
        <div class="stories">
          <div v-for="(s, i) in d.stories" :key="s.id" class="story" @click="$router.push('/p/item/' + s.id)">
            <div class="st-img" :style="s.image ? { backgroundImage: `url(${s.image})` } : { background: grads[i % grads.length] }"><span>{{ s.title.slice(0, 1) }}</span></div>
            <div class="st-b"><h4>{{ s.title }}</h4><p>{{ s.summary }}</p><small>{{ s.date }}</small></div>
          </div>
        </div>
      </section>

      <section class="p-sec" v-if="d.leaders.length">
        <div class="sec-h"><h2>致富带头人</h2><span>LEADERS</span></div>
        <div class="leaders">
          <div v-for="l in d.leaders" :key="l.id" class="leader"><b>{{ l.name }}</b><small>{{ l.industry }} · {{ l.entity }}</small><p>{{ l.story }}</p><em>带动 {{ l.drive }} 户</em></div>
        </div>
      </section>

      <div class="two" id="honor">
        <section class="p-sec">
          <div class="sec-h"><h2>红黑榜</h2><span>HONOR BOARD</span></div>
          <div class="boards">
            <div class="board red"><h4>红榜</h4><div v-for="h in d.honors.filter(x => x.board === '红榜')" :key="h.id" class="bi"><b>{{ h.target }}</b><span>{{ h.category }}</span><p>{{ h.reason }}</p></div></div>
            <div class="board black"><h4>黑榜</h4><div v-for="h in d.honors.filter(x => x.board === '黑榜')" :key="h.id" class="bi"><b>{{ h.target }}</b><span>{{ h.category }}</span><p>{{ h.reason }}</p></div></div>
          </div>
        </section>
        <section class="p-sec">
          <div class="sec-h"><h2>治理积分榜</h2><span>TOP 10</span></div>
          <ol class="rank">
            <li v-for="(r, i) in d.rank" :key="r.household_id"><span class="rn" :class="'r' + (i + 1)">{{ i + 1 }}</span><b>{{ r.household_name }} 户</b><small>{{ r.grid_name }}</small><em>{{ r.total }} 分</em></li>
          </ol>
        </section>
      </div>
    </div>
    <div v-else class="p-wrap" style="padding: 80px 0" v-loading="true" />

    <footer class="p-foot">
      <div class="p-wrap">
        <b>{{ store.settings.village_name }}村民委员会</b>
        <p>{{ d?.village.address }}<span v-if="store.settings.contact_phone"> · 服务电话 {{ store.settings.contact_phone }}</span></p>
        <p class="muted">{{ store.settings.system_title }} · {{ store.settings.footer_text }}</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import http from '@/api/http'
import { store } from '@/store'
import { wan } from '@/utils/format'

const router = useRouter()
const d = ref(null)
const cats = ['通知公告', '党务公开', '村务公开', '财务公开', '村规民约']
const cat = ref('通知公告')
const notices = computed(() => (d.value?.notices || []).filter(n => n.category === cat.value))
const grads = ['linear-gradient(135deg,#0f766e,#84cc16)', 'linear-gradient(135deg,#1d4ed8,#06b6d4)', 'linear-gradient(135deg,#b45309,#f59e0b)', 'linear-gradient(135deg,#9f1239,#f43f5e)']
const bnStyle = b => (b.image ? { backgroundImage: `linear-gradient(90deg,rgba(0,0,0,.45),rgba(0,0,0,.05)),url(${b.image})`, backgroundSize: 'cover', backgroundPosition: 'center' } : { background: b.color || grads[0] })
function to(id) { document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' }) }
function go(link) { if (link.startsWith('#/')) router.push(link.slice(1)); else window.open(link, '_blank', 'noopener') }
onMounted(async () => { d.value = await http.get('/public/portal') })
</script>

<style scoped lang="scss">
.portal { background: #f6f8f7; min-height: 100vh; }
.p-wrap { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.p-nav { position: sticky; top: 0; z-index: 10; background: rgba(255,255,255,.92); backdrop-filter: blur(10px); border-bottom: 1px solid var(--vm-border); }
.nav-in { height: 64px; display: flex; align-items: center; gap: 24px; }
.p-brand { display: flex; align-items: center; gap: 10px; flex: 1; }
.p-brand b { display: block; font-size: 18px; letter-spacing: 2px; }
.p-brand small { color: var(--vm-muted); font-size: 12px; }
.p-logo { width: 40px; height: 40px; border-radius: 12px; display: grid; place-items: center; color: #fff; font-weight: 700; font-size: 19px; background: linear-gradient(135deg, var(--vm-primary), #84cc16); }
.p-links { display: flex; gap: 22px; }
.p-links a { color: var(--vm-text); font-size: 15px; }
.p-links a:hover { color: var(--vm-primary); }
.bn { height: 100%; position: relative; color: #fff; display: flex; align-items: center; cursor: default; overflow: hidden; }
.bn-in { position: relative; z-index: 1; width: 100%; }
.bn h1 { font-size: 40px; margin: 0 0 12px; letter-spacing: 2px; text-shadow: 0 2px 12px rgba(0,0,0,.2); }
.bn p { font-size: 18px; opacity: .92; margin: 0; }
.bn-hill { position: absolute; left: 0; right: 0; bottom: 0; width: 100%; height: 110px; }
.p-banner :deep(.el-carousel__indicators--outside) { margin-top: -6px; }
.p-stats { display: grid; grid-template-columns: repeat(6, 1fr); background: #fff; border-radius: 16px; box-shadow: var(--vm-shadow); margin-top: 10px; padding: 18px 0; }
.p-stats div { display: flex; flex-direction: column; align-items: center; border-right: 1px solid var(--vm-border); }
.p-stats div:last-child { border-right: none; }
.p-stats b { font-size: 28px; color: var(--vm-primary); font-variant-numeric: tabular-nums; }
.p-stats span { font-size: 12px; color: var(--vm-muted); margin-top: -4px; }
.p-stats em { font-style: normal; font-size: 13px; margin-top: 4px; }
.p-sec { margin-top: 30px; }
.sec-h { display: flex; align-items: baseline; gap: 10px; margin-bottom: 14px; }
.sec-h h2 { margin: 0; font-size: 22px; position: relative; padding-left: 12px; }
.sec-h h2::before { content: ''; position: absolute; left: 0; top: 5px; bottom: 5px; width: 4px; border-radius: 2px; background: var(--vm-primary); }
.sec-h span { color: #b6c4c0; font-size: 12px; letter-spacing: 2px; }
.intro-box { background: #fff; border-radius: 16px; padding: 22px; box-shadow: var(--vm-shadow); display: grid; grid-template-columns: 2fr 1fr; gap: 20px; }
.intro-box p { margin: 0; color: #3d4f4b; text-indent: 2em; }
.intro-box ul { list-style: none; margin: 0; padding: 0 0 0 20px; border-left: 1px dashed var(--vm-border); display: grid; gap: 10px; font-size: 14px; }
.intro-box em { font-style: normal; color: var(--vm-muted); margin-right: 10px; }
.two { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
.two > .p-sec { background: #fff; border-radius: 16px; padding: 20px; box-shadow: var(--vm-shadow); }
.news-item { display: flex; gap: 14px; padding: 12px 0; border-bottom: 1px dashed var(--vm-border); cursor: pointer; }
.news-item:last-child { border-bottom: none; }
.nd { width: 58px; flex-shrink: 0; text-align: center; background: var(--vm-primary-soft); border-radius: 10px; padding: 6px 0; }
.nd b { display: block; font-size: 22px; color: var(--vm-primary); }
.nd small { font-size: 11px; color: var(--vm-muted); }
.nt h4 { margin: 2px 0 4px; font-size: 15px; }
.nt p { margin: 0; font-size: 13px; color: var(--vm-muted); line-height: 1.6; }
.news-item:hover h4 { color: var(--vm-primary); }
.ntabs :deep(.el-tabs__header) { margin-bottom: 4px; }
.stories { display: grid; grid-template-columns: repeat(3, 1fr); gap: 18px; }
.story { background: #fff; border-radius: 16px; overflow: hidden; box-shadow: var(--vm-shadow); cursor: pointer; transition: transform .2s; }
.story:hover { transform: translateY(-4px); }
.st-img { height: 150px; background-size: cover; background-position: center; display: flex; align-items: flex-end; padding: 14px; }
.st-img span { font-size: 54px; font-weight: 800; color: rgba(255,255,255,.35); line-height: 1; }
.st-b { padding: 14px 16px 16px; }
.st-b h4 { margin: 0 0 6px; font-size: 16px; }
.st-b p { margin: 0 0 8px; color: var(--vm-muted); font-size: 13px; line-height: 1.6; }
.st-b small { color: #9fb0ab; }
.leaders { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 14px; }
.leader { background: #fff; border-radius: 14px; padding: 16px; box-shadow: var(--vm-shadow); border-top: 3px solid #f59e0b; }
.leader b { font-size: 16px; display: block; }
.leader small { color: var(--vm-primary); }
.leader p { font-size: 13px; color: #3d4f4b; line-height: 1.6; margin: 8px 0; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }
.leader em { font-style: normal; font-size: 12px; color: var(--vm-muted); }
.boards { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.board { border-radius: 12px; padding: 12px; }
.board h4 { margin: 0 0 8px; text-align: center; letter-spacing: 6px; color: #fff; border-radius: 8px; padding: 4px 0; }
.board.red { background: #fff5f5; } .board.red h4 { background: #c8102e; }
.board.black { background: #f4f5f6; } .board.black h4 { background: #374151; }
.bi { padding: 8px 0; border-bottom: 1px dashed #e5e7eb; }
.bi:last-child { border-bottom: none; }
.bi b { font-size: 14px; } .bi span { font-size: 12px; color: var(--vm-muted); margin-left: 6px; }
.bi p { margin: 4px 0 0; font-size: 12px; color: #4b5563; line-height: 1.5; }
.rank { list-style: none; margin: 0; padding: 0; }
.rank li { display: flex; align-items: center; gap: 10px; padding: 9px 0; border-bottom: 1px dashed var(--vm-border); }
.rank b { flex-shrink: 0; } .rank small { color: var(--vm-muted); flex: 1; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.rank em { font-style: normal; color: var(--vm-primary); font-weight: 700; }
.rn { width: 24px; height: 24px; border-radius: 50%; display: grid; place-items: center; background: #eef2f1; font-size: 12px; font-weight: 700; flex-shrink: 0; }
.rn.r1 { background: #fbbf24; color: #fff; } .rn.r2 { background: #94a3b8; color: #fff; } .rn.r3 { background: #d97706; color: #fff; }
.p-foot { margin-top: 40px; background: var(--vm-primary-deep); color: rgba(255,255,255,.85); padding: 28px 0; text-align: center; font-size: 13px; }
.p-foot p { margin: 6px 0 0; } .p-foot .muted { color: rgba(255,255,255,.5); }
@media (max-width: 860px) {
  .nav-in { gap: 10px; height: 56px; }
  .bn h1 { font-size: 24px; } .bn p { font-size: 14px; }
  .p-stats { grid-template-columns: repeat(3, 1fr); row-gap: 14px; }
  .p-stats div:nth-child(3) { border-right: none; }
  .p-stats b { font-size: 22px; }
  .intro-box, .two, .stories, .boards { grid-template-columns: 1fr; }
  .intro-box ul { border-left: none; padding: 0; }
  .p-wrap { padding: 0 12px; }
}
</style>
