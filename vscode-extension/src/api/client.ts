import axios, { AxiosInstance, AxiosError } from 'axios';
import * as vscode from 'vscode';
import {
    AnalysisRequest,
    AnalysisResponse,
    Issue,
    LanguagesResponse,
    RulesResponse,
    Suggestion,
    ErrorResponse
} from './types';

/**
 * CodeHawk API client
 */
export class CodeHawkApiClient {
    private client: AxiosInstance;
    private retryCount: number = 3;
    private retryDelay: number = 500;

    /**
     * Create a new API client
     * @param apiUrl URL of the CodeHawk API
     * @param apiKey API key for authentication
     */
    constructor(apiUrl: string, apiKey: string) {
        this.client = axios.create({
            baseURL: apiUrl,
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                'X-API-Key': apiKey
            },
            timeout: 30000 // 30 seconds timeout
        });

        // Add interceptors for error handling
        this.client.interceptors.response.use(
            response => response,
            error => this.handleError(error)
        );
    }

    /**
     * Analyze code
     * @param request Analysis request
     * @returns Analysis response
     */
    public async analyzeCode(request: AnalysisRequest): Promise<AnalysisResponse> {
        return await this.withRetry(async () => {
            const response = await this.client.post<AnalysisResponse>('/api/v1/analyze', request);
            return response.data;
        });
    }

    /**
     * Get analysis by ID
     * @param id Analysis ID
     * @returns Analysis response
     */
    public async getAnalysis(id: string): Promise<AnalysisResponse> {
        return await this.withRetry(async () => {
            const response = await this.client.get<AnalysisResponse>(`/api/v1/analysis/${id}`);
            return response.data;
        });
    }

    /**
     * Get issues for an analysis
     * @param id Analysis ID
     * @param severity Optional severity filter
     * @returns Issues
     */
    public async getIssues(id: string, severity?: string): Promise<Issue[]> {
        return await this.withRetry(async () => {
            const params = severity ? { severity } : {};
            const response = await this.client.get<{ issues: Issue[] }>(`/api/v1/analysis/${id}/issues`, { params });
            return response.data.issues;
        });
    }

    /**
     * Get suggestions for an analysis
     * @param id Analysis ID
     * @returns Suggestions
     */
    public async getSuggestions(id: string): Promise<Suggestion[]> {
        return await this.withRetry(async () => {
            const response = await this.client.get<{ suggestions: Suggestion[] }>(`/api/v1/analysis/${id}/suggestions`);
            return response.data.suggestions;
        });
    }

    /**
     * Get supported languages
     * @returns Supported languages
     */
    public async getLanguages(): Promise<string[]> {
        return await this.withRetry(async () => {
            const response = await this.client.get<LanguagesResponse>('/api/v1/languages');
            return response.data.languages;
        });
    }

    /**
     * Get rules for a language
     * @param language Programming language
     * @returns Rules
     */
    public async getRules(language: string): Promise<RulesResponse> {
        return await this.withRetry(async () => {
            const response = await this.client.get<RulesResponse>(`/api/v1/rules/${language}`);
            return response.data;
        });
    }

    /**
     * Handle API errors
     * @param error Error object
     */
    private handleError(error: unknown): never {
        // Check if it's an Axios error
        if (axios.isAxiosError(error)) {
            const axiosError = error as AxiosError<ErrorResponse>;

            if (axiosError.response) {
                // The request was made and the server responded with a status code
                // that falls out of the range of 2xx
                const status = axiosError.response.status;
                const data = axiosError.response.data;

                switch (status) {
                    case 401:
                        throw new Error('Authentication failed. Please check your API key.');
                    case 403:
                        throw new Error('Access denied. You do not have permission to perform this action.');
                    case 404:
                        throw new Error('The requested resource was not found.');
                    case 429:
                        throw new Error('API rate limit exceeded. Please try again later.');
                    default:
                        throw new Error(data?.message || `API error: ${status}`);
                }
            } else if (axiosError.request) {
                // The request was made but no response was received
                throw new Error('No response from server. Please check your network connection.');
            } else {
                // Something happened in setting up the request
                throw new Error(`Error: ${axiosError.message}`);
            }
        }

        // For non-Axios errors
        throw new Error(`Unexpected error: ${String(error)}`);
    }

    /**
     * Retry a function with exponential backoff
     * @param fn Function to retry
     * @returns Result of the function
     */
    private async withRetry<T>(fn: () => Promise<T>): Promise<T> {
        let lastError: Error | null = null;
        let delay = this.retryDelay;

        for (let attempt = 0; attempt < this.retryCount + 1; attempt++) {
            try {
                return await fn();
            } catch (error) {
                lastError = error as Error;

                // Check if error is retryable
                if (!this.isRetryableError(error)) {
                    throw error;
                }

                // Last attempt, don't wait
                if (attempt === this.retryCount) {
                    break;
                }

                // Wait before retry with exponential backoff
                await new Promise(resolve => setTimeout(resolve, delay));
                delay *= 2; // Double the delay for each retry
            }
        }

        throw lastError || new Error('Max retries exceeded');
    }

    /**
     * Check if an error is retryable
     * @param error Error to check
     * @returns True if the error is retryable
     */
    private isRetryableError(error: unknown): boolean {
        if (axios.isAxiosError(error)) {
            const axiosError = error as AxiosError;

            // Network errors are retryable
            if (!axiosError.response) {
                return true;
            }

            // 429 (Too Many Requests) and 5xx errors are retryable
            const status = axiosError.response.status;
            return status === 429 || status >= 500;
        }

        return false;
    }
}

/**
 * Get API client from VS Code settings
 * @returns CodeHawk API client
 */
export function getApiClient(): CodeHawkApiClient {
    const config = vscode.workspace.getConfiguration('codehawk');
    const apiUrl = config.get<string>('apiUrl') || 'http://localhost:8080';
    const apiKey = config.get<string>('apiKey') || '';

    return new CodeHawkApiClient(apiUrl, apiKey);
}