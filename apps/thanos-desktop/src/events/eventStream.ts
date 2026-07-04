import type { WorkbenchEvent } from "../domain/models";

type Listener = (event: WorkbenchEvent) => void;

// Optional WebSocket event stream. The desktop app normally receives events via
// Tauri `listen`; this WebSocket is only used when an external events server
// (e.g. the Go `internal/events` server) is configured via VITE_THANOS_EVENTS_URL.
// When no URL is set, `connect()` is a no-op — so no failed-connection noise in
// the console. Errors and closes are handled gracefully with a capped reconnect.
export class WorkbenchEventStream {
    private socket: WebSocket | null = null;
    private listeners = new Set<Listener>();
    private closed = false;
    private retryTimer: ReturnType<typeof setTimeout> | null = null;

    constructor(private readonly url: string) {}

    connect() {
        if (!this.url || this.socket || this.closed) return;
        try {
            const socket = new WebSocket(this.url);
            this.socket = socket;
            socket.addEventListener("message", (message) => {
                try {
                    this.emit(JSON.parse(String(message.data)) as WorkbenchEvent);
                } catch {
                    // Ignore malformed frames.
                }
            });
            // Handle the error event so it never surfaces as uncaught; the browser
            // may still log the network failure, which is why we only connect when
            // a URL is explicitly configured.
            socket.addEventListener("error", () => {});
            socket.addEventListener("close", () => {
                this.socket = null;
                this.scheduleReconnect();
            });
        } catch {
            this.socket = null;
            this.scheduleReconnect();
        }
    }

    private scheduleReconnect() {
        if (this.closed || this.retryTimer || !this.url) return;
        this.retryTimer = setTimeout(() => {
            this.retryTimer = null;
            this.connect();
        }, 5000);
    }

    subscribe(listener: Listener) {
        this.listeners.add(listener);
        return () => {
            this.listeners.delete(listener);
        };
    }

    emit(event: WorkbenchEvent) {
        for (const listener of this.listeners) listener(event);
    }

    disconnect() {
        this.closed = true;
        if (this.retryTimer) {
            clearTimeout(this.retryTimer);
            this.retryTimer = null;
        }
        this.socket?.close();
        this.socket = null;
    }
}
