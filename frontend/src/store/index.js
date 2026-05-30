import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
	state: () => {
		return {
			projectDir: '',
            icons:{},
			themeName: '',
		}
	},
	actions: {
		setProjectDir(projectDir) {
			this.projectDir = projectDir
		},
		setThemeName(themeName) {
			this.themeName = themeName
		},
	},
})
