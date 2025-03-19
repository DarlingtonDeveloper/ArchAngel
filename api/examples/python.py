#!/usr/bin/env python3
# CodeHawk API Client Example - Python
# This example shows how to use the CodeHawk API in a Python application

import os
import requests
import json
from typing import Dict, List, Optional, Any, Union


class CodeHawkClient:
    """Client for interacting with the CodeHawk API"""
    
    def __init__(self, api_key: str, api_url: str = "https://api.codehawk.dev/api/v1"):
        """
        Initialize the CodeHawk API client
        
        Args:
            api_key: Your CodeHawk API key
            api_url: The API base URL (default: production API)
        """
        self.api_key = api_key
        self.api_url = api_url
        self.session = requests.Session()
        self.session.headers.update({
            "Content-Type": "application/json",
            "Accept": "application/json",
            "X-API-Key": api_key
        })
    
    def analyze_code(self, code: str, language: str, context: str = "", 
                    options: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Submit code for analysis
        
        Args:
            code: The code to analyze
            language: The programming language
            context: Optional context for the analysis
            options: Optional analysis options
            
        Returns:
            Analysis result
        """
        if options is None:
            options = {}
        
        data = {
            "code": code,
            "language": language,
            "context": context,
            "options": options
        }
        
        response = self._make_request("POST", "/analyze", json=data)
        return response
    
    def get_analysis(self, analysis_id: str) -> Dict[str, Any]:
        """
        Get analysis by ID
        
        Args:
            analysis_id: Analysis ID
            
        Returns:
            Analysis result
        """
        response = self._make_request("GET", f"/analysis/{analysis_id}")
        return response
    
    def get_issues(self, analysis_id: str, severity: Optional[str] = None) -> Dict[str, Any]:
        """
        Get issues for an analysis
        
        Args:
            analysis_id: Analysis ID
            severity: Optional severity filter (error, warning, suggestion, info)
            
        Returns:
            Issues
        """
        params = {}
        if severity:
            params["severity"] = severity
        
        response = self._make_request("GET", f"/analysis/{analysis_id}/issues", params=params)
        return response
    
    def get_suggestions(self, analysis_id: str) -> Dict[str, Any]:
        """
        Get suggestions for an analysis
        
        Args:
            analysis_id: Analysis ID
            
        Returns:
            Suggestions
        """
        response = self._make_request("GET", f"/analysis/{analysis_id}/suggestions")
        return response
    
    def get_languages(self) -> Dict[str, Any]:
        """
        Get supported languages
        
        Returns:
            Languages
        """
        response = self._make_request("GET", "/languages")
        return response
    
    def get_rules(self, language: str) -> Dict[str, Any]:
        """
        Get rules for a language
        
        Args:
            language: Programming language
            
        Returns:
            Rules
        """
        response = self._make_request("GET", f"/rules/{language}")
        return response
    
    def _make_request(self, method: str, endpoint: str, 
                     params: Dict[str, str] = None, 
                     json: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        Make an API request
        
        Args:
            method: HTTP method
            endpoint: API endpoint
            params: Query parameters
            json: JSON request body
            
        Returns:
            API response
            
        Raises:
            Exception: If the API request fails
        """
        url = f"{self.api_url}{endpoint}"
        
        try:
            response = self.session.request(
                method=method,
                url=url,
                params=params,
                json=json,
                timeout=30  # 30 seconds timeout
            )
            
            # Raise for HTTP errors
            response.raise_for_status()
            
            return response.json()
        
        except requests.exceptions.HTTPError as e:
            if e.response is not None:
                status_code = e.response.status_code
                try:
                    error_data = e.response.json()
                    error_message = error_data.get("message", str(e))
                except ValueError:
                    error_message = str(e)
                
                if status_code == 401:
                    raise Exception("Authentication failed. Please check your API key.")
                elif status_code == 403:
                    raise Exception("Access denied. You do not have permission to perform this action.")
                elif status_code == 404:
                    raise Exception("The requested resource was not found.")
                elif status_code == 429:
                    raise Exception("API rate limit exceeded. Please try again later.")
                else:
                    raise Exception(f"API error ({status_code}): {error_message}")
            else:
                raise Exception(f"HTTP error: {str(e)}")
        
        except requests.exceptions.ConnectionError:
            raise Exception("Connection error. Please check your network connection.")
        
        except requests.exceptions.Timeout:
            raise Exception("Request timed out. Please try again later.")
        
        except requests.exceptions.RequestException as e:
            raise Exception(f"Request error: {str(e)}")
        
        except Exception as e:
            raise Exception(f"Unexpected error: {str(e)}")


def main():
    """Example usage of the CodeHawk API client"""
    # Get API key from environment variable or use a default value
    api_key = os.environ.get("CODEHAWK_API_KEY", "your-api-key")
    
    # Create client
    client = CodeHawkClient(api_key)
    
    try:
        # Analyze code
        code = """def calculate_sum(numbers):
    result = 0
    for n in numbers:
        result = result + n
    return result"""
        
        print("Analyzing code...")
        analysis = client.analyze_code(code, "python")
        print(f"Analysis ID: {analysis['id']}")
        print(f"Found {len(analysis['issues'])} issues")
        
        # Get suggestions
        print("Getting suggestions...")
        suggestions = client.get_suggestions(analysis["id"])
        print(f"Found {len(suggestions['suggestions'])} suggestions")
        
        # Get supported languages
        print("Getting supported languages...")
        languages = client.get_languages()
        print(f"Supported languages: {', '.join(languages['languages'])}")
        
    except Exception as e:
        print(f"Error: {str(e)}")


if __name__ == "__main__":
    main()