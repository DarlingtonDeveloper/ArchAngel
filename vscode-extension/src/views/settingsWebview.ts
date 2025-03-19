import * as vscode from 'vscode';
import * as path from 'path';

/**
 * Manages the settings webview panel
 */
export class SettingsWebview {
    public static currentPanel: SettingsWebview | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private readonly _extensionPath: string;
    private _disposables: vscode.Disposable[] = [];

    /**
     * Create or show the settings panel
     */
    public static createOrShow(extensionPath: string): void {
        const column = vscode.window.activeTextEditor
            ? vscode.window.activeTextEditor.viewColumn
            : undefined;

        // If we already have a panel, show it
        if (SettingsWebview.currentPanel) {
            SettingsWebview.currentPanel._panel.reveal(column);
            return;
        }

        // Otherwise, create a new panel
        const panel = vscode.window.createWebviewPanel(
            'codehawkSettings',
            'CodeHawk Settings',
            column || vscode.ViewColumn.One,
            {
                // Enable JavaScript in the webview
                enableScripts: true,
                // Restrict the webview to only loading content from our extension directory
                localResourceRoots: [vscode.Uri.file(path.join(extensionPath, 'media'))]
            }
        );

        SettingsWebview.currentPanel = new SettingsWebview(panel, extensionPath);
    }

    private constructor(panel: vscode.WebviewPanel, extensionPath: string) {
        this._panel = panel;
        this._extensionPath = extensionPath;

        // Set the webview's initial html content
        this._update();

        // Listen for when the panel is disposed
        // This happens when the user closes the panel or when the panel is closed programmatically
        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

        // Handle messages from the webview
        this._panel.webview.onDidReceiveMessage(
            message => {
                switch (message.command) {
                    case 'saveSettings':
                        this._saveSettings(message.settings);
                        return;
                    case 'resetSettings':
                        this._resetSettings();
                        return;
                }
            },
            null,
            this._disposables
        );
    }

    /**
     * Clean up resources
     */
    public dispose(): void {
        SettingsWebview.currentPanel = undefined;

        // Clean up our resources
        this._panel.dispose();

        while (this._disposables.length) {
            const x = this._disposables.pop();
            if (x) {
                x.dispose();
            }
        }
    }

    /**
     * Update the webview content
     */
    private _update(): void {
        const webview = this._panel.webview;
        this._panel.title = "CodeHawk Settings";
        this._panel.webview.html = this._getHtmlForWebview(webview);
    }

    /**
     * Save settings from the webview
     */
    private _saveSettings(settings: any): void {
        const config = vscode.workspace.getConfiguration('codehawk');

        // Update settings
        for (const [key, value] of Object.entries(settings)) {
            config.update(key, value, vscode.ConfigurationTarget.Global);
        }

        vscode.window.showInformationMessage('CodeHawk settings saved successfully');
    }

    /**
     * Reset settings to defaults
     */
    private _resetSettings(): void {
        const config = vscode.workspace.getConfiguration('codehawk');

        // Reset to defaults
        config.update('apiUrl', 'http://localhost:8080', vscode.ConfigurationTarget.Global);
        config.update('apiKey', '', vscode.ConfigurationTarget.Global);
        config.update('autoAnalyze', false, vscode.ConfigurationTarget.Global);
        config.update('showInlineIssues', true, vscode.ConfigurationTarget.Global);

        this._update();
        vscode.window.showInformationMessage('CodeHawk settings reset to defaults');
    }

    /**
     * Get the HTML content for the webview
     */
    private _getHtmlForWebview(webview: vscode.Webview): string {
        // Get current settings
        const config = vscode.workspace.getConfiguration('codehawk');
        const apiUrl = config.get<string>('apiUrl') || 'http://localhost:8080';
        const apiKey = config.get<string>('apiKey') || '';
        const autoAnalyze = config.get<boolean>('autoAnalyze') || false;
        const showInlineIssues = config.get<boolean>('showInlineIssues') || true;

        // Local path to main script
        const scriptUri = webview.asWebviewUri(
            vscode.Uri.file(path.join(this._extensionPath, 'media', 'settings.js'))
        );

        // Local path to css
        const styleUri = webview.asWebviewUri(
            vscode.Uri.file(path.join(this._extensionPath, 'media', 'settings.css'))
        );

        // Local path to the CodeHawk icon
        const iconUri = webview.asWebviewUri(
            vscode.Uri.file(path.join(this._extensionPath, 'media', 'codehawk-icon.svg'))
        );

        // Use a nonce to only allow specific scripts to be run
        const nonce = getNonce();

        return `<!DOCTYPE html>
    <html lang="en">
    <head>
      <meta charset="UTF-8">
      <meta name="viewport" content="width=device-width, initial-scale=1.0">
      <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource}; img-src ${webview.cspSource} https:; script-src 'nonce-${nonce}';">
      <link href="${styleUri}" rel="stylesheet">
      <title>CodeHawk Settings</title>
    </head>
    <body>
      <div class="container">
        <div class="header">
          <img src="${iconUri}" alt="CodeHawk Logo" class="logo">
          <h1>CodeHawk Settings</h1>
        </div>
        
        <div class="settings-form">
          <div class="form-group">
            <label for="apiUrl">API URL</label>
            <input type="text" id="apiUrl" value="${apiUrl}" placeholder="http://localhost:8080">
            <small>The URL of the CodeHawk API server</small>
          </div>
          
          <div class="form-group">
            <label for="apiKey">API Key</label>
            <input type="password" id="apiKey" value="${apiKey}" placeholder="Enter your API key">
            <small>Your CodeHawk API authentication key</small>
          </div>
          
          <div class="form-group checkbox">
            <input type="checkbox" id="autoAnalyze" ${autoAnalyze ? 'checked' : ''}>
            <label for="autoAnalyze">Automatically analyze on save</label>
            <small>Run analysis every time a file is saved</small>
          </div>
          
          <div class="form-group checkbox">
            <input type="checkbox" id="showInlineIssues" ${showInlineIssues ? 'checked' : ''}>
            <label for="showInlineIssues">Show inline issues</label>
            <small>Display issues directly in the editor</small>
          </div>
          
          <div class="button-group">
            <button id="saveButton">Save Settings</button>
            <button id="resetButton">Reset to Defaults</button>
          </div>
        </div>
        
        <div class="footer">
          <p>CodeHawk version 0.1.0</p>
        </div>
      </div>
      
      <script nonce="${nonce}">
        const vscode = acquireVsCodeApi();
        
        // Store current settings
        const settings = {
          apiUrl: "${apiUrl}",
          apiKey: "${apiKey}",
          autoAnalyze: ${autoAnalyze},
          showInlineIssues: ${showInlineIssues}
        };
        
        // Initialize state
        vscode.setState(settings);
        
        // Handle save button
        document.getElementById('saveButton').addEventListener('click', () => {
          const updatedSettings = {
            apiUrl: document.getElementById('apiUrl').value,
            apiKey: document.getElementById('apiKey').value,
            autoAnalyze: document.getElementById('autoAnalyze').checked,
            showInlineIssues: document.getElementById('showInlineIssues').checked
          };
          
          vscode.postMessage({
            command: 'saveSettings',
            settings: updatedSettings
          });
        });
        
        // Handle reset button
        document.getElementById('resetButton').addEventListener('click', () => {
          vscode.postMessage({
            command: 'resetSettings'
          });
        });
      </script>
    </body>
    </html>`;
    }
}

/**
 * Generate a nonce string
 */
function getNonce(): string {
    let text = '';
    const possible = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    for (let i = 0; i < 32; i++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length));
    }
    return text;
}