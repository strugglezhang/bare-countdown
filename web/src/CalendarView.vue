<script setup>
import { computed, ref } from 'vue'
import Icon from './Icon.vue'
import { dateKey, monthCells } from './calendar.js'
import { timeInfo } from './time.js'
const props=defineProps({events:Array,now:Number})
const emit=defineEmits(['focus'])
const initial=new Date(props.now),year=ref(initial.getFullYear()),month=ref(initial.getMonth()),selected=ref(dateKey(initial))
const cells=computed(()=>monthCells(year.value,month.value))
const today=computed(()=>dateKey(new Date(props.now)))
const byDate=computed(()=>{
  const result={}
  for(const e of props.events){const date=new Date(e.effective_target_at||e.target_at);if(!Number.isFinite(date.getTime()))continue;const key=dateKey(date);(result[key]||=[]).push(e)}
  return result
})
const selectedEvents=computed(()=>byDate.value[selected.value]||[])
function move(amount){const date=new Date(year.value,month.value+amount,1,12);year.value=date.getFullYear();month.value=date.getMonth()}
function goToday(){const date=new Date(props.now);year.value=date.getFullYear();month.value=date.getMonth();selected.value=dateKey(date)}
</script>
<template>
  <div class="calendar-view">
    <div class="calendar-board"><div class="calendar-toolbar"><div><h3>{{year}} 年 {{month+1}} 月</h3><span>把期待，安排进每一个日子</span></div><div><button class="icon-button" aria-label="上个月" @click="move(-1)"><Icon name="chevron" class="calendar-prev" :size="17"/></button><button class="calendar-today-button" @click="goToday">回到今天</button><button class="icon-button" aria-label="下个月" @click="move(1)"><Icon name="chevron" class="calendar-next" :size="17"/></button></div></div>
      <div class="calendar-weekdays"><span v-for="day in ['一','二','三','四','五','六','日']" :key="day">{{day}}</span></div>
      <div class="calendar-days"><button v-for="cell in cells" :key="cell.key" class="calendar-day" :class="{'other-month':!cell.current,'is-today':cell.key===today,selected:cell.key===selected}" :data-date="cell.key" :aria-label="`${cell.key}，${(byDate[cell.key]||[]).length} 个时刻`" :aria-pressed="cell.key===selected" @click="selected=cell.key"><strong>{{cell.day}}</strong><span class="calendar-event-dots"><i v-for="event in (byDate[cell.key]||[]).slice(0,3)" :key="event.id" :class="[`theme-${event.theme}`,event.holiday?`festival-${event.holiday}`:'']"/><small v-if="(byDate[cell.key]||[]).length>3">+{{byDate[cell.key].length-3}}</small></span><span v-if="byDate[cell.key]?.length" class="calendar-day-title">{{byDate[cell.key][0].title}}</span></button></div>
      <p class="calendar-timezone">按本机时区显示日期 · {{Intl.DateTimeFormat().resolvedOptions().timeZone}}</p>
    </div>
    <aside class="calendar-agenda"><span class="eyebrow">A DAY WORTH REMEMBERING</span><h3>{{selected.slice(5).replace('-',' 月 ')}} 日</h3><p>当天的期待 · {{selectedEvents.length}} 个</p><button v-for="event in selectedEvents" :key="event.id" class="calendar-event" :class="[`theme-${event.theme}`,event.holiday?`festival-${event.holiday}`:'']" @click="emit('focus',event)"><span>{{event.category}}</span><strong>{{event.title}}</strong><small>{{timeInfo(event,now).status}} · 点击专注查看</small><Icon name="arrow" :size="16"/></button><div v-if="!selectedEvents.length" class="calendar-day-empty"><Icon name="leaf" :size="32"/><span>留白的日子<br/>也有它自己的美好。</span></div></aside>
  </div>
</template>
