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
			logEntries: [],
			packageTransactions: {},
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
		appendLogEntry(entry) {
			if (!entry?.line) {
				return
			}

			const logEntry = {
				line: entry.line,
				transactionId: entry.transactionId,
				transactionType: entry.transactionType,
				transactionTitle: entry.transactionTitle,
			}

			this.logEntries.push(logEntry)

			if (logEntry.transactionId && logEntry.transactionType === 'package') {
				if (!this.packageTransactions[logEntry.transactionId]) {
					this.packageTransactions[logEntry.transactionId] = {
						id: logEntry.transactionId,
						title: logEntry.transactionTitle,
						entries: [],
					}
				}
				this.packageTransactions[logEntry.transactionId].entries.push(logEntry)
			}

			this.trimLogs()
		},
		clearLogLines() {
			this.logEntries = []
			this.packageTransactions = {}
		},
		trimLogs() {
			if (this.logEntries.length <= 500) {
				return
			}
			const removed = this.logEntries.splice(0, this.logEntries.length - 500)
			for (const entry of removed) {
				if (entry.transactionId && this.packageTransactions[entry.transactionId]) {
					this.packageTransactions[entry.transactionId].entries.shift()
					if (!this.packageTransactions[entry.transactionId].entries.length) {
						delete this.packageTransactions[entry.transactionId]
					}
				}
			}
		},
	},
})
