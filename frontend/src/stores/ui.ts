import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

const accessibilityKey = 'edupulse_accessible_vision'

function applyAccessibleVision(enabled: boolean) {
  document.documentElement.classList.toggle('vision-accessible', enabled)
}

export const useUiStore = defineStore('ui', () => {
  const accessibleVision = ref(localStorage.getItem(accessibilityKey) === 'true')

  applyAccessibleVision(accessibleVision.value)

  watch(accessibleVision, (enabled) => {
    localStorage.setItem(accessibilityKey, String(enabled))
    applyAccessibleVision(enabled)
  })

  function toggleAccessibleVision() {
    accessibleVision.value = !accessibleVision.value
  }

  return {
    accessibleVision,
    toggleAccessibleVision,
  }
})
