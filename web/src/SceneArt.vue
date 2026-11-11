<script setup>
import { computed, onMounted, onUnmounted, ref, useId } from 'vue'
import { daylightAt, celestialPosition } from './daylight.js'
import AmbientEffect from './AmbientEffect.vue'
const props = defineProps({ scene: String, now: { type:Number, default:()=>Date.now() }, effect:{type:String,default:'flow'} })
const drawing=ref(null),size=ref({width:1200,height:360})
let observer
onMounted(()=>{
  const measure=()=>{const box=drawing.value?.getBoundingClientRect();if(box?.width&&box?.height)size.value={width:box.width,height:box.height}}
  measure();observer=new ResizeObserver(measure);observer.observe(drawing.value)
})
onUnmounted(()=>observer?.disconnect())
const celestial=computed(()=>daylightAt(props.now,props.scene))
const position=computed(()=>celestialPosition(celestial.value,size.value.width,size.value.height))
const uid = useId().replace(/[^a-z0-9-]/gi,'')
const paints = { sky:`${uid}-sky`, glow:`${uid}-glow`, sun:`${uid}-sun`, moon:`${uid}-moon`, aurora:`${uid}-aurora` }
const particles = Array.from({length:30},(_,i)=>({x:(i*197+47)%1200,y:(i*83)%360,delay:`${i%11*.43}s`,r:1+(i%3)*.65}))
const url = name=>`url(#${paints[name]})`
const stars = [[110,34,1.2],[227,67,1.5],[346,31,1],[509,66,1.8],[621,29,1],[797,58,1.3],[913,37,1.5],[1094,78,1.2],[1152,22,1]]
</script>

