import {timeInfo} from './time.js'

export function pickFeatured(events, now) {
  return events.filter(event=>{
    if((!event.pinned&&!event.demo)||!['生活','纪念日','节日'].includes(event.category)||event.kind==='timer')return false
    const info=timeInfo(event,now)
    return !info.reached&&!info.paused
  }).sort((a,b)=>Number(b.pinned)-Number(a.pinned)||timeInfo(a,now).seconds-timeInfo(b,now).seconds)[0]
}
