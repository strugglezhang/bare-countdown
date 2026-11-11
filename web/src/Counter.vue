<script setup>
import { computed, useId } from 'vue'
import { timeInfo } from './time.js'
import { sevenSegments,segmentPaths,dotPatterns } from './clock-fonts.js'
const props = defineProps({ event: Object, now: Number, forceStyle: String, clock:Boolean })
const pad=value=>String(value).padStart(2,'0')
const info = computed(() => {
  const base=timeInfo(props.event,props.now)
  if(!props.clock)return base
  const date=new Date(props.now),hours=date.getHours(),minutes=date.getMinutes(),secs=date.getSeconds(),seconds=hours*3600+minutes*60+secs
  return {...base,seconds,days:0,hours,minutes,secs,progress:seconds/86400,paused:false,reached:false,elapsed:false,status:'此刻时钟',parts:[{value:pad(hours),label:'时'},{value:pad(minutes),label:'分'},{value:pad(secs),label:'秒'}]}
})
const mode = computed(() => props.forceStyle || props.event.style)
const timerText = computed(() => props.clock?`${pad(info.value.hours)}:${pad(info.value.minutes)}:${pad(info.value.secs)}`:`${String(Math.floor(info.value.seconds / 60)).padStart(2, '0')}:${String(info.value.secs).padStart(2, '0')}`)
const mainValue=computed(()=>props.clock||props.event.kind==='timer'?timerText.value:String(info.value.days).padStart(2,'0'))
const mainLabel=computed(()=>props.clock?'此时此刻':props.event.kind==='timer'?'保持专注':info.value.elapsed?'已经相伴 · 天':'距离抵达 · 天')
const tail=computed(()=>`${pad(info.value.hours)}:${pad(info.value.minutes)}:${pad(info.value.secs)}`)
const clipId='clock-'+useId().replace(/[^a-z0-9]/gi,'')
const handAngles=computed(()=>[(info.value.hours%12+info.value.minutes/60)*30,(info.value.minutes+info.value.secs/60)*6,info.value.secs*6])
const binaryParts=computed(()=>info.value.parts.map(part=>({...part,bits:Math.max(0,Number(part.value)).toString(2).padStart(6,'0').split('')})))
const starPoints=[{x:16,y:28,r:2},{x:44,y:18,r:3},{x:78,y:36,r:2},{x:106,y:13,r:2},{x:146,y:30,r:3},{x:181,y:16,r:2},{x:217,y:42,r:2},{x:231,y:76,r:3},{x:205,y:114,r:2},{x:168,y:105,r:2},{x:135,y:123,r:3},{x:94,y:112,r:2},{x:56,y:126,r:2},{x:22,y:96,r:3},{x:9,y:66,r:2}]
const chars=computed(()=>[...mainValue.value])
const needle=computed(()=>({x:80+Math.cos(Math.PI*(1+info.value.progress))*43,y:80+Math.sin(Math.PI*(1+info.value.progress))*43}))
const ticks=Array.from({length:11},(_,i)=>{const a=Math.PI*(1+i/10);return{x1:80+Math.cos(a)*55,y1:80+Math.sin(a)*55,x2:80+Math.cos(a)*48,y2:80+Math.sin(a)*48}})
const stepRows=computed(()=>info.value.parts.map(part=>{
  const value=Number(part.value),range=part.label==='天'?Math.max(1,Math.ceil(Math.abs(Date.parse(props.event.effective_target_at||props.event.target_at)-Date.parse(props.event.created_at))/86400000)||365):part.label==='时'?24:60
  return {...part,lit:Math.min(16,Math.ceil(value/range*16))}
}))
</script>
<template>
  <div class="counter" :class="[`counter-${mode}`, { 'counter-arrived': info.reached, 'counter-is-clock':clock,'counter-is-paused':info.paused }]" :style="{'--clock-value-length':mainValue.length}" :aria-label="`${info.status} ${info.days}天${info.hours}时${info.minutes}分${info.secs}秒`">
    <template v-if="mode === 'flip'">
      <div v-for="(part, i) in info.parts" :key="part.label" class="flip-unit">
        <span v-if="i" class="flip-colon">:</span>
        <div class="flip-number" :class="{'long-number':part.value.length>2}" :style="{'--number-length':part.value.length}"><Transition name="tick" mode="out-in"><span :key="part.value">{{ part.value }}</span></Transition></div>
        <span class="unit-label">{{ part.label }}</span>
      </div>
    </template>
    <template v-else-if="mode === 'ring'">
      <div class="ring-wrap">
        <svg viewBox="0 0 144 144" class="ring-svg" aria-hidden="true"><circle cx="72" cy="72" r="64" class="ring-track"/><circle cx="72" cy="72" r="64" class="ring-progress" :style="{ strokeDasharray: 402.13, strokeDashoffset: 402.13 * (1 - info.progress) }"/></svg>
        <div class="ring-center"><strong v-if="event.kind === 'timer'" class="timer-value">{{ timerText }}</strong><strong v-else>{{ info.days }}<small>天</small></strong><span>{{ info.paused ? '给时间按个暂停' : info.reached ? '就是现在' : event.kind === 'timer' ? '专注于此刻' : `${String(info.hours).padStart(2, '0')} 时 ${String(info.minutes).padStart(2, '0')} 分` }}</span></div>
      </div>
    </template>
    <template v-else-if="mode === 'minimal'">
      <div class="minimal-number"><strong>{{ event.kind === 'timer' ? timerText : info.days }}</strong><span>{{ event.kind === 'timer' ? '' : 'DAYS' }}</span></div>
      <div class="time-progress"><span :style="{ width: `${info.progress * 100}%` }"/></div>
      <span class="mini-time">{{ info.hours }} 小时 · {{ info.minutes }} 分钟 · {{ info.secs }} 秒</span>
    </template>
    <template v-else-if="mode === 'neon'">
      <div class="neon-display"><span class="display-caption">{{mainLabel}}</span><svg :viewBox="`0 0 ${chars.length*34} 50`" class="segment-display" aria-hidden="true"><g v-for="(char,index) in chars" :key="index" :transform="`translate(${index*34} 0)`"><template v-if="char===':'"><circle cx="14" cy="17" r="2.2"/><circle cx="14" cy="33" r="2.2"/></template><path v-for="(path,id) in char===':'?{}:segmentPaths" :key="id" :d="path" :class="sevenSegments[char]?.includes(id)?'segment-on':'segment-off'"/></g></svg><div class="display-bottom"><i/><span>{{clock?'LOCAL TIME':event.kind==='timer'?'FOCUS MODE':tail}}</span><i/></div></div>
    </template>
    <template v-else-if="mode === 'matrix'">
      <div class="matrix-display"><span class="display-caption">{{mainLabel}}</span><svg :viewBox="`0 0 ${chars.length*24-4} 28`" class="dot-display" aria-hidden="true"><g v-for="(char,index) in chars" :key="index" :transform="`translate(${index*24} 0)`"><g v-for="(row,y) in dotPatterns[char]||[]" :key="y"><circle v-for="(dot,x) in row" :key="x" :cx="x*4+2" :cy="y*4+2" r="1.3" :class="dot==='1'?'dot-on':'dot-off'"/></g></g></svg><span class="matrix-tail">{{clock?'把当下，过得闪闪发光':event.kind==='timer'?'每一秒，都算数':tail}}</span></div>
    </template>
    <template v-else-if="mode === 'gauge'">
      <div class="gauge-display"><svg viewBox="0 0 160 95" aria-hidden="true"><path d="M20 80a60 60 0 0 1 120 0" class="gauge-track"/><path d="M20 80a60 60 0 0 1 120 0" class="gauge-arc" pathLength="100" :stroke-dasharray="`${info.progress*100} 100`"/><line v-for="(tick,index) in ticks" :key="index" v-bind="tick" class="gauge-tick"/><line x1="80" y1="80" :x2="needle.x" :y2="needle.y" class="gauge-needle"/><circle cx="80" cy="80" r="4" fill="currentColor"/></svg><strong :style="{'--value-length':mainValue.length}">{{mainValue}}</strong><span>{{mainLabel}}</span></div>
    </template>
    <template v-else-if="mode === 'steps'">
      <div class="step-display"><div v-for="part in stepRows" :key="part.label" class="step-row"><span><strong>{{part.value}}</strong><small>{{part.label}}</small></span><div class="step-bars"><i v-for="index in 16" :key="index" :class="{lit:index<=part.lit}"/></div></div></div>
    </template>
    <template v-else-if="mode === 'ticket'">
      <div class="ticket-display"><div class="ticket-main"><span>TIME CAPSULE / 时光珍藏</span><strong :style="{'--value-length':mainValue.length}">{{mainValue}}</strong><small>{{mainLabel}}</small></div><div class="ticket-stub"><span>{{clock?'此时此刻':info.elapsed?'一路相伴':'即将抵达'}}</span><strong>{{clock?'NOW':tail}}</strong><svg viewBox="0 0 96 20" aria-hidden="true"><rect v-for="index in 32" :key="index" :x="(index-1)*3" y="0" :width="index%3===0?2:1" :height="index%5===0?14:20"/></svg><small>每一个时刻，都值得</small></div></div>
    </template>
    <template v-else-if="mode==='analog'">
      <div class="analog-display"><svg viewBox="0 0 120 120" aria-hidden="true"><circle cx="60" cy="60" r="55" class="clock-face"/><line v-for="i in 12" :key="i" x1="60" y1="9" x2="60" :y2="i%3===0?17:13" :transform="`rotate(${i*30} 60 60)`" class="clock-tick"/><line v-for="(angle,i) in handAngles" :key="i" x1="60" y1="60" x2="60" :y2="[32,20,15][i]" :transform="`rotate(${angle} 60 60)`" :class="['clock-hand','clock-hand-'+i]"/><circle cx="60" cy="60" r="3" fill="currentColor"/></svg><span>{{clock?tail:mainValue+' · '+mainLabel}}</span></div>
    </template>
    <template v-else-if="mode==='hourglass'">
      <div class="hourglass-display"><svg viewBox="0 0 100 120" aria-hidden="true"><defs><clipPath :id="clipId"><path d="M20 12h60v10L56 58v4l24 36v10H20V98l24-36v-4L20 22Z"/></clipPath></defs><g :clip-path="`url(#${clipId})`" class="sand-fill"><rect x="20" :y="12+46*(1-info.progress)" width="60" :height="46*info.progress"/><path :d="`M20 108L50 ${108-42*(1-info.progress)}L80 108Z`"/><path v-if="!info.paused&&!info.reached" d="M50 59v43" class="sand-stream"/></g><path d="M20 12h60v10L56 58v4l24 36v10H20V98l24-36v-4L20 22Z" class="glass-outline"/><path d="M15 8h70M15 112h70" class="glass-cap"/></svg><div><strong>{{mainValue}}</strong><span>{{mainLabel}}</span><small>{{tail}}</small></div></div>
    </template>
    <template v-else-if="mode==='orbit'">
      <div class="orbit-display"><svg viewBox="0 0 140 140" aria-hidden="true"><g v-for="(value,i) in [info.hours/24,info.minutes/60,info.secs/60]" :key="i" :style="{opacity:1-i*.2}"><circle cx="70" cy="70" :r="62-i*10" class="orbit-track"/><circle cx="70" cy="70" :r="62-i*10" class="orbit-progress" pathLength="100" :stroke-dasharray="`${value*100} 100`" transform="rotate(-90 70 70)"/></g></svg><div><strong>{{mainValue}}</strong><span>{{mainLabel}}</span></div></div>
    </template>
    <template v-else-if="mode==='terminal'"><div class="terminal-display"><header><i/><i/><i/><span>time@capsule: ~</span></header><p>❯ {{clock?'clock --local':'countdown --live'}}</p><strong>{{mainValue}}<b>▍</b></strong><span>{{mainLabel}} / {{tail}}</span><footer>● {{info.paused?'PAUSED':'CONNECTED TO THE MOMENT'}}</footer></div></template>
    <template v-else-if="mode==='note'"><div class="note-display"><i/><span>Dear future,</span><strong>{{mainValue}}</strong><small>{{mainLabel}}</small><footer>{{clock?'把今天，写成喜欢的样子':tail+' · 期待如约而至'}} ♡</footer></div></template>
    <template v-else-if="mode==='liquid'"><div class="liquid-display"><svg viewBox="0 0 200 120" preserveAspectRatio="none" aria-hidden="true"><defs><clipPath :id="clipId"><rect x="1" y="1" width="198" height="118" rx="22"/></clipPath></defs><g :clip-path="`url(#${clipId})`"><rect width="200" height="120" class="liquid-base"/><g :transform="`translate(0 ${112-info.progress*100})`"><path d="M-200 0Q-150 -12 -100 0T0 0T100 0T200 0T300 0T400 0V150H-200Z" class="liquid-wave"/></g></g></svg><div><strong>{{mainValue}}</strong><span>{{mainLabel}}</span><small>{{tail}}</small></div></div></template>

    <template v-else-if="mode==='calendar'"><div class="tear-calendar"><header><i/><span>{{clock?'此刻 · LOCAL TIME':info.elapsed?'相伴 · EVERY DAY':'期待 · SAVE THE DATE'}}</span><i/></header><div><strong>{{mainValue}}</strong><span>{{mainLabel}}</span></div><footer>{{clock?'把今天，过成值得珍藏的一页':tail}}</footer></div></template>
    <template v-else-if="mode==='odometer'"><div class="odometer-display"><span>{{mainLabel}}</span><div class="odometer-digits"><template v-for="(char,i) in chars" :key="i"><b v-if="char===':'" class="odometer-colon">:</b><div v-else class="odometer-wheel"><small>{{(Number(char)+9)%10}}</small><Transition name="odometer-tick" mode="out-in"><strong :key="char">{{char}}</strong></Transition><small>{{(Number(char)+1)%10}}</small></div></template></div><footer>{{clock?'THE MOMENT IS NOW':tail+' · EVERY SECOND COUNTS'}}</footer></div></template>
    <template v-else-if="mode==='cassette'"><div class="cassette-display"><header><span>SIDE A / 时光原声</span><b>● REC</b></header><strong>{{mainValue}}</strong><div class="cassette-tape"><svg viewBox="0 0 200 42" aria-hidden="true"><path d="M40 21h120" stroke="currentColor" stroke-opacity=".25" stroke-width="9"/><g v-for="x in [35,165]" :key="x" :transform="`translate(${x} 21)`"><g class="cassette-reel clock-decoration-motion"><circle r="17"/><path d="M0-12V12M-10-6 10 6M-10 6 10-6"/><circle r="6" class="reel-center"/></g></g></svg><span>{{mainLabel}}</span></div><footer><i/>{{tail}}<i/></footer></div></template>
    <template v-else-if="mode==='polaroid'"><div class="polaroid-display"><div class="polaroid-photo"><svg viewBox="0 0 240 110" preserveAspectRatio="none" aria-hidden="true"><circle cx="177" cy="28" r="17"/><path d="m0 80 55-47 58 45 40-33 87 58v7H0Z"/><path d="m0 96 71-32 68 35 57-34 44 26v19H0Z"/></svg><strong>{{mainValue}}</strong><span>{{mainLabel}}</span></div><footer>{{clock?'把这一刻，留在心里':tail+' · 等一个好日子'}}<i>♡</i></footer></div></template>
    <template v-else-if="mode==='binary'"><div class="binary-display"><header>BINARY / {{clock?'此刻时钟':'时间编码'}}</header><div class="binary-columns"><div v-for="part in binaryParts" :key="part.label"><div class="binary-lights" aria-hidden="true"><i v-for="(bit,i) in part.bits" :key="i" :class="{lit:bit==='1'}"/></div><strong :style="{'--part-length':part.value.length}">{{part.value}}</strong><small>{{part.label}}</small></div></div><footer>把每一秒，写成独一无二的代码</footer></div></template>
    <template v-else-if="mode==='radar'"><div class="radar-display"><svg viewBox="0 0 140 140" aria-hidden="true"><circle v-for="r in [22,42,62]" :key="r" cx="70" cy="70" :r="r"/><path d="M70 8v124M8 70h124M26 26l88 88M26 114l88-88" class="radar-grid"/><g class="radar-sweep clock-decoration-motion"><path d="M70 70V8A62 62 0 0 1 132 70Z" class="radar-sector"/><path d="M70 70V8" class="radar-beam"/></g><circle cx="104" cy="39" r="3" class="radar-blip clock-decoration-motion"/></svg><div><strong>{{mainValue}}</strong><span>{{mainLabel}}</span><small>{{tail}}</small></div></div></template>
    <template v-else-if="mode==='scroll'"><div class="bamboo-display"><i class="scroll-cap cap-left"/><i class="scroll-cap cap-right"/><span class="scroll-seal">時</span><header>光阴有信 · 岁月成诗</header><strong>{{mainValue}}</strong><span>{{mainLabel}}</span><footer>{{clock?'且将新火试新茶':tail+' · 静候佳期'}}</footer></div></template>
    <template v-else-if="mode==='prism'"><div class="prism-display"><svg viewBox="0 0 240 140" preserveAspectRatio="none" aria-hidden="true"><path d="m120 7 110 121H10Z"/><path d="m120 7 15 95 95 26M10 128l125-26"/><path d="M0 55 135 102l105-67M135 102l105-38M135 102l105-9"/></svg><div><span>{{mainLabel}}</span><strong>{{mainValue}}</strong><small>{{clock?'A SPECTRUM OF NOW':tail}}</small></div><i class="prism-glint clock-decoration-motion"/></div></template>
    <template v-else-if="mode==='constellation'"><div class="constellation-display"><svg viewBox="0 0 240 140" aria-hidden="true"><path d="M16 28 44 18 78 36 106 13 146 30 181 16 217 42 231 76 205 114 168 105 135 123 94 112 56 126 22 96 9 66Z"/><circle v-for="(star,i) in starPoints" :key="i" v-bind="{cx:star.x,cy:star.y,r:star.r}" class="constellation-star clock-decoration-motion" :style="{'--star-delay':`${-i*.3}s`}"/><path d="m204 17 3 7 7 3-7 3-3 7-3-7-7-3 7-3Z" class="constellation-spark"/></svg><div><span>{{mainLabel}}</span><strong>{{mainValue}}</strong><small>{{clock?'此刻，你也在闪闪发光':tail}}</small></div></div></template>
    <template v-else>
      <div class="digital-number"><strong>{{ event.kind === 'timer' ? timerText : info.days.toLocaleString() }}</strong><span>{{ event.kind === 'timer' ? '' : '天' }}</span></div>
      <div class="digital-tail"><span>{{ String(info.hours).padStart(2, '0') }} <small>时</small></span><i>·</i><span>{{ String(info.minutes).padStart(2, '0') }} <small>分</small></span><i>·</i><span>{{ String(info.secs).padStart(2, '0') }} <small>秒</small></span></div>
    </template>
  </div>
</template>
