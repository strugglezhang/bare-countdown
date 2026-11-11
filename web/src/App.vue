<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Icon from './Icon.vue'
import Counter from './Counter.vue'
import SceneArt from './SceneArt.vue'
import CalendarView from './CalendarView.vue'
import ImmersiveStage from './ImmersiveStage.vue'
import ChoiceSelect from './ChoiceSelect.vue'
import BrandMark from './BrandMark.vue'
import FestivalArt from './FestivalArt.vue'
import {holidays,holidayById,nextHolidayDate,holidayNeedsRefresh} from './holidays.js'
import {effects} from './effects.js'
import {pickFeatured} from './featured.js'
import { scenes } from './scenes.js'
import { timeInfo, dateLabel, newDraft, localInput, kindLabels, themes, styles, recurrenceLabels } from './time.js'

const events = ref([]), now = ref(Date.now()), offset = ref(0), loading = ref(true), loadError = ref('')
const search = ref(''), category = ref('全部'), status = ref('all'), sort = ref('nearest'), view = ref('grid'), pinnedOnly = ref(false)
const editor = ref(null), deleteDialog = ref(null), focusDialog = ref(null), searchInput = ref(null), importInput = ref(null)
const draft = ref(newDraft()), editID = ref(null), editingDateChanged = ref(false), saving = ref(false), formError = ref(''), pendingDelete = ref(null), focused = ref(null)
const notifications = ref([]), soundOn = ref(localStorage.getItem('timecapsule-sound') === 'true'), dark = ref(localStorage.getItem('timecapsule-dark') === 'true')
const announced = new Set(), rolling = new Set()
const browserTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Shanghai'
const savedScene = localStorage.getItem('timecapsule-scene')
const scene = ref(scenes.some(item=>item.id===savedScene)?savedScene:'journey')
const currentScene = computed(()=>scenes.find(item=>item.id===scene.value))
const savedEffect=localStorage.getItem('timecapsule-effect')
const effectMode=ref(effects.some(e=>e.value===savedEffect)?savedEffect:localStorage.getItem('timecapsule-motion')==='off'||window.matchMedia('(prefers-reduced-motion: reduce)').matches?'still':'flow')
const lastEffect=ref(localStorage.getItem('timecapsule-last-effect')||'flow')
const motionOn=computed({get:()=>effectMode.value!=='still',set:value=>effectMode.value=value?(effects.some(e=>e.value===lastEffect.value&&e.value!=='still')?lastEffect.value:'flow'):'still'})
watch(effectMode,value=>{document.documentElement.dataset.effect=value;localStorage.setItem('timecapsule-effect',value);if(value!=='still'){lastEffect.value=value;localStorage.setItem('timecapsule-last-effect',value)}},{immediate:true})
const sortChoices=[{value:'nearest',label:'距离最近'},{value:'newest',label:'最新添加'},{value:'title',label:'标题排序'}]
const celebrating = ref(false)
const stage = ref(null), quickSaving = ref(false)
let interval, refreshInterval, celebrationTimer, toastTimers = [], audioContext

watch(scene, value=>{document.documentElement.dataset.scene=value;localStorage.setItem('timecapsule-scene',value)}, {immediate:true})
watch(motionOn, value=>{document.documentElement.dataset.motion=value?'flow':'still';localStorage.setItem('timecapsule-motion',value?'on':'off')}, {immediate:true})

