import test from 'node:test'
import assert from 'node:assert/strict'
import { timeInfo, localInput } from './time.js'

const now = Date.parse('2026-11-10T08:00:00Z')
const base = { kind:'countdown', recurrence:'none', target_at:new Date(now+90061000).toISOString(), created_at:new Date(now).toISOString(), duration_seconds:120 }
test('break down remaining time without fractional second drift',()=>{
  const t=timeInfo(base,now)
  assert.deepEqual([t.days,t.hours,t.minutes,t.secs],[1,1,1,1])
  assert.equal(timeInfo({...base,target_at:new Date(now+1).toISOString()},now).seconds,1)
})
test('expired countdown stops at zero, anniversaries count up',()=>{
  const past={...base,target_at:new Date(now-172801000).toISOString()}
  assert.equal(timeInfo(past,now).reached,true)
  assert.equal(timeInfo({...past,kind:'anniversary'},now).days,2)
  assert.equal(timeInfo({...past,kind:'anniversary'},now).elapsed,true)
})
test('paused timer remains stable when clock advances',()=>{
  const timer={...base,kind:'timer',paused_remaining:87}
  assert.equal(timeInfo(timer,now).seconds,87)
  assert.equal(timeInfo(timer,now+7200000).seconds,87)
})
test('server-provided next recurring target takes precedence',()=>{
  const e={...base,target_at:new Date(now-86400000).toISOString(),effective_target_at:new Date(now+3600000).toISOString(),recurrence:'daily'}
  assert.equal(timeInfo(e,now).seconds,3600)
})
test('local input round trips in browser timezone',()=>{
  const value=localInput('2026-11-10T10:30:00Z')
  assert.equal(new Date(value).getTime(),Date.parse('2026-11-10T10:30:00Z'))
})
