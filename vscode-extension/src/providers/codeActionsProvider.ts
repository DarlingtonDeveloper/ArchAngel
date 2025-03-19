import * as vscode from 'vscode';

/**
 * Provider for code actions that offer quick fixes for CodeHawk issues
 */
export class CodeActionsProvider implements vscode.CodeActionProvider {
    /**
     * Provide code actions for the given document and range
     */
    public provideCodeActions(
        document: vscode.TextDocument,
        range: vscode.Range | vscode.Selection,
        context: vscode.CodeActionContext,
        token: vscode.CancellationToken
    ): vscode.ProviderResult<(vscode.Command | vscode.CodeAction)[]> {
        // Filter diagnostics that belong to CodeHawk
        const relevantDiagnostics = context.diagnostics.filter(
            diagnostic => diagnostic.source === 'CodeHawk'
        );

        if (relevantDiagnostics.length === 0) {
            return [];
        }

        const actions: vscode.CodeAction[] = [];

        // Generate code actions for each diagnostic
        for (const diagnostic of relevantDiagnostics) {
            // Try to generate a fix based on the diagnostic
            const fix = this.generateFix(document, diagnostic);
            if (fix) {
                actions.push(fix);
            }

            // Add an action to ignore this rule in the future
            const ignoreAction = this.createIgnoreRuleAction(document, diagnostic);
            if (ignoreAction) {
                actions.push(ignoreAction);
            }

            // Add an action to find out more about this rule
            const infoAction = this.createRuleInfoAction(diagnostic);
            if (infoAction) {
                actions.push(infoAction);
            }
        }

        return actions;
    }

    /**
     * Generate a fix for a diagnostic based on its rule ID
     */
    private generateFix(document: vscode.TextDocument, diagnostic: vscode.Diagnostic): vscode.CodeAction | undefined {
        const ruleId = diagnostic.code?.toString() || '';
        const lineIndex = diagnostic.range.start.line;
        const currentLine = document.lineAt(lineIndex).text;

        let replacement: string | undefined;

        // Generate replacements based on rule ID
        // These are simplified examples; real implementations would be more sophisticated
        switch (ruleId) {
            case 'PY001': // Missing docstring
                replacement = '"""Add docstring here."""\n' + currentLine;
                break;

            case 'PY002': // Use logging instead of print
                if (currentLine.includes('print(')) {
                    replacement = currentLine.replace(
                        /print\((.*?)\)/g,
                        'logging.info($1)'
                    );

                    // If we don't see an import for logging, add it
                    const text = document.getText();
                    if (!text.includes('import logging')) {
                        replacement = 'import logging\n\n' + replacement;
                    }
                }
                break;

            case 'JS001': // Missing semicolon
                if (!currentLine.endsWith(';')) {
                    replacement = currentLine + ';';
                }
                break;

            case 'JS002': // Prefer const over let
                if (currentLine.startsWith('let ') && !currentLine.includes('=')) {
                    replacement = currentLine.replace(/^let /, 'const ');
                }
                break;

            default:
                return undefined;
        }

        if (!replacement) {
            return undefined;
        }

        // Create the code action
        const action = new vscode.CodeAction(
            `Fix: ${diagnostic.message}`,
            vscode.CodeActionKind.QuickFix
        );

        action.diagnostics = [diagnostic];
        action.edit = new vscode.WorkspaceEdit();
        action.edit.replace(
            document.uri,
            new vscode.Range(lineIndex, 0, lineIndex, currentLine.length),
            replacement
        );

        return action;
    }

    /**
     * Create an action to ignore this rule in the future
     */
    private createIgnoreRuleAction(document: vscode.TextDocument, diagnostic: vscode.Diagnostic): vscode.CodeAction | undefined {
        const ruleId = diagnostic.code?.toString() || '';
        if (!ruleId) {
            return undefined;
        }

        const action = new vscode.CodeAction(
            `Ignore rule: ${ruleId}`,
            vscode.CodeActionKind.QuickFix
        );

        action.command = {
            command: 'codehawk.ignoreRule',
            title: `Ignore rule: ${ruleId}`,
            arguments: [ruleId, document.uri]
        };

        return action;
    }

    /**
     * Create an action to show more information about this rule
     */
    private createRuleInfoAction(diagnostic: vscode.Diagnostic): vscode.CodeAction | undefined {
        const ruleId = diagnostic.code?.toString() || '';
        if (!ruleId) {
            return undefined;
        }

        const action = new vscode.CodeAction(
            `Learn more about ${ruleId}`,
            vscode.CodeActionKind.QuickFix
        );

        action.command = {
            command: 'codehawk.showRuleInfo',
            title: `Learn more about ${ruleId}`,
            arguments: [ruleId]
        };

        return action;
    }
}