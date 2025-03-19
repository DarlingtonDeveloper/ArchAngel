import * as vscode from 'vscode';
import { analyzeCode } from '../utils/api';
import { ResultsProvider } from '../providers/resultsProvider';
import { SuggestionsProvider } from '../providers/suggestionsProvider';

/**
 * Map VS Code language IDs to CodeHawk language identifiers
 */
const languageMap: { [key: string]: string } = {
    'javascript': 'javascript',
    'typescript': 'typescript',
    'python': 'python',
    'go': 'go',
    'java': 'java',
    'csharp': 'csharp',
    'php': 'php',
    'ruby': 'ruby',
    // Default to the original if not mapped
};

/**
 * Get the CodeHawk language identifier from VS Code language ID
 */
function getLanguageIdentifier(vscodeLanguage: string): string {
    return languageMap[vscodeLanguage] || vscodeLanguage;
}

/**
 * Get context information about the current file
 */
function getFileContext(document: vscode.TextDocument): string {
    const fileName = document.fileName.split('/').pop() || '';
    const workspaceFolders = vscode.workspace.workspaceFolders;
    const workspaceName = workspaceFolders && workspaceFolders.length > 0
        ? workspaceFolders[0].name
        : '';

    return `File: ${fileName}, Workspace: ${workspaceName}`;
}

/**
 * Process analysis results and update VS Code diagnostics
 */
import { DiagnosticManager } from '../utils/diagnosticManager';

function processResults(
    document: vscode.TextDocument,
    results: any,
    context: vscode.ExtensionContext,
    resultsProvider: ResultsProvider,
    suggestionsProvider: SuggestionsProvider
): void {
    // Get the diagnostic manager
    const diagnosticManager = DiagnosticManager.getInstance();

    // Set diagnostics for the document
    if (results.issues && Array.isArray(results.issues)) {
        diagnosticManager.setDiagnostics(document, results.issues);
    } else {
        diagnosticManager.clearDiagnostics(document.uri);
    }

    // Update the results and suggestions providers
    resultsProvider.update(results);
    suggestionsProvider.update(results);

    // Store the analysis results in extension context
    context.workspaceState.update('codehawk.lastAnalysisResults', results);
    context.workspaceState.update('codehawk.lastAnalysisDocument', document.uri.toString());

    // Show a summary in the status bar
    const issueCount = results.issues ? results.issues.length : 0;
    vscode.window.setStatusBarMessage(`CodeHawk: Found ${issueCount} issues`, 5000);
}

/**
 * Map API severity to VS Code DiagnosticSeverity
 */
function getSeverity(apiSeverity: string): vscode.DiagnosticSeverity {
    switch (apiSeverity.toLowerCase()) {
        case 'error':
            return vscode.DiagnosticSeverity.Error;
        case 'warning':
            return vscode.DiagnosticSeverity.Warning;
        case 'suggestion':
        case 'info':
            return vscode.DiagnosticSeverity.Information;
        default:
            return vscode.DiagnosticSeverity.Hint;
    }
}

/**
 * Analyze the currently active file
 */
export async function analyzeCurrentFile(
    context: vscode.ExtensionContext,
    resultsProvider: ResultsProvider,
    suggestionsProvider: SuggestionsProvider
): Promise<void> {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No active editor found');
        return;
    }

    const document = editor.document;
    const language = getLanguageIdentifier(document.languageId);
    const fileContext = getFileContext(document);
    const code = document.getText();

    try {
        const results = await analyzeCode(code, language, fileContext);
        processResults(document, results, context, resultsProvider, suggestionsProvider);
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to analyze code: ${error}`);
        throw error;
    }
}

/**
 * Analyze the currently selected code
 */
export async function analyzeSelection(
    context: vscode.ExtensionContext,
    resultsProvider: ResultsProvider,
    suggestionsProvider: SuggestionsProvider
): Promise<void> {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No active editor found');
        return;
    }

    const selection = editor.selection;
    if (selection.isEmpty) {
        vscode.window.showWarningMessage('No code selected');
        return;
    }

    const document = editor.document;
    const language = getLanguageIdentifier(document.languageId);
    const fileContext = getFileContext(document);
    const code = document.getText(selection);

    try {
        const results = await analyzeCode(code, language, fileContext);

        // Adjust line numbers to be relative to the selection
        const selectionStartLine = selection.start.line;
        if (results.issues && Array.isArray(results.issues)) {
            results.issues.forEach((issue: any) => {
                issue.line = issue.line + selectionStartLine;
            });
        }

        processResults(document, results, context, resultsProvider, suggestionsProvider);
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to analyze code: ${error}`);
        throw error;
    }
}