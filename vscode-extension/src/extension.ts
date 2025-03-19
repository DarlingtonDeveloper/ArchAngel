import * as vscode from 'vscode';
import { analyzeCurrentFile, analyzeSelection } from './commands/analyzeCode';
import { ResultsProvider } from './providers/resultsProvider';
import { SuggestionsProvider } from './providers/suggestionsProvider';
import { CodeActionsProvider } from './providers/codeActionsProvider';
import { checkApiConfiguration } from './utils/api';

let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
    console.log('CodeHawk extension is now active');

    // Create status bar item
    statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    statusBarItem.text = 'CodeHawk';
    statusBarItem.tooltip = 'Run CodeHawk analysis on current file';
    statusBarItem.command = 'codehawk.analyzeCurrentFile';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);

    // Register views
    const resultsProvider = new ResultsProvider(context);
    const suggestionsProvider = new SuggestionsProvider(context);

    // Register tree views
    vscode.window.registerTreeDataProvider('codehawkResults', resultsProvider);
    vscode.window.registerTreeDataProvider('codehawkSuggestions', suggestionsProvider);

    // Register code actions provider
    const supportedLanguages = [
        'javascript', 'typescript', 'python', 'go',
        'java', 'csharp', 'php', 'ruby'
    ];
    const codeActionsProvider = new CodeActionsProvider();
    supportedLanguages.forEach(language => {
        context.subscriptions.push(
            vscode.languages.registerCodeActionsProvider(
                { language },
                codeActionsProvider,
                { providedCodeActionKinds: [vscode.CodeActionKind.QuickFix] }
            )
        );
    });

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('codehawk.analyzeCurrentFile', async () => {
            if (!(await checkApiConfiguration())) {
                return;
            }
            statusBarItem.text = '$(sync~spin) CodeHawk: Analyzing...';
            try {
                await analyzeCurrentFile(context, resultsProvider, suggestionsProvider);
                statusBarItem.text = 'CodeHawk';
            } catch (error) {
                statusBarItem.text = '$(error) CodeHawk';
                vscode.window.showErrorMessage(`CodeHawk analysis failed: ${error}`);
            }
        }),

        vscode.commands.registerCommand('codehawk.analyzeSelection', async () => {
            if (!(await checkApiConfiguration())) {
                return;
            }
            statusBarItem.text = '$(sync~spin) CodeHawk: Analyzing...';
            try {
                await analyzeSelection(context, resultsProvider, suggestionsProvider);
                statusBarItem.text = 'CodeHawk';
            } catch (error) {
                statusBarItem.text = '$(error) CodeHawk';
                vscode.window.showErrorMessage(`CodeHawk analysis failed: ${error}`);
            }
        }),

        vscode.commands.registerCommand('codehawk.showResults', () => {
            vscode.commands.executeCommand('codehawk-explorer.focus');
        }),

        vscode.commands.registerCommand('codehawk.configure', () => {
            vscode.commands.executeCommand('workbench.action.openSettings', 'codehawk');
        })
    );

    // Set up auto-analyze on save if enabled
    const config = vscode.workspace.getConfiguration('codehawk');
    if (config.get('autoAnalyze')) {
        context.subscriptions.push(
            vscode.workspace.onDidSaveTextDocument(async (document) => {
                if (supportedLanguages.includes(document.languageId)) {
                    await vscode.commands.executeCommand('codehawk.analyzeCurrentFile');
                }
            })
        );
    }

    // Welcome message on first activation
    const hasShownWelcome = context.globalState.get('codehawk.hasShownWelcome');
    if (!hasShownWelcome) {
        vscode.window.showInformationMessage(
            'CodeHawk is now active! Configure your API settings to get started.',
            'Configure Now'
        ).then(selection => {
            if (selection === 'Configure Now') {
                vscode.commands.executeCommand('codehawk.configure');
            }
        });
        context.globalState.update('codehawk.hasShownWelcome', true);
    }
}

export function deactivate() {
    if (statusBarItem) {
        statusBarItem.dispose();
    }
    console.log('CodeHawk extension has been deactivated');
}