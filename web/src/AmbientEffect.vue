<script setup>
defineProps({kind:{type:String,default:'flow'}})
const particles=Array.from({length:28},(_,i)=>({x:(i*197+43)%1200,y:(i*79+31)%360,r:1+(i%3)*.65,delay:`${-(i%11)*.7}s`,duration:`${7+i%6}s`}))
</script>
<template>
  <g class="ambient-effects" :data-effect="kind" aria-hidden="true">
    <g v-if="kind==='meteor'" class="effect-meteors" stroke="#e6e4ff" stroke-width="1.6" fill="none"><path v-for="i in 4" :key="i" :d="`M${240+i*190} ${25+i*28}l-95 48`" :style="{'--fx-delay':`${-i*2.3}s`}"/></g>
    <g v-else-if="kind==='firefly'" class="effect-fireflies" fill="#e7ef9f"><circle v-for="(p,i) in particles" :key="i" :cx="p.x" :cy="p.y" :r="p.r" :style="{'--fx-delay':p.delay,'--fx-duration':p.duration}"/></g>
    <g v-else-if="kind==='rain'" class="effect-rain" stroke="#c1dbec" stroke-opacity=".45" stroke-width="1"><path v-for="(p,i) in particles" :key="i" :d="`M${p.x} ${p.y}l-12 32`" :style="{'--fx-delay':p.delay}"/></g>
    <g v-else-if="kind==='snow'" class="effect-snow" fill="#f1f7ff"><circle v-for="(p,i) in particles" :key="i" :cx="p.x" :cy="p.y" :r="p.r" :style="{'--fx-delay':p.delay,'--fx-duration':p.duration}"/></g>
    <g v-else-if="kind==='petals'" class="effect-petals" fill="#ffbedb"><ellipse v-for="(p,i) in particles" :key="i" :cx="p.x" :cy="p.y" :rx="p.r*2.2" :ry="p.r" :style="{'--fx-delay':p.delay,'--fx-duration':p.duration}"/></g>
  </g>
</template>
