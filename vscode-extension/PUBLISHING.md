# Publishing to VS Code Marketplace

This guide explains how to publish the Kloudlite Workspace extension to the Visual Studio Code Marketplace.

## Prerequisites

1. **Microsoft Account**: You need a Microsoft account (e.g., @outlook.com, @hotmail.com)
2. **Azure DevOps Organization**: Create a free organization at https://dev.azure.com
3. **Personal Access Token (PAT)**: Generate a token with marketplace publishing permissions

## Step 1: Create Azure DevOps Organization

1. Go to https://dev.azure.com
2. Sign in with your Microsoft account
3. Create a new organization (e.g., "kloudlite")
4. Keep the organization name - you'll need it later

## Step 2: Generate Personal Access Token

1. In Azure DevOps, click on your profile icon (top right)
2. Select "Personal Access Tokens"
3. Click "+ New Token"
4. Configure the token:
   - Name: "VS Code Marketplace Publishing"
   - Organization: Select "All accessible organizations"
   - Expiration: Set to 90 days or custom
   - Scopes: Click "Show all scopes" and select:
     - **Marketplace**: "Acquire" and "Manage"
5. Click "Create"
6. **IMPORTANT**: Copy the token immediately (you won't see it again!)

## Step 3: Create Publisher

1. Go to https://marketplace.visualstudio.com/manage
2. Sign in with the same Microsoft account
3. Click "Create publisher"
4. Fill in the publisher details:
   - **ID**: `kloudlite` (must match package.json)
   - **Name**: "Kloudlite"
   - **Email**: Your verified email address
5. Click "Create"

## Step 4: Prepare Extension for Publishing

### Add Icon (Optional but Recommended)

The extension needs a 128x128 PNG icon. To create one from the SVG:

**Option 1: Use an online converter**
1. Go to https://cloudconvert.com/svg-to-png
2. Upload `resources/icon.svg`
3. Set dimensions to 128x128
4. Download the PNG and save as `resources/icon.png`

**Option 2: Use ImageMagick (if installed)**
```bash
convert -background none -resize 128x128 resources/icon.svg resources/icon.png
```

**Option 3: Use Node.js sharp library**
```bash
npm install sharp
node -e "require('sharp')('resources/icon.svg').resize(128,128).png().toFile('resources/icon.png')"
```

Then update `package.json`:
```json
"icon": "resources/icon.png",
```

### Update Version (if needed)

If you're publishing a new version, update the version in `package.json`:
```json
"version": "0.2.0",
```

And add release notes to `CHANGELOG.md`.

## Step 5: Build and Package

```bash
# Install dependencies
npm install

# Compile TypeScript
npm run compile

# Package extension
vsce package
```

This creates a `.vsix` file (e.g., `kloudlite-workspace-0.1.19.vsix`)

## Step 6: Publish Extension

### Option 1: Publish via Command Line

```bash
# Login with your PAT
vsce login kloudlite
# Paste your Personal Access Token when prompted

# Publish the extension
vsce publish
```

### Option 2: Publish via Web Interface

1. Go to https://marketplace.visualstudio.com/manage/publishers/kloudlite
2. Click "+ New extension" > "Visual Studio Code"
3. Drag and drop your `.vsix` file
4. Click "Upload"

## Step 7: Verify Publication

1. Visit https://marketplace.visualstudio.com/items?itemName=kloudlite.kloudlite-workspace
2. Check that all information is correct:
   - Name, description, icon
   - Screenshots (if added)
   - README content
   - Version number

## Publishing Updates

When you make changes and want to publish a new version:

1. Update version in `package.json` (follow [semantic versioning](https://semver.org/))
2. Update `CHANGELOG.md` with the changes
3. Commit your changes
4. Run `vsce publish [major|minor|patch]`
   - `vsce publish patch` - Bug fixes (0.1.19 → 0.1.20)
   - `vsce publish minor` - New features (0.1.19 → 0.2.0)
   - `vsce publish major` - Breaking changes (0.1.19 → 1.0.0)

## Important Notes

- **Icon**: Marketplace requires a PNG icon (SVG not supported by vsce)
- **Publisher ID**: Must match the `publisher` field in package.json
- **License**: MIT license is included (LICENSE file)
- **Repository**: Links to GitHub repository for issues and source code
- **README**: Will be displayed on marketplace page
- **Categories**: Extension is categorized as "Other" - you may want to update this

## Troubleshooting

### "Publisher 'kloudlite' not found"
- Make sure you created the publisher at https://marketplace.visualstudio.com/manage
- Verify the publisher ID matches exactly

### "Personal Access Token is invalid"
- Generate a new PAT with correct scopes (Marketplace: Acquire, Manage)
- Make sure you're using the full token (not truncated)

### "SVGs can't be used as icons"
- Convert the SVG to PNG (128x128) as described above
- Update package.json to reference the PNG file

### Icon not displaying
- Ensure icon is exactly 128x128 pixels
- File size should be under 256KB
- Use PNG format

## Additional Resources

- [VS Code Extension Publishing Guide](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
- [vsce Documentation](https://github.com/microsoft/vscode-vsce)
- [Marketplace Publisher Portal](https://marketplace.visualstudio.com/manage)
- [Extension Manifest Reference](https://code.visualstudio.com/api/references/extension-manifest)
