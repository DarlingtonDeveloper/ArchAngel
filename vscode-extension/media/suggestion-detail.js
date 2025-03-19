// Get the VS Code API
const vscode = acquireVsCodeApi();

// Initialize when the document is loaded
document.addEventListener('DOMContentLoaded', function () {
    // Handle apply fix button
    const applyFixButton = document.getElementById('applyFixButton');
    if (applyFixButton) {
        applyFixButton.addEventListener('click', function () {
            vscode.postMessage({
                command: 'applySuggestion'
            });
        });
    }

    // Add line numbers to code context
    const codeContext = document.querySelector('.code-context');
    if (codeContext) {
        const highlightLine = codeContext.querySelector('.highlight-line');
        if (highlightLine) {
            // Add line number to the highlighted line
            const lineNumberMatch = highlightLine.nextElementSibling ?
                highlightLine.nextElementSibling.textContent.match(/(\d+)$/) : null;

            if (lineNumberMatch) {
                const lineNumber = parseInt(lineNumberMatch[1]) - 1;
                highlightLine.innerHTML += `<span class="line-number">${lineNumber}</span>`;
            }

            // Add styling for better visibility
            highlightLine.style.display = 'block';
            highlightLine.style.padding = '4px 40px 4px 8px';
        }
    }

    // Save the state for restore when the view is hidden and restored
    vscode.setState({ initialized: true });
});

// Handle theme changes or other updates from VS Code
window.addEventListener('message', function (event) {
    const message = event.data;

    switch (message.command) {
        case 'updateTheme':
            // Update theme-specific elements if needed
            break;

        case 'refreshContent':
            // Refresh the content if needed
            location.reload();
            break;
    }
});