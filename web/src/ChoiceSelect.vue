<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, useId } from 'vue'
import Icon from './Icon.vue'
const props=defineProps({modelValue:String,options:Array,label:String,icon:{type:String,default:'grid'},allowCustom:Boolean,disabled:Boolean,compact:Boolean,inverse:Boolean})
const emit=defineEmits(['update:modelValue'])
const root=ref(null),trigger=ref(null),panel=ref(null),target=ref(null),open=ref(false),active=ref(0),custom=ref(''),placement=ref({})
const uid=useId().replace(/[^a-z0-9-]/gi,'')
const options=computed(()=>{const list=(props.options||[]).map(option=>typeof option==='string'?{value:option,label:option}:option);if(props.allowCustom&&props.modelValue&&!list.some(option=>option.value===props.modelValue))list.push({value:props.modelValue,label:props.modelValue});return list})
const selected=computed(()=>options.value.find(option=>option.value===props.modelValue)?.label||props.modelValue||'请选择')
const canAdd=computed(()=>custom.value.trim().length>0&&!options.value.some(option=>option.value===custom.value.trim()))
const unavailable=computed(()=>props.disabled||(!options.value.length&&!props.allowCustom))
function position(){
  if(!trigger.value)return
  const rect=trigger.value.getBoundingClientRect(),dialog=root.value.closest('dialog')?.getBoundingClientRect()
  const leftLimit=Math.max(8,dialog?.left+8||8),rightLimit=Math.min(window.innerWidth-8,dialog?dialog.right-8:window.innerWidth-8)
  const topLimit=Math.max(8,dialog?.top+8||8),bottomLimit=Math.min(window.innerHeight-8,dialog?dialog.bottom-8:window.innerHeight-8)
  const width=Math.min(Math.max(rect.width,props.compact?190:210),rightLimit-leftLimit)
  const content=Math.min(240,options.value.length*(options.value.some(o=>o.description)?60:40))+(props.allowCustom?88:12)
  const below=bottomLimit-rect.bottom-7,above=rect.top-topLimit-7,up=below<content&&above>below
  const maxHeight=Math.max(70,Math.min(content,up?above:below)),height=Math.min(panel.value?.scrollHeight||content,maxHeight)
  placement.value={position:'fixed',left:`${Math.max(leftLimit,Math.min(rect.left,rightLimit-width))}px`,top:`${up?rect.top-7-height:rect.bottom+7}px`,width:`${width}px`,maxHeight:`${maxHeight}px`}
}
async function show(){
  if(unavailable.value)return
  target.value=root.value.closest('dialog')||document.body
  active.value=Math.max(0,options.value.findIndex(o=>o.value===props.modelValue));custom.value='';position();open.value=true
  await nextTick();position();document.getElementById(`${uid}-${active.value}`)?.scrollIntoView({block:'nearest'})
}
function choose(value){emit('update:modelValue',value);open.value=false;trigger.value?.focus()}
function add(){if(canAdd.value)choose(custom.value.trim())}
function key(event){
  if(event.key==='Escape'&&open.value){event.preventDefault();event.stopPropagation();open.value=false;trigger.value?.focus();return}
  if(event.target.tagName==='INPUT')return
  if(event.key==='Tab'){open.value=false;return}
  if(!['ArrowDown','ArrowUp','Home','End','Enter',' '].includes(event.key))return
  event.preventDefault()
  if(!open.value){show();return}
  if(!options.value.length)return
  if(event.key==='Enter'||event.key===' '){choose(options.value[active.value].value);return}
  if(event.key==='Home')active.value=0
  else if(event.key==='End')active.value=options.value.length-1
  else active.value=(active.value+(event.key==='ArrowDown'?1:-1)+options.value.length)%options.value.length
  nextTick(()=>document.getElementById(`${uid}-${active.value}`)?.scrollIntoView({block:'nearest'}))
}
function outside(event){if(!root.value?.contains(event.target)&&!panel.value?.contains(event.target))open.value=false}
function reposition(){if(open.value)position()}
onMounted(()=>{document.addEventListener('pointerdown',outside);window.addEventListener('resize',reposition);window.addEventListener('scroll',reposition,true)})
onUnmounted(()=>{document.removeEventListener('pointerdown',outside);window.removeEventListener('resize',reposition);window.removeEventListener('scroll',reposition,true)})
</script>
<template>
  <div ref="root" class="choice-select" :class="{opened:open,disabled:unavailable,'choice-compact':compact,'choice-inverse':inverse}" @keydown="key">
    <button ref="trigger" type="button" role="combobox" class="choice-trigger" :aria-label="label" aria-haspopup="listbox" :aria-controls="`${uid}-list`" :aria-expanded="open" :aria-activedescendant="open?`${uid}-${active}`:undefined" :disabled="unavailable" :title="selected" @click="open?open=false:show()"><span class="choice-leading"><Icon :name="icon" :size="15"/></span><span>{{selected}}</span><Icon name="chevron" class="choice-chevron" :size="14"/></button>
    <Teleport v-if="open" :to="target||'body'"><div ref="panel" class="choice-panel choice-floating" :class="{'choice-inverse':inverse}" :style="placement" @keydown="key">
      <div :id="`${uid}-list`" role="listbox" :aria-label="`${label}选项`" class="choice-options"><button v-for="(option,index) in options" :id="`${uid}-${index}`" :key="option.value" type="button" role="option" tabindex="-1" :aria-label="option.label" :aria-selected="modelValue===option.value" class="choice-option" :class="{selected:modelValue===option.value,highlighted:active===index}" @mouseenter="active=index" @click="choose(option.value)"><Icon v-if="option.icon" :name="option.icon" :size="15" class="choice-option-icon"/><span v-else class="choice-dot" :style="{'--choice-color':['#8c62ce','#d99e4a','#ce7598','#4e9b84'][index%4]}"/><span class="choice-option-copy"><strong>{{option.label}}</strong><small v-if="option.description">{{option.description}}</small></span><Icon v-if="modelValue===option.value" name="check" :size="15"/></button></div>
      <div v-if="allowCustom" class="choice-custom"><span>也可以创建新的分类</span><div><input v-model="custom" aria-label="新分类名称" maxlength="16" placeholder="例如：旅行、学习…" @keydown.enter.prevent="add"/><button type="button" :disabled="!canAdd" aria-label="使用新分类" @click="add"><Icon name="plus" :size="16"/></button></div></div>
    </div></Teleport>
  </div>
</template>
