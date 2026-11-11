export const kindLabels = { countdown: '倒计时', anniversary: '纪念日', timer: '短时计时', festival:'节日倒计时' }
export const themes = [
  { id: 'violet', label: '暮光紫', color: '#8764ed' },
  { id: 'rose', label: '落日粉', color: '#e5749b' },
  { id: 'ocean', label: '海盐蓝', color: '#5a96d8' },
  { id: 'amber', label: '暖杏橙', color: '#dca458' },
  { id: 'mint', label: '薄荷绿', color: '#58a592' },
]
export const styles = [
  { id: 'flip', label: '翻页时钟', icon: 'clock' },
  { id: 'digits', label: '发光数字', icon: 'sparkles' },
  { id: 'ring', label: '时光光环', icon: 'circle' },
  { id: 'minimal', label: '极简进度', icon: 'line' },
  { id: 'neon', label: '霓虹数码管', icon: 'sparkles' },
  { id: 'gauge', label: '机械仪表', icon: 'target' },
  { id: 'matrix', label: '点阵时钟', icon: 'grid' },
  { id: 'steps', label: '进度阶梯', icon: 'list' },
  { id: 'ticket', label: '时间票根', icon: 'calendar' },
  { id:'analog',label:'经典指针',icon:'clock' },
  { id:'hourglass',label:'流沙沙漏',icon:'clock' },
  { id:'orbit',label:'星际轨道',icon:'circle' },
  { id:'terminal',label:'极客终端',icon:'grid' },
  { id:'note',label:'手账便签',icon:'calendar' },
  { id:'liquid',label:'水波时光',icon:'circle' },
  { id:"calendar",label:"撕页日历",icon:"calendar" },
  { id:"odometer",label:"机械滚轮",icon:"clock" },
  { id:"cassette",label:"磁带记忆",icon:"grid" },
  { id:"polaroid",label:"拍立得时光",icon:"heart" },
  { id:"binary",label:"二进制灯阵",icon:"grid" },
  { id:"radar",label:"时光雷达",icon:"target" },
  { id:"scroll",label:"竹简时光",icon:"list" },
  { id:"prism",label:"棱镜光谱",icon:"sparkles" },
  { id:"constellation",label:"星图时钟",icon:"sparkles" },

]
export const recurrenceLabels = { none: '不重复', daily: '每天', weekly: '每周', yearly: '每年' }

export function timeInfo(event, now = Date.now()) {
  const target = Date.parse(event.effective_target_at || event.target_at)
  const elapsed = event.kind === 'anniversary' && event.recurrence === 'none' && target <= now
  const paused = event.paused_remaining != null
  let seconds = paused ? event.paused_remaining : elapsed ? Math.floor((now - target) / 1000) : Math.max(0, Math.ceil((target - now) / 1000))
  seconds = Number.isFinite(seconds) ? Math.max(0, seconds) : 0
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor(seconds % 86400 / 3600)
  const minutes = Math.floor(seconds % 3600 / 60)
  const secs = seconds % 60
  const origin = event.kind === 'timer' ? event.duration_seconds * 1000 : target - Date.parse(event.created_at)
  const progress = elapsed ? 1 : Math.max(0, Math.min(1, seconds * 1000 / Math.max(1, origin)))
  return { seconds, days, hours, minutes, secs, elapsed, paused, reached: !paused && !elapsed && seconds === 0, progress,
    parts: [{ value: String(days).padStart(2, '0'), label: '天' }, { value: String(hours).padStart(2, '0'), label: '时' }, { value: String(minutes).padStart(2, '0'), label: '分' }, { value: String(secs).padStart(2, '0'), label: '秒' }],
    status: paused ? '已暂停' : elapsed ? '已经相伴' : seconds === 0 ? '期待已抵达' : event.kind === 'timer' ? '专注进行中' : '距离这一天还有',
  }
}

export function dateLabel(event) {
  return new Intl.DateTimeFormat('zh-CN', { timeZone: event.timezone, year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', hour12: false }).format(new Date(event.effective_target_at || event.target_at))
}
export function localInput(iso = new Date(Date.now() + 86400000).toISOString()) {
  const d = new Date(iso)
  if (!Number.isFinite(d.getTime())) return ''
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function newDraft(kind = 'countdown') {
  return { title: '', description: '', category: kind === 'festival' ? '节日' : kind === 'timer' ? '专注' : kind === 'anniversary' ? '纪念日' : '生活', kind, holiday:'', theme: 'violet', style: kind === 'timer' ? 'ring' : 'flip', recurrence: 'none', timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Shanghai', pinned: false, duration_minutes: 25, date: localInput() }
}
