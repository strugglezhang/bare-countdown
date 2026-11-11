import test from 'node:test'
import assert from 'node:assert/strict'
import { daylightAt,celestialPosition } from './daylight.js'
const at=(hour,minute=0,scene='journey')=>daylightAt(new Date(2026,10,11,hour,minute).getTime(),scene)

test('sun moves left to right, rises towards noon and falls after noon',()=>{
  const morning=celestialPosition(at(8)),noon=celestialPosition(at(12)),evening=celestialPosition(at(17))
  assert(morning.x<noon.x&&noon.x<evening.x)
  assert(noon.y<morning.y&&noon.y<evening.y)
  assert.equal(at(12).body,'sun')
  assert.equal(at(12).stars,0)
})
test('night moon travels over midnight and sunrise/sunset change bodies',()=>{
  assert.equal(at(5,59).body,'moon')
  assert.equal(at(6).body,'sun')
  assert.equal(at(17,59).body,'sun')
  assert.equal(at(18).body,'moon')
  const early=celestialPosition(at(20)),midnight=celestialPosition(at(0)),late=celestialPosition(at(4))
  assert(early.x<midnight.x&&midnight.x<late.x)
  assert(midnight.y<early.y&&midnight.y<late.y)
})
test('night-themed scenes retain their moon while still moving with time',()=>{
  for(const scene of ['galaxy','aurora','cyber','rain','bamboo']) {
    assert.equal(at(12,0,scene).body,'moon')
    assert.equal(at(12,0,scene).period,'night')
    assert(celestialPosition(at(9,0,scene)).x<celestialPosition(at(15,0,scene)).x)
  }
})
test('dawn/day/dusk/night palettes follow local hours',()=>{
  assert.equal(at(7).period,'dawn')
  assert.equal(at(12).period,'day')
  assert.equal(at(17).period,'dusk')
  assert.equal(at(22).period,'night')
})
test('mobile and fullscreen arcs stay inside the visible slice',()=>{
  for(const [width,height] of [[300,344],[1400,1000],[1100,286],[300,800]]) {
    const visibleWidth=Math.min(1200,360*width/height),left=(1200-visibleWidth)/2
    for(const h of [6,9,12,15,17,18,0,4]) {
      const pos=celestialPosition(at(h),width,height)
      assert(pos.x>left&&pos.x<left+visibleWidth)
      assert(pos.y>=0&&pos.y<=360)
    }
  }
})
test('invalid time rejected and zero-size initial measurement handled',()=>{
  assert.throws(()=>daylightAt(NaN),TypeError)
  assert(Number.isFinite(celestialPosition(at(12),0,0).x))
})
