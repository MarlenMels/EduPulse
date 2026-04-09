<template>
  <div class="user-card" :class="user.role">
    <!-- avatar shows the first letter of the name -->
    <div class="user-header">
      <div class="avatar">{{ user.name.charAt(0) }}</div>
      <div class="user-info">
        <h4>{{ user.name }}</h4>
        <p class="role">{{ user.role }}</p>
      </div>
    </div>

    <div class="user-details">
      <p>📧 {{ user.email }}</p>
    </div>

    <!-- action buttons: emit events up to the parent -->
    <div class="actions">
      <!-- inline $emit — quick way to fire an event right from the template -->
      <button
        @click="$emit('delete', user.id)"
        class="btn-danger"
      >
        Delete
      </button>

      <!-- calls a handler that emits multiple events at once -->
      <button
        @click="handlePromote"
        class="btn-primary"
      >
        {{ user.role === 'admin' ? 'Demote' : 'Promote' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'

// 1. type for the user object we receive as a prop
interface User {
  id: number
  name: string
  role: 'admin' | 'user' | 'moderator'
  email: string
}

// 2. props — the parent passes a User object into this component
interface Props {
  user: User
}

const props = defineProps<Props>()

// 3. emits — declaring which events this component can send to the parent
//    tuple syntax: EventName: [param1Type, param2Type, ...]
interface Emits {
  delete: [userId: number]
  promote: [userId: number]
  custom: [data: any]
}

const emit = defineEmits<Emits>()

// 4. handler that fires two events when the promote button is clicked
const handlePromote = () => {
  emit('promote', props.user.id)

  // we can also send extra info through a generic "custom" event
  emit('custom', {
    action: 'promote_attempt',
    userId: props.user.id,
    currentRole: props.user.role
  })
}

// 5. computed props — cached, only re-evaluated when dependencies change
const isAdmin = computed(() => props.user.role === 'admin')

const roleColor = computed(() => {
  switch (props.user.role) {
    case 'admin': return '#e74c3c'
    case 'moderator': return '#f39c12'
    default: return '#3498db'
  }
})

// 6. deep watcher — tracks nested changes inside the user object
watch(() => props.user, (newUser) => {
  console.log('User updated:', newUser)
}, { deep: true })
</script>

<style scoped>
.user-card {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 15px;
  background-color: white;
  transition: all 0.3s ease;
}

.user-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.user-card.admin {
  border-left: 4px solid #e74c3c;
}

.user-card.moderator {
  border-left: 4px solid #f39c12;
}

.user-card.user {
  border-left: 4px solid #3498db;
}

.user-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #42b883;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  margin-right: 12px;
}

.user-info h4 {
  margin: 0;
  color: #2c3e50;
}

.role {
  margin: 0;
  font-size: 12px;
  text-transform: uppercase;
  color: #7f8c8d;
}

.user-details {
  margin: 10px 0;
}

.user-details p {
  margin: 5px 0;
  font-size: 14px;
  color: #555;
}

.actions {
  display: flex;
  gap: 8px;
  margin-top: 15px;
}

button {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: background-color 0.2s;
}

.btn-danger {
  background-color: #e74c3c;
  color: white;
}

.btn-danger:hover {
  background-color: #c0392b;
}

.btn-primary {
  background-color: #3498db;
  color: white;
}

.btn-primary:hover {
  background-color: #2980b9;
}
</style>