watch(dark, value => { document.documentElement.dataset.theme = value ? 'dark' : 'light'; localStorage.setItem('timecapsule-dark', String(value)) }, { immediate: true })
watch(soundOn, value => localStorage.setItem('timecapsule-sound', String(value)))
const holidaySearch=ref(''),holidayGroup=ref('all')
const visibleHolidays=computed(()=>holidays.filter(h=>(holidayGroup.value==='all'||h.group===holidayGroup.value)&&(!holidaySearch.value.trim()||`${h.label} ${h.description} ${h.dateNote||''}`.includes(holidaySearch.value.trim()))))
const holidayDateHint=computed(()=>{const h=holidayById(draft.value.holiday);if(!h)return '';const detail=h.dateNote||'';return detail+(holidayNeedsRefresh(h.value)?((h.lunar||h.table)&&!draft.value.date?' 请手动填写日期，内置历表范围为 2026–2035 年。':' 已填入最近一次日期；每年重新选择预设更新日期，避免按固定公历日期重复。'):' 已填入最近一次日期，默认每年重复；可手动调整。')})
const categories = computed(() => [...new Set(['生活', '工作', '纪念日', '专注', '节日', ...events.value.map(e => e.category)])])
const repeatChoices=[{value:'none',label:'不重复',description:'留住这一次特别的时刻'},{value:'daily',label:'每天',description:'每天同一时刻，继续期待'},{value:'weekly',label:'每周',description:'每周的这一天，给自己一个约定'},{value:'yearly',label:'每年',description:'生日、周年纪念，年年如约'}]
const infos = computed(() => Object.fromEntries(events.value.map(e => [e.id, timeInfo(e, now.value)])))
const activeCount = computed(() => events.value.filter(e => !infos.value[e.id].reached && !infos.value[e.id].paused).length)
const weekCount = computed(() => events.value.filter(e => !infos.value[e.id].elapsed && !infos.value[e.id].paused && infos.value[e.id].seconds > 0 && infos.value[e.id].seconds <= 604800).length)
const hasExamples = computed(() => events.value.some(e => e.demo))
const featured = computed(() => pickFeatured(events.value,now.value))
const filtered = computed(() => events.value.filter(e => {
  const i = infos.value[e.id]
  return (category.value === '全部' || e.category === category.value) && (!pinnedOnly.value || e.pinned) && (!search.value || `${e.title} ${e.description} ${e.category}`.toLowerCase().includes(search.value.toLowerCase())) && (status.value === 'all' || status.value === 'active' && !i.reached && !i.paused || status.value === 'paused' && i.paused || status.value === 'ended' && i.reached)
}).sort((a,b) => Number(b.pinned) - Number(a.pinned) || (sort.value === 'newest' ? Date.parse(b.created_at)-Date.parse(a.created_at) : sort.value === 'title' ? a.title.localeCompare(b.title,'zh-CN') : infos.value[a.id].seconds-infos.value[b.id].seconds)))
const today = computed(() => new Intl.DateTimeFormat('zh-CN', { month: 'long', day: 'numeric', weekday: 'long' }).format(now.value))
const preview = computed(() => ({ ...draft.value, target_at: new Date(draft.value.date || Date.now()).toISOString(), effective_target_at: draft.value.kind === 'timer' ? new Date(now.value + Math.max(1, Number(draft.value.duration_minutes)) * 60000).toISOString() : '', created_at: new Date(now.value).toISOString(), duration_seconds: Number(draft.value.duration_minutes)*60, paused_remaining: undefined }))

