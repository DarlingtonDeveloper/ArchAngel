import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import axios from 'axios';
import { getApiClient } from '../../utils/api';
import { DiagnosticManager } from '../../utils/diagnosticManager';

suite('CodeHawk Extension Test Suite', () => {
    let axiosStub: sinon.SinonStub;

    setup(() => {
        // Stub axios requests
        axiosStub = sinon.stub(axios, 'request');
    });

    teardown(() => {
        // Restore axios stub
        axiosStub.restore();
    });

    test('Extension should be present', () => {
        assert.ok(vscode.extensions.getExtension('codehawk.codehawk'));
    });

    test('Extension should register commands', async () => {
        const commands = await vscode.commands.getCommands(true);

        assert.ok(commands.includes('codehawk.analyzeCurrentFile'));
        assert.ok(commands.includes('codehawk.analyzeSelection'));
        assert.ok(commands.includes('codehawk.showResults'));
        assert.ok(commands.includes('codehawk.configure'));
    });

    test('DiagnosticManager should set diagnostics', () => {
        const manager = DiagnosticManager.getInstance();

        // Create a dummy document
        const document = {
            uri: vscode.Uri.parse('file:///test/document.js'),
            lineAt: (line: number) => {
                return {
                    text: 'const test = "value";',
                    range: new vscode.Range(line, 0, line, 20)
                };
            },
            getText: () => 'const test = "value";'
        } as unknown as vscode.TextDocument;

        // Create test issues
        const issues = [
            {
                line: 1,
                column: 7,
                message: 'Test issue',
                severity: 'warning',
                ruleId: 'test-rule'
            }
        ];

        // Set diagnostics
        manager.setDiagnostics(document, issues);

        // Check if diagnostics were set correctly
        const diagnostics = vscode.languages.getDiagnostics(document.uri);
        assert.strictEqual(diagnostics.length, 1);
        assert.strictEqual(diagnostics[0].message, 'Test issue');
        assert.strictEqual(diagnostics[0].severity, vscode.DiagnosticSeverity.Warning);
        assert.strictEqual(diagnostics[0].code, 'test-rule');

        // Clear diagnostics
        manager.clearDiagnostics(document.uri);

        // Check if diagnostics were cleared
        const clearedDiagnostics = vscode.languages.getDiagnostics(document.uri);
        assert.strictEqual(clearedDiagnostics.length, 0);
    });

    test('API client should make requests with correct headers', async () => {
        // Configure API client
        await vscode.workspace.getConfiguration('codehawk').update('apiUrl', 'http://test-api.com', true);
        await vscode.workspace.getConfiguration('codehawk').update('apiKey', 'test-api-key', true);

        // Set up stub response
        axiosStub.resolves({
            data: { status: 'success' }
        });

        // Get API client
        const client = getApiClient();

        // Make a test request
        await client.get('/test-endpoint');

        // Check if request was made with correct headers
        assert.ok(axiosStub.calledOnce);
        const config = axiosStub.getCall(0).args[0];
        assert.strictEqual(config.url, '/test-endpoint');
        assert.strictEqual(config.baseURL, 'http://test-api.com');
        assert.strictEqual(config.headers['X-API-Key'], 'test-api-key');
    });

    test('analyze command should process results correctly', async () => {
        // Mock API response
        axiosStub.resolves({
            data: {
                status: 'success',
                id: 'test-analysis-id',
                issues: [
                    {
                        line: 1,
                        column: 5,
                        message: 'Test issue',
                        severity: 'warning',
                        ruleId: 'test-rule'
                    }
                ]
            }
        });

        // Create a dummy document
        await vscode.workspace.openTextDocument({
            content: 'const test = "value";',
            language: 'javascript'
        }).then(document => {
            return vscode.window.showTextDocument(document);
        });

        // Execute the command
        await vscode.commands.executeCommand('codehawk.analyzeCurrentFile');

        // Check if API request was made
        assert.ok(axiosStub.calledOnce);

        // Allow time for the diagnostics to be processed
        await new Promise(resolve => setTimeout(resolve, 100));

        // Check if diagnostics were set
        const document = vscode.window.activeTextEditor?.document;
        if (document) {
            const diagnostics = vscode.languages.getDiagnostics(document.uri);
            assert.ok(diagnostics.length > 0);
            assert.strictEqual(diagnostics[0].message, 'Test issue');
        }
    });
});