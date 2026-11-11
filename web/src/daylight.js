const nightScenes = new Set(['galaxy','aurora','cyber','rain','bamboo'])

// A local-time illustration, not a location-based astronomical calculation.
// The day arc runs 06:00–18:00; the night arc runs 18:00–06:00.
export function daylightAt(timestamp, scene = 'journey') {
  const date = new Date(timestamp)
  if (!Number.isFinite(date.getTime())) throw new TypeError('Invalid scene timestamp')
  const hour = date.getHours() + date.getMinutes()/60 + date.getSeconds()/3600 + date.getMilliseconds()/3600000
  const daytime = hour >= 6 && hour < 18
  const progress = daytime ? (hour-6)/12 : ((hour+6)%24)/12
  const fixedNight = nightScenes.has(scene)
  const body = fixedNight || !daytime ? 'moon' : 'sun'
  const period = fixedNight ? 'night' : hour >= 5 && hour < 8 ? 'dawn' : hour >= 16 && hour < 19 ? 'dusk' : daytime ? 'day' : 'night'
  return { hour, progress, body, period, stars:fixedNight||!daytime?1:Math.max(0,1-Math.sin(Math.PI*progress)*5) }
}

// SVG uses xMidYMid slice. Keep the arc inside the part of the drawing that
// actually remains visible in a tall phone banner or fullscreen landscape.
export function celestialPosition(state, width = 1200, height = 360) {
  const ratio = width > 0 && height > 0 ? width/height : 1200/360
  const visibleWidth = Math.min(1200,360*ratio)
  const visibleHeight = Math.min(360,1200/ratio)
  const left = (1200-visibleWidth)/2
  const top = (360-visibleHeight)/2
  const margin = Math.min(visibleWidth*.25,Math.max(48,visibleWidth*.1))
  return {
    x:left+margin+(visibleWidth-margin*2)*state.progress,
    y:top+visibleHeight*(.75-.55*Math.sin(Math.PI*state.progress)),
  }
}
