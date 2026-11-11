import test from 'node:test'
import assert from 'node:assert/strict'
import {dateKey,monthCells} from './calendar.js'
test('calendar has six Monday-first rows across month boundaries',()=>{
  const days=monthCells(2026,10)
  assert.equal(days.length,42)
  assert.equal(days[0].date.getDay(),1)
  assert.equal(days.filter(d=>d.current).length,30)
  assert.equal(new Set(days.map(d=>d.key)).size,42)
})
test('leap February and December correctly cross years',()=>{
  assert.equal(monthCells(2028,1).filter(d=>d.current).length,29)
  assert.equal(monthCells(2027,1).filter(d=>d.current).length,28)
  assert(monthCells(2026,11).some(d=>d.key.startsWith('2027-01')))
  assert.equal(dateKey(new Date(2026,0,2,12)),'2026-01-02')
})
