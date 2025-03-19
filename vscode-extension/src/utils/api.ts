import * as vscode from 'vscode';
import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';

let apiClient: AxiosInstance | null = null;

/**
 * Initialize the API client with configuration from VS Code settings
 */
export function initializeApiClient(): AxiosInstance {
    const config = vscode.workspace.getConfiguration('codehawk');
    const apiUrl = config.get<string>('apiUrl');
    const apiKey = config.get<string>('apiKey');

    const axiosConfig: AxiosRequestConfig = {
        baseURL: apiUrl,
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
        },
        timeout: 30000 // 30 seconds timeout
    };

    if (apiKey) {
        axiosConfig.headers!['X-API-Key'] = apiKey;
    }

    apiClient = axios.create(axiosConfig);

    // Add a response interceptor to handle errors
    apiClient.interceptors.response.use(
        response => response,
        error => {
            if (error.response) {
                // The request was made and the server responded with a status code
                // that falls out of the range of 2xx
                const { status, data } = error.response;

                switch (status) {
                    case 401:
                        vscode.window.showErrorMessage('CodeHawk API authentication failed. Please check your API key.');
                        break;
                    case 403:
                        vscode.window.showErrorMessage('Access denied. You do not have permission to perform this action.');
                        break;
                    case 429:
                        vscode.window.showErrorMessage('API rate limit exceeded. Please try again later.');
                        break;
                    default:
                        vscode.window.showErrorMessage(`CodeHawk API error: ${data.message || 'Unknown error'}`);
                }
            } else if (error.request) {
                // The request was made but no response was received
                vscode.window.showErrorMessage(`Cannot connect to CodeHawk API server. Please check your network and API URL.`);
            } else {
                // Something happened in setting up the request that triggered an Error
                vscode.window.showErrorMessage(`Error setting up API request: ${error.message}`);
            }

            return Promise.reject(error);
        }
    );

    return apiClient;
}

/**
 * Get the API client, initializing it if needed
 */
export function getApiClient(): AxiosInstance {
    if (!apiClient) {
        return initializeApiClient();
    }
    return apiClient;
}

/**
 * Check if the API is properly configured
 */
export async function checkApiConfiguration(): Promise<boolean> {
    const config = vscode.workspace.getConfiguration('codehawk');
    const apiUrl = config.get<string>('apiUrl');
    const apiKey = config.get<string>('apiKey');

    if (!apiUrl) {
        const configure = 'Configure';
        const response = await vscode.window.showErrorMessage(
            'CodeHawk API URL is not configured',
            configure
        );

        if (response === configure) {
            vscode.commands.executeCommand('codehawk.configure');
        }

        return false;
    }

    if (!apiKey) {
        const configure = 'Configure';
        const response = await vscode.window.showErrorMessage(
            'CodeHawk API key is not configured',
            configure
        );

        if (response === configure) {
            vscode.commands.executeCommand('codehawk.configure');
        }

        return false;
    }

    return true;
}

/**
 * Send code to the API for analysis
 */
export async function analyzeCode(code: string, language: string, context?: string): Promise<any> {
    const client = getApiClient();

    try {
        const response = await client.post('/api/v1/analyze', {
            code,
            language,
            context: context || ''
        });

        return response.data;
    } catch (error) {
        console.error('API error during code analysis:', error);
        throw error;
    }
}

/**
 * Get analysis results by ID
 */
export async function getAnalysisById(id: string): Promise<any> {
    const client = getApiClient();

    try {
        const response = await client.get(`/api/v1/analysis/${id}`);
        return response.data;
    } catch (error) {
        console.error('API error while fetching analysis results:', error);
        throw error;
    }
}