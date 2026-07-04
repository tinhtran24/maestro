/// <reference types="vite/client" />

interface ImportMetaEnv {
  // Optional external events WebSocket URL (e.g. ws://127.0.0.1:1421/events).
  // When unset, the desktop app relies on Tauri events and skips the WebSocket.
  readonly VITE_THANOS_EVENTS_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
