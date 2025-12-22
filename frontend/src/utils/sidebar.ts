import { ref, watch } from 'vue'

const COLLAPSED_KEY = 'sidebar-collapsed'
const WIDTH_KEY = 'sidebar-width'

export const collapsed = ref(localStorage.getItem(COLLAPSED_KEY) === 'true')
export const sidebarWidth = ref(parseInt(localStorage.getItem(WIDTH_KEY) || '260', 10))

export const toggleSidebar = () => {
    collapsed.value = !collapsed.value
    if (!collapsed.value && sidebarWidth.value < 180) {
        sidebarWidth.value = 180
    }
    saveSidebarState()
}

export const saveSidebarState = () => {
    localStorage.setItem(COLLAPSED_KEY, String(collapsed.value))
    localStorage.setItem(WIDTH_KEY, String(sidebarWidth.value))
}

// Ensure state stays in sync across tabs if needed
window.addEventListener('storage', (e) => {
    if (e.key === COLLAPSED_KEY) {
        collapsed.value = e.newValue === 'true'
    } else if (e.key === WIDTH_KEY) {
        sidebarWidth.value = parseInt(e.newValue || '260', 10)
    }
})
