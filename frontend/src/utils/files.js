export function localFileUrl(path, version = 0) {
    return "/local/file?" + new URLSearchParams({
        path: String(path || ""),
        v: String(version || 0),
    })
}
