<script setup>
import { computed, onUnmounted, ref } from 'vue'
import SceneArt from './SceneArt.vue'
import AmbientSound from './AmbientSound.vue'
import Counter from './Counter.vue'
import Icon from './Icon.vue'
import BrandMark from './BrandMark.vue'
import FestivalArt from './FestivalArt.vue'
import ChoiceSelect from './ChoiceSelect.vue'
import {scenes} from './scenes.js'
import {timeInfo,styles} from './time.js'
const props=defineProps({scene:String,events:Array,now:Number,effect:String})
const emit=defineEmits(['update:scene','error'])
const dialog=ref(null),opened=ref(false),mode=ref('clock'),selected=ref(''),style=ref('flip'),sound=ref(null)
const clockStyle=ref('digital')
const sceneChoices=scenes.map(s=>({value:s.id,label:s.label,icon:s.icon}))
const styleChoices=styles.map(s=>({value:s.id,label:s.label,icon:s.icon}))
const clockChoices=[{value:'digital',label:'经典大数字'},...styleChoices]
const eventChoices=computed(()=>props.events.map(e=>({value:e.id,label:e.title,description:e.category})))
const clockEvent=computed(()=>({kind:'timer',duration_seconds:86400,recurrence:'none',target_at:new Date(props.now+86400000).toISOString(),created_at:new Date(props.now).toISOString()}))
const currentScene=computed(()=>scenes.find(s=>s.id===props.scene)||scenes[0])
const current=computed(()=>props.events.find(e=>e.id===selected.value))
const clock=computed(()=>new Intl.DateTimeFormat('en-GB',{hour:'2-digit',minute:'2-digit',second:'2-digit',hourCycle:'h23'}).format(new Date(props.now)).split(':'))
const date=computed(()=>new Intl.DateTimeFormat('zh-CN',{year:'numeric',month:'long',day:'numeric',weekday:'long'}).format(new Date(props.now)))
function open(){selected.value=props.events[0]?.id||'';opened.value=true;dialog.value.showModal()}
function close(){sound.value?.stop();if(document.fullscreenElement===dialog.value)document.exitFullscreen().catch(()=>{});dialog.value.close();opened.value=false}
async function fullscreen(){try{if(document.fullscreenElement)await document.exitFullscreen();else await dialog.value.requestFullscreen()}catch{emit('error','此浏览器暂不支持系统全屏，可以继续使用大屏视图')}}
onUnmounted(()=>sound.value?.stop())
defineExpose({open})
</script>
<template>
  <dialog ref="dialog" class="immersive-stage" :class="mode==='countdown'&&current?.holiday?`festival-${current.holiday}`:''" aria-label="沉浸时光大屏" @cancel.prevent="close">
    <template v-if="opened"><SceneArt :scene="scene" :now="now" :effect="effect"/><div class="stage-shade"/><FestivalArt v-if="mode==='countdown'" :holiday="current?.holiday"/>
      <header class="stage-header"><span class="stage-brand"><BrandMark/><span class="stage-wordmark"><strong>时光舱</strong><small>TIME CAPSULE</small></span></span><div><ChoiceSelect :model-value="scene" :options="sceneChoices" label="大屏氛围" icon="sparkles" compact inverse @update:model-value="emit('update:scene',$event)"/><button class="icon-button" aria-label="大屏切换全屏" @click="fullscreen"><Icon name="expand"/></button><button class="icon-button" aria-label="关闭大屏" @click="close"><Icon name="close"/></button></div></header>
      <div class="stage-controls"><button :class="{selected:mode==='clock'}" :aria-pressed="mode==='clock'" @click="mode='clock'">此刻时钟</button><button :class="{selected:mode==='countdown'}" :aria-pressed="mode==='countdown'" :disabled="!events.length" @click="mode='countdown'">我的倒计时</button></div>
      <div class="stage-main"><span class="stage-eyebrow">{{currentScene.english}}</span>
        <template v-if="mode==='clock'||!current"><div v-if="clockStyle==='digital'" class="stage-digital"><strong>{{clock[0]}}<i>:</i>{{clock[1]}}</strong><span>{{clock[2]}}</span></div><Counter v-else :event="clockEvent" :now="now" :force-style="clockStyle" clock/><p class="stage-date">{{date}}</p><ChoiceSelect v-model="clockStyle" :options="clockChoices" label="当前时钟样式" icon="clock" compact inverse class="stage-clock-select"/></template>
        <template v-else><h2>{{current.title}}</h2><p class="stage-status">{{timeInfo(current,now).status}}</p><Counter :event="current" :now="now" :force-style="style"/><div class="stage-event-select"><ChoiceSelect v-model="selected" :options="eventChoices" label="大屏选择倒计时" compact inverse/><ChoiceSelect v-model="style" :options="styleChoices" label="大屏时钟样式" icon="clock" compact inverse/></div></template>
        <p class="stage-quote">{{currentScene.caption}}</p>
      </div>
      <footer class="stage-footer"><span><i/>给此刻，留一点自己的时间。</span><AmbientSound ref="sound" @error="emit('error',$event)"/></footer>
    </template>
  </dialog>
</template>
