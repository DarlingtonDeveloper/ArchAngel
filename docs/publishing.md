# Publishing the CodeHawk VS Code Extension

This guide will walk you through the process of publishing the CodeHawk VS Code extension to the Visual Studio Code Marketplace.

## Prerequisites

Before you start, make sure you have the following:

1. **Node.js and npm**: Latest LTS version recommended
2. **Visual Studio Code Extension Manager (vsce)**: Install it globally with `npm install -g vsce`
3. **Azure DevOps Account**: Required for obtaining a Personal Access Token (PAT)
4. **Visual Studio Marketplace Publisher Account**: You'll need to create a publisher on the VS Code Marketplace

## Step 1: Prepare Your Extension

1. **Update package.json** with accurate metadata:
   - Make sure `name`, `displayName`, `description`, `version`, and `publisher` are set correctly
   - Ensure all required fields are present (repository, license, etc.)
   - Update the `engines.vscode` field to specify the minimum VS Code version required

2. **Create or update README.md** with:
   - Clear description of the extension
   - Features list with screenshots
   - Installation and usage instructions
   - Configuration options
   - FAQ section

3. **Prepare visual assets**:
   - Create a compelling icon (at least 128x128 pixels)
   - Prepare screenshots for the marketplace listing
   - Consider creating animated GIFs to demonstrate features

4. **Update CHANGELOG.md** to reflect the latest changes

## Step 2: Create a Publisher

If you haven't already created a publisher on the VS Code Marketplace:

1. Go to [Visual Studio Marketplace](https://marketplace.visualstudio.com/manage)
2. Sign in with your Azure DevOps account
3. Click on "Create Publisher"
4. Fill in the required information (Publisher ID, Display Name, etc.)
5. Agree to the terms and create your publisher

## Step 3: Get a Personal Access Token (PAT)

1. Go to [Azure DevOps](https://dev.azure.com/)
2. Click on your profile icon in the top right corner
3. Select "Personal access tokens"
4. Click "New Token"
5. Name your token (e.g., "VS Code Extension Publishing")
6. Set the organization to "All accessible organizations"
7. Set the expiration time (1 year is recommended)
8. Under "Scopes," select "Custom defined" and check "Marketplace > Manage"
9. Click "Create" and copy the generated token (you won't be able to see it again)

## Step 4: Package and Publish the Extension

### Manual Publishing

1. **Login to vsce**:
   ```bash
   vsce login <publisher-name>
   ```
   When prompted, paste the Personal Access Token you created.

2. **Package the extension**:
   ```bash
   cd codehawk/vscode-extension
   vsce package
   ```
   This creates a `.vsix` file in your directory.

3. **Test the packaged extension**:
   ```bash
   code --install-extension codehawk-0.1.0.vsix
   ```
   Verify that the extension works correctly.

4. **Publish the extension**:
   ```bash
   vsce publish
   ```
   This will publish the extension to the VS Code Marketplace.

### Automated Publishing

For automated publishing, we use GitHub Actions. The workflow is already set up in `.github/workflows/deploy.yml`.

1. Add your PAT as a GitHub Secret:
   - Go to your GitHub repository
   - Navigate to Settings > Secrets > Actions
   - Click "New repository secret"
   - Name: `VSCE_PAT`
   - Value: Your Personal Access Token
   - Click "Add secret"

2. Create a new release tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

3. The GitHub Actions workflow will automatically:
   - Build the extension
   - Package it
   - Publish it to the VS Code Marketplace

## Step 5: Update the Extension

When you need to update the extension:

1. Make your changes to the code
2. Update the version in `package.json` (following [Semantic Versioning](https://semver.org/))
3. Update `CHANGELOG.md` with the changes
4. Commit your changes
5. Create a new tag:
   ```bash
   git tag v0.1.1
   git push origin v0.1.1
   ```
6. The GitHub Actions workflow will handle the rest

## Best Practices

- **Incremental Updates**: Make small, frequent updates rather than large, infrequent ones
- **Testing**: Thoroughly test your extension before publishing
- **Documentation**: Keep documentation up-to-date with each release
- **Feedback**: Monitor user reviews and issues on GitHub to guide improvements
- **Version Numbering**: Follow semantic versioning (MAJOR.MINOR.PATCH)

## Troubleshooting

- **Authentication Issues**: 
  - Ensure your PAT has the correct permissions
  - Check that the PAT hasn't expired
  
- **Publishing Errors**:
  - Verify all required fields in `package.json`
  - Make sure the version hasn't already been published
  
- **Extension Not Appearing**:
  - Marketplaces can take a few minutes to update
  - Verify the extension was published successfully

- **Installation Issues**:
  - Check the extension's compatibility with the VS Code version
  - Look for conflicting extensions

## Resources

- [VS Code Extension Publishing Documentation](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
- [vsce Command Line Tool Documentation](https://github.com/microsoft/vscode-vsce)
- [Azure DevOps Personal Access Tokens Documentation](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate)