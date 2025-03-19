import * as vscode from 'vscode';
import * as path from 'path';
import { Suggestion, Issue } from '../api/types';

/**
 * Detailed view for a single suggestion
 */
export class SuggestionDetailView {
    private panel: vscode.WebviewPanel | undefined;

    constructor(private context: vscode.ExtensionContext) { }

    /**
     * Show the detail view for a suggestion
     */
    public show(suggestion: Suggestion | Issue, document: vscode.TextDocument): void {
        // Create panel if it doesn't exist
        if (!this.panel) {
            this.panel = vscode.window.createWebviewPanel(
                'codehawkSuggestionDetail',
                'CodeHawk Suggestion',
                vscode.ViewColumn.Beside,
                {
                    enableScripts: true,
                    localResourceRoots: [vscode.Uri.file(path.join(this.context.extensionPath, 'media'))]
                }
            );

            // Handle panel disposal
            this.panel.onDidDispose(() => {
                this.panel = undefined;
            });
        }

        // Get the relevant code
        const lineIndex = Math.max(0, suggestion.line - 1);

        // Get surrounding context (3 lines before and after)
        const startLine = Math.max(0, lineIndex - 3);
        const endLine = Math.min(document.lineCount - 1, lineIndex + 3);
        const contextLines: string[] = [];

        for (let i = startLine; i <= endLine; i++) {
            const contextLine = document.lineAt(i).text;
            if (i === lineIndex) {
                contextLines.push(`<div class="highlight-line">${this.escapeHtml(contextLine)}</div>`);
            } else {
                contextLines.push(`<div class="code-line">${this.escapeHtml(contextLine)}<span class="line-number">${i + 1}</span></div>`);
            }
        }

        // Set panel title
        this.panel.title = `CodeHawk: ${suggestion.severity.charAt(0).toUpperCase() + suggestion.severity.slice(1)}`;

        // Set panel HTML
        this.panel.webview.html = this.getHtmlForDetail(suggestion, contextLines, document.lineAt(lineIndex).text);

        // Set up message handling
        this.panel.webview.onDidReceiveMessage(message => {
            switch (message.command) {
                case 'applySuggestion':
                    this.applySuggestion(suggestion, document);
                    break;
            }
        });

        // Show the panel
        this.panel.reveal(vscode.ViewColumn.Beside);
    }

    /**
     * Apply the suggestion to the document
     */
    private async applySuggestion(suggestion: Suggestion | Issue, document: vscode.TextDocument): Promise<void> {
        try {
            const editor = await vscode.window.showTextDocument(document);

            // Convert to 0-based if the API is 1-based
            const lineIndex = Math.max(0, suggestion.line - 1);

            // Get the current line and the suggested replacement
            const currentLine = document.lineAt(lineIndex).text;
            const replacement = suggestion.fix?.replacement || '';

            if (!replacement) {
                vscode.window.showInformationMessage('This suggestion does not include an automatic fix');
                return;
            }

            // Create edit
            await editor.edit(editBuilder => {
                const range = new vscode.Range(
                    lineIndex, 0,
                    lineIndex, currentLine.length
                );

                editBuilder.replace(range, replacement);
            });

            vscode.window.showInformationMessage('Suggestion applied successfully');

            // Close the panel
            if (this.panel) {
                this.panel.dispose();
            }
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to apply suggestion: ${(error as Error).message}`);
        }
    }

    /**
     * Get HTML for the detail view
     */
    private getHtmlForDetail(suggestion: Suggestion | Issue, contextLines: string[], currentLine: string): string {
        const styleUri = this.panel!.webview.asWebviewUri(
            vscode.Uri.file(path.join(this.context.extensionPath, 'media', 'suggestion-detail.css'))
        );

        const scriptUri = this.panel!.webview.asWebviewUri(
            vscode.Uri.file(path.join(this.context.extensionPath, 'media', 'suggestion-detail.js'))
        );

        // Check if this is an AI suggestion
        const isAI = !suggestion.ruleId || suggestion.ruleId === 'ai-suggestion';
        const confidence = 'confidence' in suggestion ? suggestion.confidence : undefined;

        return `<!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <link href="${styleUri}" rel="stylesheet">
            <title>CodeHawk Suggestion</title>
        </head>
        <body>
            <div class="container">
                <div class="header ${suggestion.severity}">
                    <h1>
                        ${isAI ? '<span class="ai-badge">AI</span>' : ''}
                        ${suggestion.severity.toUpperCase()}: ${this.escapeHtml(suggestion.message)}
                    </h1>
                    ${confidence !== undefined ? `<div class="confidence">Confidence: ${Math.round(confidence * 100)}%</div>` : ''}
                </div>
                
                <div class="location-info">
                    <div class="info-item">
                        <span class="info-label">Location:</span>
                        <span class="info-value">Line ${suggestion.line}${suggestion.column ? `, Column ${suggestion.column}` : ''}</span>
                    </div>
                    ${suggestion.ruleId ? `
                    <div class="info-item">
                        <span class="info-label">Rule:</span>
                        <span class="info-value">${suggestion.ruleId}</span>
                    </div>
                    ` : ''}
                </div>
                
                <div class="section">
                    <h2>Code Context</h2>
                    <div class="code-context">
                        ${contextLines.join('\n')}
                    </div>
                </div>
                
                ${suggestion.fix ? `
                <div class="section">
                    <h2>Suggested Fix</h2>
                    <div class="fix-description">${this.escapeHtml(suggestion.fix.description || 'Replace with:')}</div>
                    <pre class="code-fix"><code>${this.escapeHtml(suggestion.fix.replacement)}</code></pre>
                    <button id="applyFixButton" class="primary-button">Apply Fix</button>
                </div>
                ` : ''}
                
                ${suggestion.context ? `
                <div class="section">
                    <h2>Additional Context</h2>
                    <div class="additional-context">${this.escapeHtml(suggestion.context)}</div>
                </div>
                ` : ''}
                
                <div class="footer">
                    <div class="info-item">
                        <span class="info-label">Severity:</span>
                        <span class="severity-badge ${suggestion.severity}">${suggestion.severity}</span>
                    </div>
                </div>
            </div>
            
            <script src="${scriptUri}"></script>
        </body>
        </html>`;
    }

    /**
     * Escape HTML special characters
     */
    private escapeHtml(text: string): string {
        return text
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }
}