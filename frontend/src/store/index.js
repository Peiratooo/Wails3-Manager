import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
	state: () => {
		return {
			projectDir: '',
            icons:{},
		}
	},
	actions: {
		setProjectDir(projectDir) {
			this.projectDir = projectDir
		},
	},
})
