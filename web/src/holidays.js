import extraHolidays from './holiday-catalog.json' with {type:'json'}
import {localInput} from './time.js'
import lunarDates from './holiday-dates.json' with {type:'json'}
// Source: Hong Kong Observatory official Gregorian–Lunar calendar tables, 2026–2035.
// https://www.hko.gov.hk/en/gts/time/calendar/text/files/T2027e.txt
export const holidays=[
 {value:'spring',label:'中国新年',icon:'sparkles',description:'中国红 · 灯笼 · 岁岁团圆',title:'一起过个好年',message:'灯火可亲，家人闲坐。新的一年，万事胜意。',theme:'rose',lunar:[1,1]},
 {value:'christmas',label:'圣诞节',icon:'sparkles',description:'松针绿 · 圣诞树 · 金色星光',title:'圣诞快乐，平安喜乐',message:'把愿望挂上树梢，让温暖如约而至。',theme:'mint',month:12,day:25},
 {value:'midautumn',label:'中秋节',icon:'moon',description:'月夜蓝 · 玉兔 · 一轮满月',title:'月圆之时，与你团圆',message:'但愿人长久，千里共婵娟。',theme:'ocean',lunar:[8,15]},
 {value:'dragonboat',label:'端午节',icon:'leaf',description:'竹叶青 · 粽子 · 艾草清香',title:'端午安康，粽有所爱',message:'把思念包进粽叶，把平安送给你。',theme:'mint',lunar:[5,5]},
 {value:'newyear',label:'元旦',icon:'sparkles',description:'香槟金 · 烟花 · 新年心愿',title:'新的一年，闪闪发光',message:'旧事翻篇，下一站尽是好风景。',theme:'amber',month:1,day:1},
 {value:'halloween',label:'万圣节',icon:'moon',description:'南瓜橙 · 小幽灵 · 奇妙夜晚',title:'今晚，一起捣个蛋',message:'装满糖果，也装满古灵精怪的小快乐。',theme:'amber',month:10,day:31},
 ...extraHolidays].map(h=>({group:['spring','midautumn','dragonboat','newyear'].includes(h.value)?'china':'world',...h}))
export const holidayById=id=>holidays.find(h=>h.value===id)
export const holidayNeedsRefresh=id=>{const h=holidayById(id);return !!(h?.lunar||h?.rule||h?.table)}
function nthWeekday(year,month,weekday,nth){const first=new Date(year,month-1,1);return new Date(year,month-1,1+(weekday-first.getDay()+7)%7+(nth-1)*7)}
function easter(year){
 const a=year%19,b=Math.floor(year/100),c=year%100,d=Math.floor(b/4),e=b%4,f=Math.floor((b+8)/25),g=Math.floor((b-f+1)/3),h=(19*a+b-d-g+15)%30,i=Math.floor(c/4),k=c%4,l=(32+2*e+2*i-h-k)%7,m=Math.floor((a+11*h+22*l)/451),month=Math.floor((h+l-7*m+114)/31),day=(h+l-7*m+114)%31+1
 return new Date(year,month-1,day)
}
function ruleDate(h,year){
 if(h.rule==='thanksgiving')return nthWeekday(year,11,4,4)
 if(h.rule==='oktoberfest'){const date=new Date(year,8,16);date.setDate(16+(6-date.getDay()+7)%7);return date}
 const date=easter(year)
 if(h.rule==='carnival')date.setDate(date.getDate()-47)
 return date
}
export function nextHolidayDate(holiday,now=Date.now()){
 const h=holidayById(holiday);if(!h)return ''
 const today=new Date(now)
 if(h.lunar||h.table){
  const dates=Object.values(lunarDates).map(year=>year[holiday]).filter(Boolean).sort()
  const date=dates.find(value=>new Date(value+'T00:00').getTime()>now)
  return date?date+'T00:00':''
 }
 const getDate=year=>h.rule?ruleDate(h,year):new Date(year,h.month-1,h.day)
 let date=getDate(today.getFullYear());if(date.getTime()<=now)date=getDate(today.getFullYear()+1)
 return localInput(date.toISOString())
}
