<script setup>
import { onUnmounted, ref, watch } from 'vue'
import Icon from './Icon.vue'
import ChoiceSelect from './ChoiceSelect.vue'
const soundChoices=[{value:'rain',label:'雨声白噪音'},{value:'waves',label:'海浪白噪音'},{value:'brown',label:'低频棕噪音'}]
const emit=defineEmits(['error'])
const mode=ref('rain'),playing=ref(false),busy=ref(false),volume=ref(30)
let context,source,filter,gain,oscillator,modulation,generation=0
function release(){
  for(const node of [source,oscillator]){try{node?.stop()}catch{}}
  for(const node of [source,filter,gain,oscillator,modulation]){try{node?.disconnect()}catch{}}
  source=filter=gain=oscillator=modulation=undefined
}
function level(){if(!context||!gain)return;const value=Number(volume.value)/100*.08;gain.gain.setTargetAtTime(value*(mode.value==='waves'?0.65:1),context.currentTime,.08);modulation?.gain.setTargetAtTime(value*.35,context.currentTime,.08)}
function stop(){generation++;release();playing.value=false;busy.value=false;if(context?.state==='running')context.suspend().catch(()=>{})}
async function start(){
  const ticket=++generation;busy.value=true;release();playing.value=false
  try{
    const Audio=window.AudioContext||window.webkitAudioContext
    if(!Audio)throw Error('当前浏览器不支持白噪音')
    context ||= new Audio();await context.resume();if(ticket!==generation)return
    const buffer=context.createBuffer(1,context.sampleRate*3,context.sampleRate),data=buffer.getChannelData(0)
    let brown=0
    for(let i=0;i<data.length;i++){const white=Math.random()*2-1;brown=(brown+white*.02)/1.02;data[i]=mode.value==='brown'?Math.max(-1,Math.min(1,brown*3.5)):white}
    source=context.createBufferSource();source.buffer=buffer;source.loop=true
    filter=context.createBiquadFilter();filter.type='lowpass';filter.frequency.value=mode.value==='rain'?5000:mode.value==='waves'?900:450
    gain=context.createGain();source.connect(filter);filter.connect(gain);gain.connect(context.destination)
    if(mode.value==='waves'){oscillator=context.createOscillator();oscillator.frequency.value=.12;modulation=context.createGain();oscillator.connect(modulation);modulation.connect(gain.gain);oscillator.start()}
    level();source.start();playing.value=true
  }catch(error){release();playing.value=false;emit('error',error.message||'声音暂时无法开启')}
  finally{if(ticket===generation)busy.value=false}
}
watch(volume,level)
function changeMode(value){mode.value=value;if(playing.value)start()}
onUnmounted(()=>{stop();context?.close().catch(()=>{})})
defineExpose({stop})
</script>
<template>
  <div class="sound-player" :data-playing="playing">
    <button class="sound-toggle" :disabled="busy" :aria-label="playing?'关闭白噪音':'播放白噪音'" :aria-pressed="playing" @click="playing?stop():start()"><Icon :name="playing?'pause':'play'" :size="17"/></button>
    <div class="sound-choice"><span>白噪音 · 手动播放</span><ChoiceSelect :model-value="mode" :options="soundChoices" label="白噪音类型" icon="volume" compact inverse :disabled="busy" @update:model-value="changeMode"/></div>
    <label class="sound-volume"><Icon name="volume" :size="15"/><input v-model="volume" type="range" min="0" max="100" aria-label="白噪音音量"/><output>{{volume}}%</output></label>
  </div>
</template>
