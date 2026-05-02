<template>
  <div class="simple-card" :class="{ important: isImportant }">
    <h3>{{ title }}</h3>
    <p>{{ content }}</p>
    <span v-if="isImportant" class="badge">⭐ Important</span>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
  content?: string
  isImportant?: boolean
}

withDefaults(defineProps<Props>(), {
  content: 'Default content',
  isImportant: false
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
