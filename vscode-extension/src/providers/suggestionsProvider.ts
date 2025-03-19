import * as vscode from 'vscode';
import { Suggestion, Issue, AnalysisResponse } from '../api/types';

/**
 * Tree item representing a code suggestion
 */
export class SuggestionItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly suggestion: Suggestion | Issue,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState,
        public readonly isAiSuggestion: boolean = false
    ) {
        super(label, collapsibleState);

        this.tooltip = suggestion.message;
        this.description = `Line ${suggestion.line}${suggestion.column ? `, Col ${suggestion.column}` : ''}`;

        // Set the command to apply the suggestion when clicked
        this.command = {
            command: 'codehawk.applySuggestion',
            title: 'Apply Suggestion',
            arguments: [this.suggestion]
        };

        // Set icon based on whether it's an AI suggestion or not
        if (isAiSuggestion) {
            this.iconPath = new vscode.ThemeIcon('sparkle');
            this.contextValue = 'ai-suggestion';
        } else {
            this.iconPath = new vscode.ThemeIcon('lightbulb');
            this.contextValue = 'suggestion';
        }
    }
}

/**
 * Tree item representing a category of suggestions
 */
export class SuggestionCategoryItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly suggestions: (Suggestion | Issue)[],
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(label, collapsibleState);
        this.description = `(${suggestions.length})`;
        this.contextValue = 'suggestion-category';
    }
}

/**
 * Tree data provider for AI-powered code suggestions
 */
export class SuggestionsProvider implements vscode.TreeDataProvider<vscode.TreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<vscode.TreeItem | undefined | void> = new vscode.EventEmitter<vscode.TreeItem | undefined | void>();
    readonly onDidChangeTreeData: vscode.Event<vscode.TreeItem | undefined | void> = this._onDidChangeTreeData.event;

    private results: AnalysisResponse | null = null;
    private categories: { [key: string]: (Suggestion | Issue)[] } = {};

    constructor(private context: vscode.ExtensionContext) { }

    /**
     * Update the results and refresh the view
     */
    public update(results: AnalysisResponse | null): void {
        this.results = results;
        this.categorize();
        this._onDidChangeTreeData.fire();
    }

    /**
     * Categorize suggestions
     */
    private categorize(): void {
        this.categories = {};

        if (!this.results) {
            return;
        }

        // Collect all suggestions
        const allSuggestions: (Suggestion | Issue)[] = [];

        // Add explicit suggestions
        if (this.results.suggestions && Array.isArray(this.results.suggestions)) {
            allSuggestions.push(...this.results.suggestions);
        }

        // Add suggestions from issues
        if (this.results.issues && Array.isArray(this.results.issues)) {
            // Filter issues that have fix suggestions or are marked as suggestions
            const suggestiveIssues = this.results.issues.filter(issue =>
                issue.severity.toLowerCase() === 'suggestion' || issue.fix);

            allSuggestions.push(...suggestiveIssues);
        }

        // Categorize by type
        for (const suggestion of allSuggestions) {
            let category = 'General';

            // Determine category based on message or rule ID
            if (suggestion.ruleId) {
                if (suggestion.ruleId.includes('format') || suggestion.ruleId.includes('style')) {
                    category = 'Formatting';
                } else if (suggestion.ruleId.includes('doc') || suggestion.message.toLowerCase().includes('document')) {
                    category = 'Documentation';
                } else if (suggestion.ruleId.includes('perf') || suggestion.message.toLowerCase().includes('performance')) {
                    category = 'Performance';
                } else if (suggestion.ruleId.includes('unused') || suggestion.message.toLowerCase().includes('unused')) {
                    category = 'Unused Code';
                } else if (suggestion.ruleId.includes('name') || suggestion.message.toLowerCase().includes('name')) {
                    category = 'Naming';
                }
            } else {
                // Categorize based on message
                const message = suggestion.message.toLowerCase();
                if (message.includes('format') || message.includes('indent') || message.includes('whitespace')) {
                    category = 'Formatting';
                } else if (message.includes('document') || message.includes('comment')) {
                    category = 'Documentation';
                } else if (message.includes('performance') || message.includes('efficient')) {
                    category = 'Performance';
                } else if (message.includes('unused') || message.includes('dead code')) {
                    category = 'Unused Code';
                } else if (message.includes('name') || message.includes('variable')) {
                    category = 'Naming';
                }
            }

            // Add to category
            if (!this.categories[category]) {
                this.categories[category] = [];
            }

            this.categories[category].push(suggestion);
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
        if (!this.results || Object.keys(this.categories).length === 0) {
            return Promise.resolve([
                new vscode.TreeItem('No suggestions available', vscode.TreeItemCollapsibleState.None)
            ]);
        }

        if (!element) {
            // Root level - show categories
            return Promise.resolve(
                Object.entries(this.categories).map(([category, suggestions]) =>
                    new SuggestionCategoryItem(
                        category,
                        suggestions,
                        vscode.TreeItemCollapsibleState.Expanded
                    )
                )
            );
        } else if (element instanceof SuggestionCategoryItem) {
            // Category level - show suggestions in this category
            return Promise.resolve(
                element.suggestions.map(suggestion => {
                    // Check if it's an AI suggestion
                    const isAiSuggestion = !suggestion.ruleId || suggestion.ruleId === 'ai-suggestion';

                    return new SuggestionItem(
                        suggestion.message,
                        suggestion,
                        suggestion.fix ? vscode.TreeItemCollapsibleState.Collapsed : vscode.TreeItemCollapsibleState.None,
                        isAiSuggestion
                    );
                })
            );
        } else if (element instanceof SuggestionItem && element.suggestion.fix) {
            // Show fix details
            return Promise.resolve([
                new vscode.TreeItem(
                    `${element.suggestion.fix.description || 'Replace with'}:`,
                    vscode.TreeItemCollapsibleState.None
                ),
                new vscode.TreeItem(
                    `"${element.suggestion.fix.replacement}"`,
                    vscode.TreeItemCollapsibleState.None
                )
            ]);
        }

        return Promise.resolve([]);
    }

    /**
     * Apply a suggestion to the code
     */
    public async applySuggestion(suggestion: Suggestion | Issue): Promise<void> {
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
                // Determine range to replace
                let range: vscode.Range;

                if (replacement.includes('\n')) {
                    // Multi-line replacement - determine the full range
                    // For simplicity, just replace the current line for now
                    range = new vscode.Range(
                        lineIndex, 0,
                        lineIndex, currentLine.length
                    );
                } else {
                    // Single line replacement
                    if (suggestion.column) {
                        // If we have column information, use it
                        const columnIndex = Math.max(0, suggestion.column - 1);

                        // Try to determine the end of the item to replace
                        // This is a simplistic approach - in a real implementation,
                        // you would parse the code to find the exact range
                        let endColumn = columnIndex;
                        for (let i = columnIndex; i < currentLine.length; i++) {
                            const ch = currentLine[i];
                            if (/[\s;,{}()]/.test(ch)) {
                                break;
                            }
                            endColumn = i + 1;
                        }

                        range = new vscode.Range(
                            lineIndex, columnIndex,
                            lineIndex, endColumn
                        );
                    } else {
                        // If no column information, replace the entire line
                        range = new vscode.Range(
                            lineIndex, 0,
                            lineIndex, currentLine.length
                        );
                    }
                }

                editBuilder.replace(range, replacement);
            });

            vscode.window.showInformationMessage('Suggestion applied successfully');
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to apply suggestion: ${(error as Error).message}`);
        }
    }
}