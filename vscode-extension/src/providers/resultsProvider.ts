import * as vscode from 'vscode';
import * as path from 'path';

/**
 * Tree item representing an analysis result issue
 */
export class IssueItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly issue: any,
        public readonly collapsibleState: vscode.TreeItemCollapsibleState = vscode.TreeItemCollapsibleState.None
    ) {
        super(label, collapsibleState);

        this.tooltip = issue.message;
        this.description = `Line ${issue.line}${issue.column ? `, Col ${issue.column}` : ''}`;

        // Set the command to navigate to the issue location when clicked
        this.command = {
            command: 'codehawk.navigateToIssue',
            title: 'Navigate to Issue',
            arguments: [this.issue]
        };

        // Set the appropriate icon based on severity
        switch (issue.severity.toLowerCase()) {
            case 'error':
                this.iconPath = new vscode.ThemeIcon('error');
                break;
            case 'warning':
                this.iconPath = new vscode.ThemeIcon('warning');
                break;
            case 'suggestion':
            case 'info':
                this.iconPath = new vscode.ThemeIcon('info');
                break;
            default:
                this.iconPath = new vscode.ThemeIcon('lightbulb');
        }
    }
}

/**
 * Tree item representing a category of issues
 */
export class CategoryItem extends vscode.TreeItem {
    constructor(
        public readonly label: string,
        public readonly issues: any[],
        public readonly collapsibleState: vscode.TreeItemCollapsibleState
    ) {
        super(label, collapsibleState);
        this.description = `(${issues.length})`;
        this.tooltip = `${issues.length} ${label.toLowerCase()} issues found`;
    }
}

/**
 * Tree data provider for analysis results
 */
export class ResultsProvider implements vscode.TreeDataProvider<vscode.TreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<vscode.TreeItem | undefined | void> = new vscode.EventEmitter<vscode.TreeItem | undefined | void>();
    readonly onDidChangeTreeData: vscode.Event<vscode.TreeItem | undefined | void> = this._onDidChangeTreeData.event;

    private results: any = null;

    constructor(private context: vscode.ExtensionContext) {
        // Register navigation command
        vscode.commands.registerCommand('codehawk.navigateToIssue', (issue) => {
            this.navigateToIssue(issue);
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
     * Navigate to the location of an issue in the editor
     */
    private navigateToIssue(issue: any): void {
        const lastDocumentUri = this.context.workspaceState.get<string>('codehawk.lastAnalysisDocument');
        if (!lastDocumentUri) {
            vscode.window.showWarningMessage('Cannot navigate to issue: No document information available');
            return;
        }

        const uri = vscode.Uri.parse(lastDocumentUri);

        vscode.workspace.openTextDocument(uri).then(document => {
            vscode.window.showTextDocument(document).then(editor => {
                // Convert to 0-based if the API is 1-based
                const lineIndex = Math.max(0, issue.line - 1);
                const columnIndex = issue.column ? Math.max(0, issue.column - 1) : 0;

                // Create position and selection
                const position = new vscode.Position(lineIndex, columnIndex);
                const lineText = document.lineAt(lineIndex).text;
                const endPosition = new vscode.Position(lineIndex, lineText.length);

                // Set selection and reveal it
                editor.selection = new vscode.Selection(position, position);
                editor.revealRange(
                    new vscode.Range(position, endPosition),
                    vscode.TextEditorRevealType.InCenter
                );
            });
        });
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
        if (!this.results || !this.results.issues || !Array.isArray(this.results.issues) || this.results.issues.length === 0) {
            return Promise.resolve([
                new vscode.TreeItem('No issues found', vscode.TreeItemCollapsibleState.None)
            ]);
        }

        if (!element) {
            // Root level - group by severity
            const errors = this.results.issues.filter((issue: any) =>
                issue.severity.toLowerCase() === 'error');

            const warnings = this.results.issues.filter((issue: any) =>
                issue.severity.toLowerCase() === 'warning');

            const suggestions = this.results.issues.filter((issue: any) =>
                ['suggestion', 'info'].includes(issue.severity.toLowerCase()));

            const categories: vscode.TreeItem[] = [];

            if (errors.length > 0) {
                categories.push(new CategoryItem('Errors', errors,
                    vscode.TreeItemCollapsibleState.Expanded));
            }

            if (warnings.length > 0) {
                categories.push(new CategoryItem('Warnings', warnings,
                    vscode.TreeItemCollapsibleState.Expanded));
            }

            if (suggestions.length > 0) {
                categories.push(new CategoryItem('Suggestions', suggestions,
                    vscode.TreeItemCollapsibleState.Expanded));
            }

            return Promise.resolve(categories);
        } else if (element instanceof CategoryItem) {
            // Category level - show all issues in this category
            return Promise.resolve(
                element.issues.map(issue =>
                    new IssueItem(issue.message, issue)
                )
            );
        }

        return Promise.resolve([]);
    }
}