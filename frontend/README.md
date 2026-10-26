# Wails3 Manager Frontend

Vue 3 frontend for Wails3 Manager.

## Commands

```bash
npm ci
npm run dev
npm run i18n:check
npm run build
```

The app is run by the root Wails task during normal development:

```bash
cd ..
wails3 task dev
```

The complete `bindings/` directory is committed so a fresh checkout can build the frontend. After changing a Go service, run `wails3 generate bindings` from the repository root and commit the regenerated bindings with the Go changes. The Wails build tasks also regenerate bindings automatically. Do not edit generated files by hand.
