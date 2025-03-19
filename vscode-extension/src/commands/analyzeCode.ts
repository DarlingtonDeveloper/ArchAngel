import * as vscode from 'vscode';
import { getApiClient } from '../api/client';
import { AnalysisResponse } from '../api/types';
import { ResultsProvider } from '../providers/resultsProvider';
import { SuggestionsProvider } from '../providers/suggestionsProvider';
import { DiagnosticManager } from '../utils/diagnosticManager';

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
export function processResults(
    document: vscode.TextDocument,
    results: AnalysisResponse,
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
    const suggestionCount = results.suggestions ? results.suggestions.length : 0;

    let message = `CodeHawk: Found ${issueCount} issues`;
    if (suggestionCount > 0) {
        message += ` and ${suggestionCount} suggestions`;
    }

    if (results.aiEnhanced) {
        message += " (AI-enhanced)";
    }

    vscode.window.setStatusBarMessage(message, 5000);
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
        // Get API client
        const client = getApiClient();

        // Show progress notification
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "CodeHawk Analysis",
            cancellable: false
        }, async (progress) => {
            progress.report({ message: "Analyzing code..." });

            // Make request
            const results = await client.analyzeCode({
                code,
                language,
                context: fileContext,
                options: {
                    // Get additional options from settings
                    ai_suggestions: vscode.workspace.getConfiguration('codehawk').get<boolean>('aiSuggestions', true)
                }
            });

            // Process results
            processResults(document, results, context, resultsProvider, suggestionsProvider);
        });
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to analyze code: ${(error as Error).message}`);
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
    const fileContext = getFileContext(document) + ' (selection)';
    const code = document.getText(selection);

    try {
        // Get API client
        const client = getApiClient();

        // Show progress notification
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "CodeHawk Analysis",
            cancellable: false
        }, async (progress) => {
            progress.report({ message: "Analyzing selected code..." });

            // Make request
            const results = await client.analyzeCode({
                code,
                language,
                context: fileContext,
                options: {
                    // Get additional options from settings
                    ai_suggestions: vscode.workspace.getConfiguration('codehawk').get<boolean>('aiSuggestions', true)
                }
            });

            // Adjust line numbers to be relative to the selection
            const selectionStartLine = selection.start.line;
            if (results.issues && Array.isArray(results.issues)) {
                results.issues.forEach((issue) => {
                    issue.line = issue.line + selectionStartLine;
                });
            }

            if (results.suggestions && Array.isArray(results.suggestions)) {
                results.suggestions.forEach((suggestion) => {
                    suggestion.line = suggestion.line + selectionStartLine;
                });
            }

            // Process results
            processResults(document, results, context, resultsProvider, suggestionsProvider);
        });
    } catch (error) {
        vscode.window.showErrorMessage(`Failed to analyze code: ${(error as Error).message}`);
        throw error;
    }
}