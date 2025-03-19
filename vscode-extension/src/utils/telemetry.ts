import * as vscode from 'vscode';
import * as os from 'os';
import axios from 'axios';
import * as md5 from 'md5';

/**
 * Service for collecting anonymous telemetry
 */
export class TelemetryService {
    private static instance: TelemetryService;
    private context: vscode.ExtensionContext;
    private machineId: string;
    private sessionId: string;
    private enabled: boolean;
    private queue: any[] = [];
    private flushInterval: NodeJS.Timeout | null = null;

    private constructor(context: vscode.ExtensionContext) {
        this.context = context;
        this.machineId = this.getMachineId();
        this.sessionId = this.generateSessionId();
        this.enabled = this.isTelemetryEnabled();

        // Set up flush interval
        this.startFlushInterval();

        // Listen for configuration changes
        vscode.workspace.onDidChangeConfiguration(e => {
            if (e.affectsConfiguration('codehawk.telemetryEnabled')) {
                this.enabled = this.isTelemetryEnabled();

                if (this.enabled) {
                    this.startFlushInterval();
                } else {
                    this.stopFlushInterval();
                }
            }
        });
    }

    /**
     * Get the TelemetryService instance
     */
    public static getInstance(context: vscode.ExtensionContext): TelemetryService {
        if (!TelemetryService.instance) {
            TelemetryService.instance = new TelemetryService(context);
        }
        return TelemetryService.instance;
    }

    /**
     * Send a telemetry event
     */
    public sendEvent(eventName: string, properties?: Record<string, any>): void {
        if (!this.enabled) {
            return;
        }

        // Don't collect PII data
        const sanitizedProperties = this.sanitizeProperties(properties || {});

        const event = {
            eventName,
            properties: sanitizedProperties,
            machineId: this.machineId,
            sessionId: this.sessionId,
            timestamp: new Date().toISOString(),
            extensionVersion: this.getExtensionVersion(),
            vscodeVersion: vscode.version,
            osType: os.type(),
            osRelease: os.release(),
        };

        this.queue.push(event);

        // If queue gets too large, flush it
        if (this.queue.length >= 20) {
            this.flush();
        }
    }

    /**
     * Flush the event queue
     */
    public flush(): void {
        if (!this.enabled || this.queue.length === 0) {
            return;
        }

        const events = [...this.queue];
        this.queue = [];

        // Send to API endpoint
        const config = vscode.workspace.getConfiguration('codehawk');
        const apiUrl = config.get<string>('apiUrl');

        if (!apiUrl) {
            return;
        }

        // Don't await the response, fire and forget
        axios.post(`${apiUrl}/api/v1/telemetry`, {
            events
        }).catch(error => {
            // Just log to console, don't bother the user
            console.error('Failed to send telemetry:', error);
        });
    }

    /**
     * Dispose the telemetry service
     */
    public dispose(): void {
        this.stopFlushInterval();
        this.flush();
    }

    /**
     * Start the flush interval
     */
    private startFlushInterval(): void {
        if (this.flushInterval) {
            clearInterval(this.flushInterval);
        }

        if (this.enabled) {
            this.flushInterval = setInterval(() => {
                this.flush();
            }, 5 * 60 * 1000); // Flush every 5 minutes
        }
    }

    /**
     * Stop the flush interval
     */
    private stopFlushInterval(): void {
        if (this.flushInterval) {
            clearInterval(this.flushInterval);
            this.flushInterval = null;
        }
    }

    /**
     * Get machine ID (anonymized)
     */
    private getMachineId(): string {
        // Get a unique but anonymous machine ID
        const id = `${os.hostname()}-${os.userInfo().username}-${os.platform()}`;
        return md5(id);
    }

    /**
     * Generate a session ID
     */
    private generateSessionId(): string {
        return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
    }

    /**
     * Check if telemetry is enabled
     */
    private isTelemetryEnabled(): boolean {
        const config = vscode.workspace.getConfiguration('codehawk');
        return config.get<boolean>('telemetryEnabled', true);
    }

    /**
     * Get extension version
     */
    private getExtensionVersion(): string {
        const extension = vscode.extensions.getExtension('codehawk.codehawk');
        return extension ? extension.packageJSON.version : '0.0.0';
    }

    /**
     * Sanitize properties to remove PII
     */
    private sanitizeProperties(properties: Record<string, any>): Record<string, any> {
        const sanitized: Record<string, any> = {};

        for (const [key, value] of Object.entries(properties)) {
            // Skip properties that might contain PII
            if (key.toLowerCase().includes('path') ||
                key.toLowerCase().includes('file') ||
                key.toLowerCase().includes('name') ||
                key.toLowerCase().includes('email') ||
                key.toLowerCase().includes('user')) {
                continue;
            }

            // Sanitize string values
            if (typeof value === 'string') {
                // If it looks like a path or contains an email, skip it
                if (value.includes('/') || value.includes('\\') || value.includes('@')) {
                    continue;
                }
                sanitized[key] = value;
            }
            // Include primitive values as-is
            else if (typeof value === 'number' || typeof value === 'boolean') {
                sanitized[key] = value;
            }
            // For arrays and objects, include only their type and size
            else if (Array.isArray(value)) {
                sanitized[key] = `array:${value.length}`;
            }
            else if (typeof value === 'object' && value !== null) {
                sanitized[key] = `object:${Object.keys(value).length}`;
            }
        }

        return sanitized;
    }
}