function toast(text, type = 'success') {
  const id = crypto.randomUUID(); notifications.value.push({ id, text, type }); notifications.value=notifications.value.slice(-3)
  toastTimers.push(setTimeout(() => { notifications.value = notifications.value.filter(n => n.id !== id) }, 4800))
}
async function api(path, options = {}) {
  const response = await fetch(`/api${path}`, { ...options, headers: { 'Content-Type': 'application/json', ...options.headers } })
  if (!response.ok) { let message = `请求失败（${response.status}）`; try { message = (await response.json()).error || message } catch {} throw new Error(message) }
  return response.status === 204 ? null : response.json()
}
async function load(first = false) {
  try {
    const data = await api('/events'); events.value = data.events
    offset.value = Date.parse(data.server_time) - Date.now(); now.value = Date.now() + offset.value
    loadError.value = ''; return true
  } catch (error) { if (first) loadError.value = error.message; else toast(error.message, 'error'); return false }
  finally { loading.value = false }
}
function replace(item) { const index=events.value.findIndex(e=>e.id===item.id); if(index<0) events.value.unshift(item); else events.value[index]=item; if(focused.value?.id===item.id) focused.value=item }
function openEditor(kind = 'countdown', event = null) {
  holidaySearch.value='';holidayGroup.value='all'
  editID.value = event?.id || null; formError.value = ''; editingDateChanged.value = false
  draft.value = event ? { ...event, date: localInput(event.target_at), duration_minutes: event.duration_seconds / 60 || 25 } : newDraft(kind)
  if(!event&&kind==='festival')applyHoliday('spring')
  editor.value.showModal(); nextTick(() => editor.value.querySelector('input')?.focus())
}
function applyHoliday(value){const h=holidayById(value);if(!h)return;const date=nextHolidayDate(value,now.value);Object.assign(draft.value,{holiday:value,category:'节日',theme:h.theme,title:h.title,description:h.message,recurrence:holidayNeedsRefresh(value)?'none':'yearly'});draft.value.date=date;editingDateChanged.value=true}
function setKind(kind) { draft.value.kind = kind;if(kind==='festival'){applyHoliday(draft.value.holiday||'spring');return}draft.value.holiday='';if(draft.value.category==='节日')draft.value.category=kind==='anniversary'?'纪念日':'生活'; if(kind==='timer'){draft.value.recurrence='none';draft.value.style='ring';draft.value.category='专注'} }
async function save() {
  formError.value = ''; saving.value = true
  try {
    const d = draft.value
    const target = editID.value && !editingDateChanged.value ? events.value.find(e=>e.id===editID.value)?.target_at : d.kind === 'timer' ? '' : new Date(d.date).toISOString()
    const data = { title:d.title, description:d.description, category:d.category, kind:d.kind, theme:d.theme, style:d.style, holiday:d.kind==='festival'?d.holiday:'', target_at:target || '', timezone: editID.value && !editingDateChanged.value ? d.timezone : Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Shanghai', recurrence:d.kind==='timer'?'none':d.recurrence, duration_seconds:Math.round(Number(d.duration_minutes)*60), pinned:d.pinned }
    const item=await api(editID.value ? `/events/${editID.value}` : '/events',{method:editID.value?'PUT':'POST',body:JSON.stringify(data)})
    replace(item); editor.value.close(); toast(editID.value ? '这个时刻，已经更新' : '新的期待，已放进时光舱')
  } catch (error) { formError.value = error.message }
  finally { saving.value = false }
}
async function action(event, action) { try {replace(await api(`/events/${event.id}/action`,{method:'POST',body:JSON.stringify({action})}));if(action==='pin')toast(event.pinned?'已取消置顶':'已置顶这个时刻')}catch(error){toast(error.message,'error')} }
function askDelete(event) { pendingDelete.value = event; deleteDialog.value.showModal() }
async function confirmDelete() {
  if(saving.value)return;saving.value=true
  try {
    if(pendingDelete.value?.id){await api(`/events/${pendingDelete.value.id}`,{method:'DELETE'});events.value=events.value.filter(e=>e.id!==pendingDelete.value.id)}
    else {await api('/examples',{method:'DELETE'});events.value=events.value.filter(e=>!e.demo)}
    deleteDialog.value.close();toast('已删除，给新的期待留点位置')
  }catch(error){toast(error.message,'error')}finally{saving.value=false}
}
function clearExamples() {pendingDelete.value=null;deleteDialog.value.showModal()}
function focus(event){focused.value=event;focusDialog.value.showModal()}
async function fullscreen(){try{if(document.fullscreenElement)await document.exitFullscreen();else await focusDialog.value.requestFullscreen()}catch{toast('当前浏览器不支持全屏，已保持专注视图','error')}}
function closeFocus(){if(document.fullscreenElement)document.exitFullscreen().catch(()=>{});focusDialog.value.close()}
function selectCategory(value){category.value=value;pinnedOnly.value=false}
function exportData(){const blob=new Blob([JSON.stringify({version:1,exported_at:new Date().toISOString(),events:events.value},null,2)],{type:'application/json'});const url=URL.createObjectURL(blob);const a=document.createElement('a');a.href=url;a.download=`时光舱-${new Date().toISOString().slice(0,10)}.json`;a.click();setTimeout(()=>URL.revokeObjectURL(url),1000);toast('备份已导出，期待也有了备份')}
async function importData(event){const file=event.target.files?.[0];if(!file)return;try{if(file.size>2*1024*1024)throw Error('备份文件不能超过 2 MB');const body=JSON.parse(await file.text());const result=await api('/events/import',{method:'POST',body:JSON.stringify(body)});await load();toast(`已导入 ${result.imported} 个时刻`)}catch(error){toast(error.message==='Unexpected end of JSON input'?'文件不是有效的 JSON 备份':error.message,'error')}finally{event.target.value=''}}
function enableSound(){soundOn.value=!soundOn.value;if(soundOn.value){try{audioContext ||= new (window.AudioContext||window.webkitAudioContext)();audioContext.resume()}catch{soundOn.value=false;toast('当前浏览器不支持声音提醒','error')}}}
function chime(){if(!soundOn.value||!audioContext)return;for(const [i,f] of [523.25,659.25,783.99].entries()){const oscillator=audioContext.createOscillator(),gain=audioContext.createGain(),t=audioContext.currentTime+i*.17;oscillator.frequency.value=f;gain.gain.setValueAtTime(.0001,t);gain.gain.exponentialRampToValueAtTime(.15,t+.02);gain.gain.exponentialRampToValueAtTime(.0001,t+.7);oscillator.connect(gain);gain.connect(audioContext.destination);oscillator.start(t);oscillator.stop(t+.75)}}
function celebrate(){if(!motionOn.value)return;celebrating.value=true;clearTimeout(celebrationTimer);celebrationTimer=setTimeout(()=>celebrating.value=false,3000)}
async function quickFocus(minutes){
  if(quickSaving.value)return;quickSaving.value=true
  try{
    const item=await api('/events',{method:'POST',body:JSON.stringify({title:`${minutes} 分钟，认真做一件事`,description:'把注意力留给眼前，把时间留给自己。',category:'专注',kind:'timer',theme:'mint',style:'ring',target_at:'',timezone:browserTimezone,recurrence:'none',duration_seconds:minutes*60,pinned:false})})
    replace(item);toast(`${minutes} 分钟专注，已经开始`);focus(item)
  }catch(error){toast(error.message,'error')}finally{quickSaving.value=false}
}
function tick(){
  const previous=now.value;now.value=Date.now()+offset.value
  for(const e of events.value){if(e.paused_remaining!=null)continue;const target=Date.parse(e.effective_target_at||e.target_at);const key=`${e.id}:${target}`
    if(target>previous&&target<=now.value&&e.kind!=='anniversary'&&!announced.has(key)){announced.add(key);toast(`「${e.title}」时间到了！`);chime();celebrate()}
    if(e.recurrence!=='none'&&target<=now.value&&!rolling.has(key)){rolling.add(key);load().then(ok=>{if(!ok)rolling.delete(key)})}
  }
}
function keydown(e){if((e.metaKey||e.ctrlKey)&&e.key==='k'){e.preventDefault();searchInput.value?.focus()}}
onMounted(()=>{load(true);interval=setInterval(tick,1000);refreshInterval=setInterval(()=>load(),60000);window.addEventListener('keydown',keydown)})
onUnmounted(()=>{clearInterval(interval);clearInterval(refreshInterval);clearTimeout(celebrationTimer);toastTimers.forEach(clearTimeout);window.removeEventListener('keydown',keydown);audioContext?.close()})
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <a class="brand" href="/" aria-label="时光舱首页"><span class="brand-mark"><BrandMark/></span><span class="brand-wordmark"><strong>时光舱</strong><small>TIME CAPSULE</small></span></a>
      <p class="nav-label">我的时光</p>
      <nav class="main-nav" aria-label="倒计时分类">
        <button :class="{active:category==='全部'&&!pinnedOnly}" @click="selectCategory('全部')"><Icon name="grid"/><span>全部时刻</span><em>{{events.length}}</em></button>
        <button :class="{active:pinnedOnly}" @click="category='全部';pinnedOnly=true"><Icon name="pin"/><span>特别期待</span><em>{{events.filter(e=>e.pinned).length}}</em></button>
        <div class="nav-divider"/>
        <p class="nav-label">按生活分类</p>
        <button v-for="(item,index) in categories" :key="item" :class="{active:category===item&&!pinnedOnly}" @click="selectCategory(item)"><span class="nav-category-dot" :style="{'--dot':['#9a83df','#deb16a','#e392b1','#6bb5a5'][index%4]}"/><span>{{item}}</span><em>{{events.filter(e=>e.category===item).length}}</em></button>
      </nav>
      <div class="sidebar-bottom">
        <div class="little-note"><span class="note-spark">✦</span><p>让每一份期待<br/>都有它的倒计时。</p><span class="note-orbit"/></div>
        <button class="sidebar-tool" @click="exportData"><Icon name="download" :size="17"/>导出时光备份</button>
        <button class="sidebar-tool" @click="importInput.click()"><Icon name="upload" :size="17"/>导入我的时刻</button>
        <input ref="importInput" type="file" accept="application/json,.json" class="sr-only" aria-label="导入 JSON 备份" @change="importData"/>
        <div class="sidebar-footer"><span class="live-dot"/>好好生活，慢慢期待<span>v1.0</span></div>
      </div>
    </aside>

    <div class="main-shell">
      <header class="topbar"><div class="breadcrumb">我的时光 <span>/</span> <strong>{{pinnedOnly?'特别期待':category==='全部'?'全部时刻':category}}</strong></div><div class="topbar-actions"><span class="date-today">{{today}}</span><button class="icon-button" :class="{enabled:soundOn}" :aria-label="soundOn?'关闭到期声音':'开启到期声音'" :title="soundOn?'到期声音已开启':'开启到期声音（页面打开时有效）'" @click="enableSound"><Icon :name="soundOn?'volume':'bell'"/></button><button class="icon-button" :aria-label="dark?'切换浅色模式':'切换深色模式'" @click="dark=!dark"><Icon :name="dark?'sun':'moon'"/></button><span class="avatar">时</span></div></header>
      <main>
        <section class="page-heading"><div><div class="eyebrow"><span/> LIVE FREELY. LOVE DEEPLY.</div><h1>让期待，有迹可循<span> ✦</span></h1><p>让我们红尘作伴，活得潇潇洒洒。<span class="heading-aside">把值得的日子，过成喜欢的模样。</span></p></div><button class="primary-button" @click="openEditor()"><Icon name="plus"/>新建倒计时</button></section>

        <section class="hero" :class="featured?.holiday?`festival-${featured.holiday}`:''" aria-label="精选倒计时"><FestivalArt :holiday="featured?.holiday"/>
          <SceneArt :scene="scene" :now="now" :effect="effectMode"/><div class="hero-shade"/><div class="hero-grain"/><span class="hero-star star-two">✧</span>
          <div class="hero-copy"><span class="hero-tag"><span/> {{featured?(featured.pinned?'置顶的特别时光':'下一份特别期待'):'写给未来的自己'}}</span><h2>{{featured?.title || '把日子过成，值得期待的样子。'}}</h2><p>{{featured?.description || '时间会向前走，你的期待也会如约抵达。'}}</p><button v-if="featured" class="hero-link" @click="focus(featured)">离美好，又近了一点 <Icon name="arrow" :size="16"/></button><button v-else class="hero-link" @click="openEditor()">开启第一份期待 <Icon name="arrow" :size="16"/></button></div>
          <div v-if="featured" class="hero-clock"><span class="hero-count-label">{{featured.kind==='anniversary'&&infos[featured.id]?.elapsed?'EVERY DAY TOGETHER COUNTS':'COUNTING DOWN TO SOMETHING GOOD'}}</span><Counter :event="featured" :now="now"/><span class="hero-clock-date"><Icon name="calendar" :size="13"/>{{dateLabel(featured)}}</span></div>
          <div v-else class="hero-empty-clock"><Icon name="hourglass" :size="100"/></div>
          <div class="hero-signature"><span class="signature-line"/>{{currentScene.caption}}<span>{{currentScene.english}}</span></div>
        </section>

        <section class="mood-bar" aria-label="时光氛围">
          <div class="mood-heading"><Icon name="sparkles" :size="18"/><div><strong>今天，想去哪里？</strong><span>给此刻，换一种心情</span></div></div>
          <div class="mood-options"><button v-for="item in scenes" :key="item.id" :aria-label="item.label" :aria-pressed="scene===item.id" :class="{selected:scene===item.id}" :style="{'--mood-color':item.color}" @click="scene=item.id"><span class="mood-swatch"><Icon :name="item.icon" :size="15"/></span>{{item.label}}<span v-if="scene===item.id" class="mood-selected-dot"/></button></div>
          <div class="motion-controls"><ChoiceSelect v-model="effectMode" :options="effects" label="背景动效" icon="sparkles" compact/><button class="motion-toggle" :aria-label="motionOn?'暂停氛围动效':'开启氛围动效'" :aria-pressed="motionOn" @click="motionOn=!motionOn"><Icon :name="motionOn?'pause':'play'" :size="14"/></button></div>
          <button class="open-stage-button" @click="stage.open()"><Icon name="expand" :size="15"/>打开沉浸大屏<Icon name="arrow" :size="14"/></button>
        </section>

        <section class="quick-focus-bar" aria-label="快速专注"><div><span class="quick-focus-icon"><Icon name="coffee" :size="19"/></span><span><strong>此刻，专心做一件事。</strong><small>选一个时长，立即开始</small></span></div><div class="quick-focus-buttons"><button v-for="minute in [5,15,25,45]" :key="minute" :disabled="quickSaving" :aria-label="`创建 ${minute} 分钟专注`" @click="quickFocus(minute)"><strong>{{minute}}</strong><span>分钟</span><Icon name="play" :size="12"/></button></div></section>

        <section class="stats-row" aria-label="时光统计"><div><span class="stat-icon violet"><Icon name="calendar"/></span><span><small>珍藏的时刻</small><strong>{{events.length}}<em>个</em></strong></span><span class="stat-caption">每一个都值得</span></div><div><span class="stat-icon mint"><Icon name="clock"/></span><span><small>正在期待</small><strong>{{activeCount}}<em>个</em></strong></span><span class="stat-caption"><span class="live-dot"/>时间在靠近</span></div><div><span class="stat-icon rose"><Icon name="sparkles"/></span><span><small>一周内抵达</small><strong>{{weekCount}}<em>个</em></strong></span><span class="stat-caption">好事即将发生</span></div></section>

        <section class="moments-section">
          <div class="section-heading"><div><h2>{{pinnedOnly?'特别期待':category==='全部'?'我的时光清单':`${category}里的小期待`}}<span>{{filtered.length}}</span></h2><p>把喜欢的生活，一点点安排进未来。</p></div><div class="view-switch" aria-label="显示布局"><button :class="{selected:view==='grid'}" :aria-pressed="view==='grid'" aria-label="卡片视图" @click="view='grid'"><Icon name="grid" :size="17"/></button><button :class="{selected:view==='list'}" :aria-pressed="view==='list'" aria-label="列表视图" @click="view='list'"><Icon name="list" :size="19"/></button><button :class="{selected:view==='calendar'}" :aria-pressed="view==='calendar'" aria-label="日历视图" @click="view='calendar'"><Icon name="calendar" :size="18"/></button></div></div>
          <div class="filter-toolbar"><div class="status-filters"><button v-for="item in [{id:'all',label:'全部'},{id:'active',label:'进行中'},{id:'paused',label:'已暂停'},{id:'ended',label:'已到期'}]" :key="item.id" :class="{selected:status===item.id}" @click="status=item.id">{{item.label}}</button></div><div class="filter-right"><label class="search-box"><Icon name="search" :size="16"/><input ref="searchInput" v-model="search" aria-label="搜索时刻" placeholder="搜索你的期待…"/><kbd>⌘ K</kbd></label><ChoiceSelect v-model="sort" :options="sortChoices" label="排序方式" icon="line" compact class="sort-control"/></div></div>
          <div v-if="hasExamples" class="demo-hint"><Icon name="sparkles" :size="14"/><span>先用几个示例，感受时间的不同模样。</span><button @click="clearExamples">清空示例</button></div>
          <div v-if="loading" class="empty-state"><span class="loading-spinner"/>正在打开你的时光舱…</div>
          <div v-else-if="loadError" class="empty-state"><Icon name="cloud" :size="40"/><h3>时光舱暂时没有连接上</h3><p>{{loadError}}</p><button class="primary-button" @click="loading=true;load(true)">重新连接</button></div>
          <div v-else-if="!filtered.length" class="empty-state"><span class="empty-art"><Icon name="hourglass" :size="36"/></span><h3>{{events.length?'还没有符合条件的时刻':'未来的美好，从这里开始'}}</h3><p>{{events.length?'换个关键词或筛选条件，找找你的期待。':'旅行、生日、重要的截止日期……把第一份期待放进来吧。'}}</p><button v-if="!events.length" class="primary-button" @click="openEditor()"><Icon name="plus"/>添加第一个倒计时</button><button v-else class="secondary-button" @click="search='';status='all';category='全部';pinnedOnly=false">清除筛选</button></div>
          <CalendarView v-else-if="view==='calendar'" :events="filtered" :now="now" @focus="focus"/>
          <TransitionGroup v-else name="cards" move-class="cards-static" tag="div" class="cards-grid" :class="{'list-view':view==='list'}">
            <article v-for="event in filtered" :key="event.id" class="moment-card" :class="[`theme-${event.theme}`,event.holiday?`festival-${event.holiday}`:'',{paused:infos[event.id].paused,arrived:infos[event.id].reached}]">
              <FestivalArt :holiday="event.holiday"/><div class="card-decoration"/><div class="card-heading"><span class="category-badge"><Icon :name="event.kind==='timer'?'coffee':event.kind==='anniversary'?'heart':event.category==='工作'?'briefcase':event.category==='生活'?'plane':'calendar'" :size="14"/>{{event.category}}</span><div class="card-top-actions"><span v-if="event.demo" class="demo-label">示例</span><button class="card-icon-button" :class="{pinned:event.pinned}" :aria-label="event.pinned?`取消置顶${event.title}`:`置顶${event.title}`" @click="action(event,'pin')"><Icon name="pin" :size="16"/></button></div></div>
              <div class="card-title"><h3>{{event.title}}</h3><p>{{infos[event.id].status}}<span v-if="event.recurrence!=='none'"> · {{recurrenceLabels[event.recurrence]}}重复</span></p></div>
              <Counter :event="event" :now="now"/>
              <div class="card-note">{{event.description || '为未来，留一点值得期待的事。'}}</div>
              <div class="card-footer"><span class="card-date"><Icon name="calendar" :size="13"/>{{event.kind==='timer'?`${Math.round(event.duration_seconds/60*10)/10} 分钟 · ${kindLabels[event.kind]}`:dateLabel(event).split(' ')[0]}}</span><div class="card-actions"><button v-if="event.kind==='timer'&&!infos[event.id].reached" class="card-icon-button" :aria-label="infos[event.id].paused?`继续${event.title}`:`暂停${event.title}`" @click="action(event,infos[event.id].paused?'resume':'pause')"><Icon :name="infos[event.id].paused?'play':'pause'" :size="15"/></button><button v-if="event.kind==='timer'" class="card-icon-button" :aria-label="`重置${event.title}`" @click="action(event,'restart')"><Icon name="reset" :size="15"/></button><button class="card-icon-button" :aria-label="`专注查看${event.title}`" @click="focus(event)"><Icon name="expand" :size="15"/></button><button class="card-icon-button" :aria-label="`编辑${event.title}`" @click="openEditor(event.kind,event)"><Icon name="edit" :size="15"/></button><button class="card-icon-button delete-button" :aria-label="`删除${event.title}`" @click="askDelete(event)"><Icon name="trash" :size="15"/></button></div></div>
            </article>
            <button key="add-card" class="add-card" @click="openEditor()"><span><Icon name="plus" :size="27"/></span><strong>下一份期待，是什么？</strong><p>给未来的自己，留一个小惊喜</p><em>添加一个时刻 <Icon name="arrow" :size="15"/></em></button>
          </TransitionGroup>
        </section>
        <section class="inspiration"><div><span class="inspiration-star">✧</span><div><strong>不知道从哪里开始？</strong><p>生活的每一面，都有值得期待的瞬间。</p></div></div><div class="template-actions"><button @click="openEditor('countdown')"><Icon name="calendar" :size="17"/>重要日子<Icon name="plus" :size="13"/></button><button @click="openEditor('anniversary')"><Icon name="heart" :size="17"/>纪念时刻<Icon name="plus" :size="13"/></button><button @click="openEditor('festival')"><Icon name="sparkles" :size="17"/>节日将至<Icon name="plus" :size="13"/></button><button @click="openEditor('timer')"><Icon name="coffee" :size="17"/>专注一会<Icon name="plus" :size="13"/></button></div></section>
        <footer class="page-footer"><span>✦ 日子不必赶路，值得的美好会慢慢抵达。</span><span>MADE FOR THE MOMENTS THAT MATTER</span></footer>
      </main>
    </div>

    <dialog ref="editor" class="editor-dialog" aria-labelledby="editor-title" @cancel.prevent="!saving&&editor.close()">
      <form @submit.prevent="save">
        <div class="dialog-header"><div><span class="eyebrow">MAKE ROOM FOR SOMETHING GOOD</span><h2 id="editor-title">{{editID?'编辑这个时刻':'收藏一份新的期待'}} <span>✦</span></h2></div><button type="button" class="icon-button" aria-label="关闭编辑" :disabled="saving" @click="editor.close()"><Icon name="close"/></button></div>
        <div class="editor-layout"><div class="editor-fields"><div class="kind-switch"><button v-for="(label,kind) in kindLabels" :key="kind" type="button" :class="{selected:draft.kind===kind}" @click="setKind(kind)">{{label}}</button></div>
          <div v-if="draft.kind==='festival'" class="festival-picker"><span class="field-label">选一个节日，让期待有自己的颜色</span><div class="festival-tools"><div class="festival-tabs" aria-label="节日地区"><button v-for="g in [{id:'all',label:'全部'},{id:'china',label:'中国节日'},{id:'world',label:'世界节庆'}]" :key="g.id" type="button" :class="{selected:holidayGroup===g.id}" :aria-pressed="holidayGroup===g.id" @click="holidayGroup=g.id">{{g.label}}</button></div><label class="festival-search"><Icon name="search" :size="14"/><input v-model="holidaySearch" aria-label="搜索节日" placeholder="找节日、配色、城市…"/></label></div><div class="festival-result-count">{{visibleHolidays.length}} / {{holidays.length}} 个节日 · 当前：{{holidayById(draft.holiday)?.label}}</div><div class="festival-choices"><p v-if="!visibleHolidays.length" class="festival-empty">没有找到这个节日，换个关键词试试</p><button v-for="h in visibleHolidays" :key="h.value" type="button" :class="[`festival-${h.value}`,{selected:draft.holiday===h.value}]" :aria-pressed="draft.holiday===h.value" @click="applyHoliday(h.value)"><FestivalArt :holiday="h.value"/><strong>{{h.label}}</strong><small>{{h.description.split(' · ')[0]}}</small></button></div><p class="festival-date-note">{{holidayDateHint}}</p></div>
          <label class="field">给这个时刻起个名字 <span>*</span><input v-model="draft.title" required maxlength="60" placeholder="比如：下一站，去看海" aria-label="时刻标题"/></label>
          <label v-if="draft.kind!=='timer'" class="field">{{draft.kind==='anniversary'?'从哪一天开始':'期待在哪一刻抵达'}} <span>*</span><input v-model="draft.date" type="datetime-local" required aria-label="目标日期" @input="editingDateChanged=true"/><small>以本机时区填写 · {{browserTimezone}}</small></label>
          <div v-else class="field">想专注多久？<div class="duration-presets"><button v-for="minute in [5,15,25,60]" :key="minute" type="button" :class="{selected:Number(draft.duration_minutes)===minute}" @click="draft.duration_minutes=minute">{{minute}} 分</button></div><label class="duration-input"><input v-model="draft.duration_minutes" type="number" min="0.016667" max="10080" step="any" required aria-label="计时分钟"/><span>分钟</span></label></div>
          <div class="two-fields"><div class="field"><span>放在哪个分类</span><ChoiceSelect v-model="draft.category" :options="categories" label="分类" icon="grid" allow-custom :disabled="draft.kind==='festival'"/></div><div class="field"><span>重复方式</span><ChoiceSelect v-model="draft.recurrence" :options="repeatChoices" label="重复方式" icon="reset" :disabled="draft.kind==='timer'||(draft.kind==='festival'&&holidayNeedsRefresh(draft.holiday))"/></div></div>
          <label class="field">写一句话给未来<textarea v-model="draft.description" rows="2" maxlength="500" placeholder="让这份期待，再具体一点点…" aria-label="备注"/></label>
          <label class="checkbox-label"><input v-model="draft.pinned" type="checkbox"/>置顶这份特别的期待<Icon name="pin" :size="14"/></label>
        </div><div class="editor-appearance"><span class="field-label">预览你的时光卡片</span><div class="editor-preview moment-card" :class="[`theme-${draft.theme}`,draft.kind==='festival'?`festival-${draft.holiday}`:'']"><FestivalArt :holiday="draft.kind==='festival'?draft.holiday:''"/><div class="card-heading"><span class="category-badge"><Icon name="sparkles" :size="14"/>{{draft.category||'生活'}}</span><span class="preview-tag">实时预览</span></div><div class="card-title"><h3>{{draft.title||'下一份美好，正在靠近'}}</h3><p>{{timeInfo(preview,now).status}}</p></div><Counter :event="preview" :now="now"/></div><div class="field-label appearance-label">选一个心情配色</div><p v-if="draft.kind==='festival'" class="festival-date-note">节日使用专属配色与装饰</p><div v-else class="theme-picker"><button v-for="theme in themes" :key="theme.id" type="button" :style="{background:theme.color}" :class="{selected:draft.theme===theme.id}" :aria-label="theme.label" :title="theme.label" @click="draft.theme=theme.id"><Icon v-if="draft.theme===theme.id" name="check" :size="17"/></button></div><div class="field-label appearance-label">让时间换个样子</div><div class="style-picker"><button v-for="item in styles" :key="item.id" type="button" :class="{selected:draft.style===item.id}" @click="draft.style=item.id"><Icon :name="item.icon" :size="19"/>{{item.label}}</button></div><p class="appearance-note"><Icon name="leaf" :size="14"/>不必急，日子会慢慢开出花来。</p></div></div>
        <div v-if="formError" class="form-error" role="alert">{{formError}}</div><div class="dialog-footer"><span><Icon name="heart" :size="14"/>给未来一个温柔的约定</span><div><button type="button" class="secondary-button" :disabled="saving" @click="editor.close()">再想一下</button><button type="submit" class="primary-button" :disabled="saving"><Icon :name="saving?'clock':'check'" :size="17"/>{{saving?'正在保存…':editID?'保存修改':'放进时光舱'}}</button></div></div>
      </form>
    </dialog>

    <dialog ref="deleteDialog" class="confirm-dialog" aria-labelledby="delete-title" @cancel.prevent="!saving&&deleteDialog.close()"><span class="confirm-icon"><Icon name="trash" :size="25"/></span><h2 id="delete-title">{{pendingDelete?'要移走这个时刻吗？':'清空所有示例时刻？'}}</h2><p>{{pendingDelete?`「${pendingDelete.title}」将被删除，此操作无法撤销。`:'只删除标为「示例」的卡片，你自己创建或编辑过的时刻会保留。'}}</p><div><button class="secondary-button" :disabled="saving" @click="deleteDialog.close()">保留期待</button><button class="danger-button" :disabled="saving" @click="confirmDelete">{{saving?'正在删除…':'确认删除'}}</button></div></dialog>

    <dialog ref="focusDialog" class="focus-dialog" :class="[`theme-${focused?.theme||'violet'}`,focused?.holiday?`festival-${focused.holiday}`:'']" aria-label="专注时光视图" @cancel.prevent="closeFocus"><div v-if="focused" class="focus-content"><FestivalArt :holiday="focused.holiday"/><div class="focus-top"><span class="eyebrow">THIS MOMENT IS YOURS</span><div><button class="icon-button" aria-label="切换全屏" @click="fullscreen"><Icon name="expand"/></button><button class="icon-button" aria-label="关闭专注视图" @click="closeFocus"><Icon name="close"/></button></div></div><Icon name="sparkles" :size="36"/><h2>{{focused.title}}</h2><p>{{timeInfo(focused,now).status}}</p><Counter :event="focused" :now="now"/><p class="focus-description">{{focused.description}}</p><span class="focus-date">{{dateLabel(focused)}}</span><div v-if="focused.kind==='timer'" class="focus-actions"><button class="primary-button" @click="action(focused,timeInfo(focused,now).paused?'resume':'pause')"><Icon :name="timeInfo(focused,now).paused?'play':'pause'" :size="18"/>{{timeInfo(focused,now).paused?'继续计时':'暂停一下'}}</button><button class="secondary-button" @click="action(focused,'restart')"><Icon name="reset" :size="17"/>重新开始</button></div></div></dialog>
    <ImmersiveStage :effect="effectMode" ref="stage" v-model:scene="scene" :events="events" :now="now" @error="toast($event,'error')"/>
    <div v-if="celebrating&&motionOn" class="celebration" aria-hidden="true"><i v-for="index in 36" :key="index" :style="{'--x':`${(index*37)%100}%`,'--drift':`${((index*17)%160)-80}px`,'--delay':`${(index%7)*.07}s`,'--confetti-color':['#a589df','#eab1bd','#8fc6c3','#ead0a1'][index%4]}"/></div>
    <div class="toast-stack" aria-live="polite"><TransitionGroup name="toast"><div v-for="notification in notifications" :key="notification.id" class="toast-message" :class="notification.type"><Icon :name="notification.type==='error'?'close':'check'" :size="18"/>{{notification.text}}<button class="icon-button" aria-label="关闭提示" @click="notifications=notifications.filter(n=>n.id!==notification.id)"><Icon name="close" :size="14"/></button></div></TransitionGroup></div>
  </div>
</template>
