// CodeHawk API Client Example - JavaScript
// This example shows how to use the CodeHawk API in a Node.js application

const axios = require('axios');

class CodeHawkClient {
    constructor(apiKey, apiUrl = 'https://api.codehawk.dev/api/v1') {
        this.apiKey = apiKey;
        this.apiUrl = apiUrl;

        // Configure axios instance
        this.client = axios.create({
            baseURL: this.apiUrl,
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                'X-API-Key': this.apiKey
            },
            timeout: 30000 // 30 seconds timeout
        });
    }

    /**
     * Submit code for analysis
     * @param {string} code - The code to analyze
     * @param {string} language - The programming language
     * @param {string} context - Optional context for the analysis
     * @param {object} options - Optional analysis options
     * @returns {Promise<object>} - Analysis result
     */
    async analyzeCode(code, language, context = '', options = {}) {
        try {
            const response = await this.client.post('/analyze', {
                code,
                language,
                context,
                options
            });

            return response.data;
        } catch (error) {
            this._handleError(error);
        }
    }

    /**
     * Get analysis by ID
     * @param {string} id - Analysis ID
     * @returns {Promise<object>} - Analysis result
     */
    async getAnalysis(id) {
        try {
            const response = await this.client.get(`/analysis/${id}`);
            return response.data;
        } catch (error) {
            this._handleError(error);
        }
    }

    /**
     * Get issues for an analysis
     * @param {string} id - Analysis ID
     * @param {string} severity - Optional severity filter
     * @returns {Promise<object>} - Issues
     */
    async getIssues(id, severity = null) {
        try {
            const params = severity ? { severity } : {};
            const response = await this.client.get(`/analysis/${id}/issues`, { params });
            return response.data;
        } catch (error) {
            this._handleError(error);
        }
    }

    /**
     * Get suggestions for an analysis
     * @param {string} id - Analysis ID
     * @returns {Promise<object>} - Suggestions
     */
    async getSuggestions(id) {
        try {
            const response = await this.client.get(`/analysis/${id}/suggestions`);
            return response.data;
        } catch (error) {
            this._handleError(error);
        }
    }

    /**
     * Get supported languages
     * @returns {Promise<object>} - Languages
     */
    async getLanguages() {
        try {
            const response = await this.client.get('/languages');
            return response.data;
        } catch (error) {
            this._handleError(error);
        }
    }

    /**
     * Get rules for a language
     * @param {string} language - Programming language
     * @returns {Promise<object>} - Rules
     */
    async getRules(language) {
        try {
            const response = await this.client.get(`/rules/${language}`);
            return response.data;
        } catch (error) {
            this._handleError(error);
        }
    }

    /**
     * Handle API errors
     * @private
     * @param {Error} error - Axios error
     */
    _handleError(error) {
        if (error.response) {
            // The request was made and the server responded with a status code
            // that falls out of the range of 2xx
            const { status, data } = error.response;

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
                    throw new Error(data.message || `API error: ${status}`);
            }
        } else if (error.request) {
            // The request was made but no response was received
            throw new Error('No response from server. Please check your network connection.');
        } else {
            // Something happened in setting up the request
            throw new Error(`Error: ${error.message}`);
        }
    }
}

// Example usage
async function main() {
    const apiKey = process.env.CODEHAWK_API_KEY || 'your-api-key';
    const client = new CodeHawkClient(apiKey);

    try {
        // Analyze code
        const code = `function sum(a, b) {
  return a + b;
}`;

        console.log('Analyzing code...');
        const analysis = await client.analyzeCode(code, 'javascript');
        console.log(`Analysis ID: ${analysis.id}`);
        console.log(`Found ${analysis.issues.length} issues`);

        // Get suggestions
        console.log('Getting suggestions...');
        const suggestions = await client.getSuggestions(analysis.id);
        console.log(`Found ${suggestions.suggestions.length} suggestions`);

        // Get supported languages
        console.log('Getting supported languages...');
        const languages = await client.getLanguages();
        console.log(`Supported languages: ${languages.languages.join(', ')}`);
    } catch (error) {
        console.error('Error:', error.message);
    }
}

// Run the example if this file is executed directly
if (require.main === module) {
    main();
}

module.exports = CodeHawkClient;