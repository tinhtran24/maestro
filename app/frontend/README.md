# Thanos App Frontend

This is the React frontend for the Wails-based Thanos workbench in `app/`.

The app loads real workspace metadata through the Wails `RealProvider`. It does
not call a Thanos CLI and does not use Tauri.

## Development

```sh
npm install
npm run dev
```

For the desktop shell:

```sh
npm run wails:dev
```
