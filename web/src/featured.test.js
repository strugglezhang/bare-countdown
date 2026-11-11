import test from 'node:test'
import assert from 'node:assert/strict'
import {pickFeatured} from './featured.js'
const now=Date.parse('2026-11-11T04:00:00Z')
const item=(id,overrides={})=>({id,title:id,demo:true,category:'生活',kind:'countdown',recurrence:'none',target_at:new Date(now+3600000).toISOString(),created_at:new Date(now-3600000).toISOString(),pinned:false,...overrides})

test('pinned personal life and anniversary records appear before unpinned examples',()=>{
  for(const category of ['生活','纪念日']){
    const own=item('own',{demo:false,pinned:true,category,target_at:new Date(now+7200000).toISOString()})
    assert.equal(pickFeatured([item('example'),own],now).id,'own')
    own.pinned=false
    assert.equal(pickFeatured([own,item('example')],now).id,'example')
    assert.equal(pickFeatured([own],now),undefined)
  }
})
test('work and focus records do not enter the hero even when pinned',()=>{
  const events=[item('work',{category:'工作',pinned:true}),item('focus',{category:'专注',kind:'timer',pinned:true}),item('timer',{kind:'timer',pinned:true}),item('life')]
  assert.equal(pickFeatured(events,now).id,'life')
})
test('pinned candidates win; ties use the nearest time',()=>{
  const soon=item('soon'),later=item('later',{target_at:new Date(now+7200000).toISOString(),pinned:true})
  assert.equal(pickFeatured([soon,later],now).id,'later')
  later.pinned=false;assert.equal(pickFeatured([later,soon],now).id,'soon')
})
test('count-up anniversaries can be featured; expired countdowns and paused records cannot',()=>{
  const past=new Date(now-3600000).toISOString()
  assert.equal(pickFeatured([item('since',{demo:false,pinned:true,category:'纪念日',kind:'anniversary',target_at:past})],now).id,'since')
  assert.equal(pickFeatured([item('past',{target_at:past,pinned:true}),item('paused',{pinned:true,paused_remaining:3600})],now),undefined)
  assert.equal(pickFeatured([],now),undefined)
})
