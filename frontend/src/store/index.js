import { defineStore } from 'pinia'

export const useAppStore = defineStore('app', {
	state: () => {
		return {
			projectDir: '',
            icons:{},
			settings: {
				isDark: true,
				language: 'zh-CN',
				recordLogs: true,
			},
            env:{
                data:{},
                loaded:false,
            },
			logLines: [],
            panels:{
                settings:false
            }
		}
	},
	actions: {
		setProjectDir(projectDir) {
			this.projectDir = projectDir
		},
		setSettings(settings) {
			this.settings = {
				isDark: settings?.isDark ?? true,
				language: settings?.language || 'zh-CN',
				recordLogs: settings?.recordLogs ?? true,
			}
		},
		updateSettings(partialSettings) {
			this.setSettings({
				...this.settings,
				...partialSettings,
			})
		},
		appendLogLine(line) {
			if (!line) {
				return
			}
			this.logLines.push(line)
			if (this.logLines.length > 500) {
				this.logLines.splice(0, this.logLines.length - 500)
			}
		},
		clearLogLines() {
			this.logLines = []
		},
	},
})
