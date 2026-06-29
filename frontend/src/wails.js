// Thin wrappers over the bindings Wails injects into the webview at runtime
// (window.go / window.runtime). Using the globals directly means we don't have
// to depend on the generated `wailsjs/` directory.

const app = () => window.go.main.App
const rt = () => window.runtime

export const ListDrives = () => app().ListDrives()
export const ListDriveInfo = () => app().ListDriveInfo()
export const Scan = (path) => app().Scan(path)
export const GetChildren = (id, sortBy, asc) => app().GetChildren(id, sortBy, asc)
export const Cancel = () => app().Cancel()
export const OpenPath = (path) => app().OpenPath(path)

export const onProgress = (cb) => rt().EventsOn('scan:progress', cb)
