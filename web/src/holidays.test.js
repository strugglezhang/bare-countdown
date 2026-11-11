import test from 'node:test'
import assert from 'node:assert/strict'
import {holidays,nextHolidayDate,holidayNeedsRefresh} from './holidays.js'
test('solar holiday defaults advance to the next local calendar date',()=>{
 const now=new Date(2026,11,26,12).getTime()
 assert.equal(nextHolidayDate('christmas',now),'2027-12-25T00:00')
 assert.equal(nextHolidayDate('newyear',now),'2027-01-01T00:00')
 assert.equal(nextHolidayDate('unknown',now),'')
})
test('official lunar dates avoid ICU offset errors and advance after a holiday',()=>{
 const now=new Date(2026,10,11,12).getTime()
 assert.equal(nextHolidayDate('spring',now),'2027-02-06T00:00')
 assert.equal(nextHolidayDate('dragonboat',now),'2027-06-09T00:00')
 assert.equal(nextHolidayDate('midautumn',now),'2027-09-15T00:00')
 assert.equal(nextHolidayDate('spring',new Date(2027,1,7).getTime()),'2028-01-26T00:00')
 assert.equal(nextHolidayDate('spring',new Date(2040,0,1).getTime()),'')
 for(const h of holidays.filter(h=>h.lunar))assert.ok(nextHolidayDate(h.value,now))
})

test('curated catalog keeps only the requested twenty festivals',()=>{
 assert.equal(holidays.length,20)
 assert.equal(new Set(holidays.map(h=>h.value)).size,20)
 assert.deepEqual(holidays.filter(h=>h.group==='world').map(h=>h.value).sort(),['christmas','halloween','oktoberfest','thanksgiving','carnival','dayofdead','sanfermin','gion'].sort())
 assert.equal(holidays.filter(h=>h.group==='china').length,12)
})
test('floating world festivals calculate next occurrence without Gregorian annual repetition',()=>{
 const now=new Date(2026,10,11,12).getTime()
 assert.equal(nextHolidayDate('thanksgiving',now),'2026-11-26T00:00')
 assert.equal(nextHolidayDate('thanksgiving',new Date(2026,10,27).getTime()),'2027-11-25T00:00')
 assert.equal(nextHolidayDate('oktoberfest',now),'2027-09-18T00:00')
 assert.equal(nextHolidayDate('carnival',now),'2027-02-09T00:00')
 for(const id of ['thanksgiving','oktoberfest','carnival','lantern','qingming','eve'])assert.equal(holidayNeedsRefresh(id),true)
 for(const id of ['gion','sanfermin','christmas'])assert.equal(holidayNeedsRefresh(id),false)
 assert.equal(nextHolidayDate('lantern',now),'2027-02-20T00:00')
 assert.equal(nextHolidayDate('qingming',now),'2027-04-05T00:00')
 assert.equal(nextHolidayDate('qixi',now),'2027-08-08T00:00')
 assert.equal(nextHolidayDate('laba',now),'2027-01-15T00:00')
 assert.equal(nextHolidayDate('eve',now),'2027-02-05T00:00')
})
