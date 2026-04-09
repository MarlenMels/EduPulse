<template>
  <div class="simple-card" :class="{ important: isImportant }">
    <h3>{{ title }}</h3>
    <p>{{ content }}</p>
    <!-- show a badge only when isImportant is true -->
    <span v-if="isImportant" class="badge">⭐ Important</span>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'

// defining the shape of props this component accepts
interface Props {
  title: string         // required
  content?: string      // optional, has a default
  isImportant?: boolean // optional, defaults to false
}

// withDefaults lets us set fallback values for optional props
const props = withDefaults(defineProps<Props>(), {
  content: 'Default content',
  isImportant: false
})

// we can access props right in <script setup> — runs once on mount
console.log('Card title:', props.title)

// computed property — recalculates automatically when props change
const cardType = computed(() => {
  return props.isImportant ? 'important' : 'regular'
})

// watcher — fires every time props.title changes
watch(() => props.title, (newTitle) => {
  console.log('Title changed to:', newTitle)
})
</script>

<style scoped>
.simple-card {
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 8px;
  margin: 10px 0;
  background-color: white;
  transition: all 0.3s ease;
}

.simple-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.simple-card.important {
  border-color: #f39c12;
  background-color: #fff9e6;
}

.simple-card h3 {
  margin: 0 0 10px 0;
  color: #2c3e50;
}

.simple-card p {
  margin: 0;
  color: #555;
}

.badge {
  display: inline-block;
  background-color: #f39c12;
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
  margin-top: 8px;
}
</style>
