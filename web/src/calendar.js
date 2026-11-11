export function dateKey(date) {
  const pad=n=>String(n).padStart(2,'0')
  return `${date.getFullYear()}-${pad(date.getMonth()+1)}-${pad(date.getDate())}`
}

export function monthCells(year,month) {
  const first=new Date(year,month,1,12)
  const offset=(first.getDay()+6)%7
  return Array.from({length:42},(_,i)=>{
    const date=new Date(year,month,1-offset+i,12)
    return {date,key:dateKey(date),day:date.getDate(),current:date.getMonth()===month}
  })
}