<template>
  <svg ref="drawing" class="scene-art" :data-period="celestial.period" :data-body="celestial.body" :data-time-progress="celestial.progress.toFixed(6)" :style="{'--star-visibility':celestial.stars}" viewBox="0 0 1200 360" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
    <defs>
      <linearGradient :id="paints.sky" x1="0" y1="0" x2="0.7" y2="1"><stop stop-color="var(--time-sky-top,var(--sky-top))"/><stop offset="1" stop-color="var(--time-sky-bottom,var(--sky-bottom))"/></linearGradient>
      <radialGradient :id="paints.glow"><stop :stop-color="celestial.body==='sun'?'var(--sun-color)':'#b5c8ff'" stop-opacity=".2"/><stop offset="1" stop-color="var(--sun-color)" stop-opacity="0"/></radialGradient>
      <linearGradient :id="paints.sun" x1="0" y1="0" x2="0" y2="1"><stop stop-color="#fff5df"/><stop offset="1" stop-color="var(--sun-color)"/></linearGradient>
      <mask :id="paints.moon" maskUnits="userSpaceOnUse" x="-35" y="-35" width="70" height="70"><circle r="29" fill="white"/><circle cx="13" cy="-9" r="25" fill="black"/></mask>
      <linearGradient :id="paints.aurora" x1="0" y1="0" x2="0" y2="1"><stop stop-color="#67ffd4" stop-opacity="0"/><stop offset=".55" stop-color="#36e2af" stop-opacity=".6"/><stop offset="1" stop-color="#a58cff" stop-opacity="0"/></linearGradient>
    </defs>
    <path d="M0 0h1200v360H0z" :fill="url('sky')"/>
    <g class="scene-celestial" :transform="`translate(${position.x} ${position.y})`" :data-x="position.x.toFixed(2)" :data-y="position.y.toFixed(2)">
      <circle class="celestial-aura" r="130" :fill="url('glow')"/>
      <template v-if="scene==='cyber'"><circle r="62" fill="none" stroke="#ed67dc" stroke-width="2"/><circle r="51" fill="none" stroke="#ed67dc" stroke-opacity=".3"/></template>
      <circle v-else-if="celestial.body==='sun'" r="44" :fill="url('sun')"/>
      <circle v-else r="29" fill="#e9efff" :mask="url('moon')"/>
    </g>
    <g v-if="scene==='galaxy'||scene==='aurora'" class="deep-stars" fill="#ecedff"><circle v-for="(p,i) in particles" :key="i" :cx="p.x" :cy="p.y*.65" :r="p.r" :style="{'--star-delay':p.delay}"/></g>
    <g v-if="scene==='aurora'" class="aurora-curtain" :fill="url('aurora')"><path d="M300-70q160 180 330 20t380 0v210q-190-80-380 10T300 70Z"/><path d="M600-60q160 180 330 20t380 0v150q-190-60-380 40T600 70Z" opacity=".6"/></g>
    <g v-if="scene==='galaxy'" class="shooting-stars" fill="none" stroke="#e7dcff" stroke-width="1.8"><path class="meteor" d="m860 50-70 35"/><path class="meteor meteor-two" d="m440 40-65 32"/></g>
    <g v-if="scene==='cyber'" class="cyber-city"><path d="M0 280h1200M0 310h1200M0 345h1200M600 250 100 360M600 250 350 360M600 250v110m0-110 250 110M600 250l500 110" fill="none" stroke="#b44fe8" stroke-opacity=".55"/><path d="M850 270V170h30v100m10 0V139h42v131m10 0V200h24v70m15 0V120h33v150m13 0V184h41v86m15 0V153h46v117" fill="#1e1741" stroke="#7d60cf" stroke-width="1.5"/><path class="neon-line" d="M856 180v75m40-105v100m90-120v128m49-65v52m50-82v88" stroke="#64e6ec" stroke-width="2"/></g>
    <g class="scene-stars" fill="#fff0de"><circle v-for="([x,y,r],index) in stars" :key="x" :cx="x" :cy="y" :r="r" :style="{'--star-delay':`${index*.7}s`}"/></g>
    <g class="scene-clouds" fill="none" stroke="#fce7d7" stroke-opacity=".12" stroke-linecap="round"><path d="M455 87h90m-64 8h99M727 131h106m-76 8h118M99 130h78"/></g>
    <g v-if="scene==='sea'" class="scene-water">
      <path d="M0 236Q150 208 300 236T600 236T900 236T1200 236V360H0Z" fill="var(--ridge-back)"/>
      <path class="wave wave-one" d="M-150 277Q0 246 150 277T450 277T750 277T1050 277T1350 277V360H-150Z" fill="var(--ridge-mid)"/>
      <path class="wave wave-two" d="M-150 314Q0 290 150 314T450 314T750 314T1050 314T1350 314V360H-150Z" fill="var(--ridge-front)"/>
      <path d="m525 193 8 5 8-5m27 15 6 4 6-4" fill="none" stroke="#d9e8dc" stroke-opacity=".55" stroke-width="1.5"/>
      <g class="scene-boat"><path d="M929 232h36l-7 8h-23z" fill="#dbc5ae" opacity=".65"/><path d="M946 208v24m2-23 14 19h-14" fill="#e3d4b9" opacity=".65"/></g>
    </g>
    <g v-else-if="scene!=='cyber'">
      <path d="M0 261 131 209 223 243 377 157 494 236 623 194 772 250 917 168 1089 236 1200 199V360H0Z" fill="var(--ridge-back)"/>
      <path d="M0 302 125 252 263 291 394 228 528 272 672 241 809 283 1007 214 1200 281V360H0Z" fill="var(--ridge-mid)"/>
      <path d="M0 324Q181 287 316 321T600 317T900 321T1200 309V360H0Z" fill="var(--ridge-front)"/>
      <path v-if="scene==='journey'" d="M710 360c-114-36 37-32-10-49s-57-8-27-29" fill="none" stroke="#c8a78f" stroke-opacity=".27" stroke-width="5"/>
      <g v-if="scene==='forest'" fill="var(--tree-color)" opacity=".75"><path d="m922 190-19 46h12l-20 34h18l-20 31h58l-20-31h18l-20-34h12Z"/><path d="m1019 229-14 34h9l-15 25h13l-15 23h43l-15-23h13l-15-25h9Z"/><path d="m1094 168-24 58h16l-26 43h22l-25 39h74l-25-39h22l-26-43h16Z"/></g>
      <g v-if="scene==='romance'" fill="none" stroke="#f8c6c9" stroke-opacity=".45"><path d="m876 178 7 7 7-7a5 5 0 0 0-7-7 5 5 0 0 0-7 7Z"/><path d="m955 111 4 4 4-4a3 3 0 0 0-4-4 3 3 0 0 0-4 4Z"/></g>
      <path v-if="scene==='desert'" d="M0 285Q270 192 490 270T900 265T1200 240V360H0Z" fill="#bb7039"/><path v-if="scene==='desert'" d="M390 360Q580 260 810 279t390-13v94Z" fill="#8b452b"/>
      <g v-if="scene==='snow'" fill="#d9e9ec"><path d="m377 157-50 28 38-5 12 10 12-7 20 4Z"/><path d="m917 168-45 34 30-9 14 9 19-11 19 6Z"/></g>
      <g v-if="scene==='bamboo'" stroke="#a7d3a7" stroke-opacity=".45" fill="none"><path d="M930 360V135m38 225V178m43 182V119" stroke-width="5"/><path d="m930 178-23-21m23 69 24-27m14 49-20-16m63-64-24-22m24 86 28-24" stroke-width="2"/></g>
    </g>
    <g v-if="scene==='snow'||scene==='sakura'" :class="scene==='snow'?'scene-snow':'scene-petals'" :fill="scene==='snow'?'#effaff':'#ffb5d7'"><ellipse v-for="(p,i) in particles" :key="i" :cx="p.x" :cy="p.y" :rx="scene==='snow'?p.r:p.r*2" :ry="p.r" :style="{'--particle-delay':p.delay}"/></g>
    <g v-if="scene==='rain'" class="scene-rain" stroke="#9fbfdc" stroke-opacity=".45" stroke-width="1"><path v-for="(p,i) in particles" :key="i" :d="`M${p.x} ${p.y}l-10 32`" :style="{'--particle-delay':p.delay}"/></g>
    <AmbientEffect :kind="effect"/>
  </svg>
</template>
