<template>
  <section class="gateway-stat-grid">
    <article v-for="metric in metrics" :key="metric.key" class="card gateway-stat" :data-metric="metric.key">
      <div class="gateway-stat-label">
        <span class="gateway-stat-icon" :data-tone="metric.tone"><Icon :name="metric.icon" size="sm" aria-hidden="true" /></span>
        {{ metric.label }}
      </div>
      <strong>{{ metric.value }}</strong>
      <p v-if="metric.note">{{ metric.note }}</p>
    </article>
  </section>
</template>
<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'

defineProps<{ metrics: { key: string; label: string; value: string; icon: InstanceType<typeof Icon>['$props']['name']; tone?: string; note?: string }[] }>()
</script>
