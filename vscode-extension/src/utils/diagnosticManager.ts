import * as vscode from 'vscode';

/**
 * Manages diagnostics for CodeHawk issues
 */
export class DiagnosticManager {
    private static instance: DiagnosticManager;
    private diagnosticCollection: vscode.DiagnosticCollection;

    private constructor() {
        this.diagnosticCollection = vscode.languages.createDiagnosticCollection('codehawk');
    }

    /**
     * Get the DiagnosticManager instance
     */
    public static getInstance(): DiagnosticManager {
        if (!DiagnosticManager.instance) {
            DiagnosticManager.instance = new DiagnosticManager();
        }
        return DiagnosticManager.instance;
    }

    /**
     * Set diagnostics for a document
     */
    public setDiagnostics(document: vscode.TextDocument, issues: any[]): void {
        if (!issues || !Array.isArray(issues) || issues.length === 0) {
            // Clear diagnostics if no issues
            this.diagnosticCollection.set(document.uri, []);
            return;
        }

        const diagnostics: vscode.Diagnostic[] = issues.map(issue => {
            try {
                // Convert line numbers (API might use 1-based, VS Code uses 0-based)
                const lineIndex = Math.max(0, (issue.line || 1) - 1);
                const columnIndex = issue.column ? Math.max(0, issue.column - 1) : 0;

                // Get the line of code
                let line;
                try {
                    line = document.lineAt(lineIndex);
                } catch (error) {
                    // If the line doesn't exist in the document, use the first line
                    line = document.lineAt(0);
                }

                // Create a range for the diagnostic
                // If column is specified, start from there, otherwise highlight the whole line
                const range = new vscode.Range(
                    lineIndex, columnIndex,
                    lineIndex, line.text.length
                );

                // Map severity from API to VS Code DiagnosticSeverity
                const severity = this.mapSeverity(issue.severity);

                // Create the diagnostic
                const diagnostic = new vscode.Diagnostic(
                    range,
                    issue.message,
                    severity
                );

                // Add source and code information
                diagnostic.source = 'CodeHawk';
                diagnostic.code = issue.ruleId || '';

                // Add additional metadata to help with quick fixes
                diagnostic.relatedInformation = [];

                if (issue.fix) {
                    diagnostic.relatedInformation.push(
                        new vscode.DiagnosticRelatedInformation(
                            new vscode.Location(document.uri, range),
                            `Suggested fix: ${issue.fix.description || 'Available'}`
                        )
                    );
                }

                return diagnostic;
            } catch (error) {
                console.error('Error creating diagnostic:', error);
                return null;
            }
        }).filter(Boolean) as vscode.Diagnostic[];

        // Set the diagnostics for the document
        this.diagnosticCollection.set(document.uri, diagnostics);
    }

    /**
     * Clear diagnostics for a document
     */
    public clearDiagnostics(uri: vscode.Uri): void {
        this.diagnosticCollection.delete(uri);
    }

    /**
     * Clear all diagnostics
     */
    public clearAllDiagnostics(): void {
        this.diagnosticCollection.clear();
    }

    /**
     * Dispose of the diagnostic collection
     */
    public dispose(): void {
        this.diagnosticCollection.dispose();
    }

    /**
     * Map API severity to VS Code DiagnosticSeverity
     */
    private mapSeverity(severity: string): vscode.DiagnosticSeverity {
        const severityMap: Record<string, vscode.DiagnosticSeverity> = {
            'error': vscode.DiagnosticSeverity.Error,
            'warning': vscode.DiagnosticSeverity.Warning,
            'suggestion': vscode.DiagnosticSeverity.Information,
            'info': vscode.DiagnosticSeverity.Information,
            'hint': vscode.DiagnosticSeverity.Hint
        };

        return severityMap[severity?.toLowerCase()] || vscode.DiagnosticSeverity.Information;
    }
}