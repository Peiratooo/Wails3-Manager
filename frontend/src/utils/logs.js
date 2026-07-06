export function logEntryLevel(entry) {
    const line = String(entry?.line || "").toLowerCase()

    if (/\b(error|failed|failure|fatal|panic|exception|compile aborted|exit status)\b/.test(line)) {
        return "error"
    }
    if (/\b(warn|warning|missing|not found|skipped|disabled)\b/.test(line)) {
        return "warning"
    }
    if (/\b(success|completed|generated|created|available|up to date)\b/.test(line)) {
        return "success"
    }
    if (/\b(running command|working directory|go build|wails3|iscc|create-dmg|npm|task: \[)\b/.test(line)) {
        return "command"
    }
    if (entry?.transactionType === "package") {
        return "package"
    }
    return "info"
}
