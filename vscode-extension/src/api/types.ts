/**
 * API request and response types for the CodeHawk API
 */

export interface AnalysisRequest {
    /** The code to analyze */
    code: string;

    /** The programming language */
    language: string;

    /** Optional context for the analysis */
    context?: string;

    /** Optional analysis options */
    options?: Record<string, any>;
}

export interface AnalysisResponse {
    /** Analysis ID */
    id: string;

    /** Status of the analysis */
    status: string;

    /** Programming language */
    language: string;

    /** Context for the analysis */
    context?: string;

    /** Timestamp of the analysis */
    timestamp: string;

    /** Issues found in the code */
    issues: Issue[];

    /** Suggestions for code improvement */
    suggestions: Suggestion[];

    /** Whether AI was used for suggestions */
    aiEnhanced?: boolean;
}

export interface Issue {
    /** Line number (1-based) */
    line: number;

    /** Column number (1-based, optional) */
    column?: number;

    /** Description of the issue */
    message: string;

    /** Severity of the issue */
    severity: 'error' | 'warning' | 'suggestion' | 'info';

    /** Rule identifier */
    ruleId?: string;

    /** Additional context */
    context?: string;

    /** Suggested fix */
    fix?: IssueFix;
}

export interface Suggestion {
    /** Line number (1-based) */
    line: number;

    /** Column number (1-based, optional) */
    column?: number;

    /** Description of the suggestion */
    message: string;

    /** Severity (typically 'suggestion') */
    severity: string;

    /** Rule identifier */
    ruleId?: string;

    /** Additional context */
    context?: string;

    /** Suggested fix */
    fix?: IssueFix;

    /** AI confidence score (if applicable) */
    confidence?: number;
}

export interface IssueFix {
    /** Description of the fix */
    description: string;

    /** Replacement text */
    replacement: string;
}

export interface LanguagesResponse {
    /** Status of the request */
    status: string;

    /** Supported languages */
    languages: string[];
}

export interface RulesResponse {
    /** Status of the request */
    status: string;

    /** Language for the rules */
    language: string;

    /** Rules for the language */
    rules: Rule[];
}

export interface Rule {
    /** Rule identifier */
    id: string;

    /** Rule name */
    name: string;

    /** Rule description */
    description: string;

    /** Default severity */
    severity: 'error' | 'warning' | 'suggestion' | 'info';
}

export interface ErrorResponse {
    /** Status of the request (always 'error') */
    status: string;

    /** Error message */
    message: string;

    /** Error code (optional) */
    code?: string;
}