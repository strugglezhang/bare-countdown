<script setup>
import {useId} from 'vue'
const uid=`hourglass-${useId().replace(/[^a-z0-9-]/gi,'')}`
</script>
<template>
  <svg class="brand-icon" viewBox="0 0 48 48" fill="none" aria-hidden="true">
    <defs>
      <linearGradient :id="uid" x1="6" y1="2" x2="43" y2="48" gradientUnits="userSpaceOnUse"><stop stop-color="#64518f"/><stop offset="1" stop-color="#211c36"/></linearGradient>
      <linearGradient :id="`${uid}-gold`" x1="12" y1="8" x2="36" y2="40" gradientUnits="userSpaceOnUse"><stop stop-color="#fff0cd"/><stop offset=".48" stop-color="#c9a369"/><stop offset="1" stop-color="#f7dfad"/></linearGradient>
    </defs>
    <rect x="1" y="1" width="46" height="46" rx="14" :fill="`url(#${uid})`"/>
    <rect x="3" y="3" width="42" height="42" rx="12" stroke="#f1dbac" stroke-opacity=".22"/>
    <path d="M15 11v5c0 3 3 5 7 8-4 3-7 5-7 8v5h18v-5c0-3-3-5-7-8 4-3 7-5 7-8v-5Z" fill="#d9ccf0" fill-opacity=".1" stroke="#e6daf8" stroke-width="1.2" stroke-linejoin="round"/>
    <path d="M18 15h12c-1 3-4 5-6 6-2-1-5-3-6-6Zm-1 20c2-2 5-6 7-6s5 4 7 6Z" :fill="`url(#${uid}-gold)`"/>
    <path d="M24 23v7" class="brand-sand" stroke="#fce2b1" stroke-width="1.2" stroke-dasharray="1 2" stroke-linecap="round"/>
    <path d="M12 12v24m24-24v24" :stroke="`url(#${uid}-gold)`" stroke-width="1.4"/>
    <path d="M12 9h24M12 39h24" :stroke="`url(#${uid}-gold)`" stroke-width="3.5" stroke-linecap="round"/>
    <path d="M18 13v3c0 2 1 3 3 5" stroke="white" stroke-opacity=".5" stroke-linecap="round"/>
    <path d="m39 15 1 3 3 1-3 1-1 3-1-3-3-1 3-1Z" fill="#f7e3b9"/>
  </svg>
</template>
