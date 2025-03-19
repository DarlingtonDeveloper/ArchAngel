import * as vscode from 'vscode';

/**
 * Tree item representing a code suggestion
 */
export class SuggestionItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly suggestion: any,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(label, collapsibleState);

        this.tooltip = suggestion.description || suggestion.message;
        this.description = `Line ${suggestion.line}`;

        // Set the command to apply the suggestion when clicked
        this.command = {
            command: 'codehawk.applySuggestion',
            title: 'Apply Suggestion',
            arguments: [this.suggestion]
        };

        this.iconPath = new vscode.ThemeIcon('lightbulb');
        this.contextValue = 'suggestion';
    }
}

/**
 * Tree data provider for AI-powered code suggestions
 */
export class SuggestionsProvider implements vscode.TreeDataProvider<vscode.TreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<vscode.TreeItem | undefined | void> = new vscode.EventEmitter<vscode.TreeItem | undefined | void>();
    readonly onDidChangeTreeData: vscode.Event<vscode.TreeItem | undefined | void> = this._onDidChangeTreeData.event;

    private results: any = null;

    constructor(private context: vscode.ExtensionContext) {
        // Register apply suggestion command
        vscode.commands.registerCommand('codehawk.applySuggestion', (suggestion) => {
            this.applySuggestion(suggestion);
        });
    }

    /**
     * Update the results and refresh the view
     */
    public update(results: any): void {
        this.results = results;
        this._onDidChangeTreeData.fire();
    }

    /**
     * Apply a suggestion to the code
     */
    private async applySuggestion(suggestion: any): Promise<void> {
        const lastDocumentUri = this.context.workspaceState.get<string>('codehawk.lastAnalysisDocument');
        if (!lastDocumentUri) {
            vscode.window.showWarningMessage('Cannot apply suggestion: No document information available');
            return;
        }

        const uri = vscode.Uri.parse(lastDocumentUri);

        try {
            const document = await vscode.workspace.openTextDocument(uri);
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
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to apply suggestion: ${error}`);
        }
    }

    /**
     * Get tree item for a given element
     */
    getTreeItem(element: vscode.TreeItem): vscode.TreeItem {
        return element;
    }

    /**
     * Get children for a given element
     */
    getChildren(element?: vscode.TreeItem): Thenable<vscode.TreeItem[]> {
        if (!this.results || !this.results.suggestions || !Array.isArray(this.results.suggestions) || this.results.suggestions.length === 0) {
            // If no explicit suggestions, try to extract suggestions from issues
            if (this.results && this.results.issues && Array.isArray(this.results.issues)) {
                // Filter issues that have fix suggestions
                const suggestiveIssues = this.results.issues.filter((issue: any) =>
                    issue.severity.toLowerCase() === 'suggestion' || issue.fix);

                if (suggestiveIssues.length > 0) {
                    return Promise.resolve(
                        suggestiveIssues.map((issue: any) => {
                            // Create a suggestion object from the issue
                            const suggestion = {
                                ...issue,
                                description: issue.message,
                                fix: issue.fix || this.generateFix(issue)
                            };

                            return new SuggestionItem(issue.message, suggestion,
                                issue.fix ? vscode.TreeItemCollapsibleState.Collapsed : vscode.TreeItemCollapsibleState.None);
                        })
                    );
                }
            }

            return Promise.resolve([
                new vscode.TreeItem('No suggestions available', vscode.TreeItemCollapsibleState.None)
            ]);
        }

        if (!element) {
            // Root level - show all suggestions
            return Promise.resolve(
                this.results.suggestions.map((suggestion: any) =>
                    new SuggestionItem(
                        suggestion.message || suggestion.description,
                        suggestion,
                        suggestion.fix ? vscode.TreeItemCollapsibleState.Collapsed : vscode.TreeItemCollapsibleState.None
                    )
                )
            );
        } else if (element instanceof SuggestionItem && element.suggestion.fix) {
            // Show fix details
            const fix = element.suggestion.fix;

            return Promise.resolve([
                new vscode.TreeItem(`Replace with: "${fix.replacement}"`, vscode.TreeItemCollapsibleState.None)
            ]);
        }

        return Promise.resolve([]);
    }

    /**
     * Generate a simple fix suggestion for common issues
     * This would be expanded in a real implementation
     */
    private generateFix(issue: any): any | undefined {
        // These are just examples; in a real implementation, 
        // these would be more sophisticated and context-aware
        if (issue.ruleId === 'PY001' && issue.message.includes('docstring')) {
            return {
                replacement: '"""Add docstring here."""\n' // Simple docstring insertion
            };
        }

        if (issue.ruleId === 'PY002' && issue.message.includes('print')) {
            return {
                replacement: 'import logging\nlogging.info("Log message")' // Replace print with logging
            };
        }

        if (issue.ruleId === 'JS001' && issue.message.includes('semicolon')) {
            return {
                replacement: 'const x = 1;' // Add semicolon
            };
        }

        // For most issues, we can't generate automatic fixes without more context
        return undefined;
    }
}