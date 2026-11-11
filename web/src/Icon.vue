<script setup>
defineProps({ name: String, size: { type: Number, default: 20 } })
const paths = {
  hourglass: 'M6 3h12M6 21h12M7 3v5l5 4 5-4V3M7 21v-5l5-4 5 4v5',
  grid: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
  calendar: 'M5 5h14a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2ZM7 3v4M17 3v4M3 11h18M8 15h2M14 15h2',
  heart: 'M20.8 4.6a5.5 5.5 0 0 0-7.8 0l-1 1-1-1a5.5 5.5 0 0 0-7.8 7.8L12 21l8.8-8.6a5.5 5.5 0 0 0 0-7.8Z',
  target: 'M21 12a9 9 0 1 1-9-9M17 12a5 5 0 1 1-5-5M12 12l9-9M16 3h5v5',
  clock: 'M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0ZM12 7v5l3 2',
  plus: 'M12 5v14M5 12h14', search: 'M21 21l-5-5M18 10a8 8 0 1 1-16 0 8 8 0 0 1 16 0Z',
  pin: 'm16 3 5 5-4 1-4 5v4l-3-3-6 6 6-6-3-3h4l5-4z',
  sparkles: 'm12 3 2.3 6.7L21 12l-6.7 2.3L12 21l-2.3-6.7L3 12l6.7-2.3ZM20 2v4M18 4h4',
  circle: 'M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z', line: 'M3 8h18M3 12h12M3 16h7',
  chevron: 'm6 9 6 6 6-6', arrow: 'M4 12h16m-6-6 6 6-6 6', close: 'm6 6 12 12M6 18 18 6',
  play: 'm8 5 11 7-11 7Z', pause: 'M8 5v14M16 5v14', reset: 'M3 11a9 9 0 1 1 2 7M3 4v7h7',
  edit: 'm16 3 5 5-12 12-6 1 1-6ZM14 5l5 5', trash: 'M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7M14 10v7',
  expand: 'M8 3H3v5M16 3h5v5M3 16v5h5M21 16v5h-5',
  download: 'M12 3v12m-5-5 5 5 5-5M4 16v5h16v-5', upload: 'M12 16V4m-5 5 5-5 5 5M4 16v5h16v-5',
  check: 'm5 12 4 4L19 6', moon: 'M21 13A9 9 0 0 1 11 3a9 9 0 1 0 10 10Z', sun: 'M16 12a4 4 0 1 1-8 0 4 4 0 0 1 8 0ZM12 2v2M12 20v2M2 12h2M20 12h2M5 5l1 1M18 18l1 1M5 19l1-1M18 6l1-1',
  bell: 'M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4',
  volume: 'm11 4-6 5H2v6h3l6 5ZM15 8a5 5 0 0 1 0 8M18 5a9 9 0 0 1 0 14',
  list: 'M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01',
  plane: 'm21 3-7 18-4-7-7-4ZM10 14 21 3', coffee: 'M3 8h13v9a3 3 0 0 1-3 3H6a3 3 0 0 1-3-3ZM16 9h2a3 3 0 0 1 0 6h-2M7 2v3M12 2v3',
  briefcase: 'M3 7h18v14H3ZM8 7V3h8v4M3 12l9 3 9-3M12 12v5',
  leaf: 'M20 3c-11-1-18 4-15 12 7 7 17 0 15-12ZM4 21 16 8',
}
</script>
<template><svg :width="size" :height="size" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.65" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path :d="paths[name] || paths.sparkles" /></svg></template>
