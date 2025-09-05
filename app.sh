#!/bin/bash

# Script to build translator.app for macOS
# This creates a proper macOS application bundle in dist/ directory

set -e

echo "🚀 Building translator.app in dist/ directory..."

# Create dist directory if it doesn't exist
echo "📁 Creating dist directory..."
rm -rf dist
mkdir -p dist

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf dist/translator.app
rm -f dist/translator

# Build the Go binary
echo "🔨 Building Go binary..."
go build -ldflags="-s -w" -o dist/translator main.go

# Create app bundle structure
echo "📦 Creating app bundle structure..."
mkdir -p dist/translator.app/Contents/MacOS
mkdir -p dist/translator.app/Contents/Resources

# Copy binary to app bundle
echo "📋 Copying binary to app bundle..."
cp dist/translator dist/translator.app/Contents/MacOS/

# Create icon set from icon.png
echo "🎨 Creating app icon..."
if [ -f icon.png ]; then
    # Create .iconset directory
    mkdir -p dist/translator.app/Contents/Resources/icon.iconset
    
    # Generate different icon sizes for macOS
    echo "   Creating icon sizes..."
    sips -z 16 16 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_16x16.png
    sips -z 32 32 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_16x16@2x.png
    sips -z 32 32 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_32x32.png
    sips -z 64 64 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_32x32@2x.png
    sips -z 128 128 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_128x128.png
    sips -z 256 256 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_128x128@2x.png
    sips -z 256 256 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_256x256.png
    sips -z 512 512 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_256x256@2x.png
    sips -z 512 512 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_512x512.png
    sips -z 1024 1024 icon.png --out dist/translator.app/Contents/Resources/icon.iconset/icon_512x512@2x.png
    
    # Create .icns file
    echo "   Creating .icns file..."
    iconutil -c icns dist/translator.app/Contents/Resources/icon.iconset -o dist/translator.app/Contents/Resources/icon.icns
    
    # Clean up .iconset directory
    rm -rf dist/translator.app/Contents/Resources/icon.iconset
    
    echo "✅ Icon created successfully!"
else
    echo "⚠️  icon.png not found, skipping icon creation"
fi

# Create Info.plist
echo "📝 Creating Info.plist..."
cat > dist/translator.app/Contents/Info.plist << 'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"\>
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>translator</string>
    <key>CFBundleIdentifier</key>
    <string>com.translator.hotkey</string>
    <key>CFBundleName</key>
    <string>Hotkey Translator</string>
    <key>CFBundleDisplayName</key>
    <string>Hotkey Translator</string>
    <key>CFBundleVersion</key>
    <string>1.0</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleSignature</key>
    <string>????</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
    <key>NSAppleScriptEnabled</key>
    <true/>
    <key>NSAppleEventsUsageDescription</key>
    <string>This app needs AppleScript access to simulate keyboard events for translation.</string>
    <key>CFBundleIconFile</key>
    <string>icon</string>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2024 Hotkey Translator. All rights reserved.</string>
    <key>LSEnvironment</key>
    <dict>
        <key>LANG</key>
        <string>en_US.UTF-8</string>
        <key>LC_ALL</key>
        <string>en_US.UTF-8</string>
        <key>LC_CTYPE</key>
        <string>en_US.UTF-8</string>
    </dict>
</dict>
</plist>
PLIST

# Copy config files to app bundle (if they exist)
if [ -f config.json ]; then
    echo "📋 Copying config.json to app bundle..."
    cp config.json dist/translator.app/Contents/Resources/
fi

if [ -f .env ]; then
    echo "📋 Copying .env file to app bundle..."
    cp .env dist/translator.app/Contents/Resources/
fi

# Set permissions
echo "🔐 Setting permissions..."
chmod +x dist/translator.app/Contents/MacOS/translator

# Create a launcher script with proper Unicode support
echo "�� Creating launcher script..."
cat > dist/translator.app/Contents/MacOS/launcher.sh << 'LAUNCHER'
#!/bin/bash

# Set Unicode environment variables
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8
export LC_CTYPE=en_US.UTF-8

# Navigate to the app directory
cd "$(dirname "$0")"

# Run the translator
./translator
LAUNCHER

chmod +x dist/translator.app/Contents/MacOS/launcher.sh

# Update Info.plist to use launcher
sed -i '' 's|<string>translator</string>|<string>launcher.sh</string>|' dist/translator.app/Contents/Info.plist
# zip the app
echo "🔐 Zipping app..."
zip -r dist/app.zip dist/translator.app

echo "✅ translator.app created successfully in dist/ directory!"
echo ""
echo "📁 Build output structure:"
echo "   dist/"
echo "   ├── translator (standalone binary)"
echo "   └── translator.app/"
echo "       └── Contents/"
echo "           ├── Info.plist (with Unicode environment)"
echo "           ├── MacOS/"
echo "           │   ├── translator (binary)"
echo "           │   └── launcher.sh (with UTF-8 support)"
echo "           └── Resources/"
echo "               ├── icon.icns"
echo "               ├── config.json (if exists)"
echo "               └── .env (if exists)"
echo ""
echo "🚀 To run the app:"
echo "   open dist/translator.app"
echo "   # or run the standalone binary:"
echo "   ./dist/translator"
echo ""
echo "📋 To install system-wide:"
echo "   cp -r dist/translator.app /Applications/"
echo ""
echo "⚠️  Don't forget to:"
echo "   1. Grant Accessibility permissions in System Preferences"
echo "   2. Grant AppleScript permissions when prompted"
echo "   3. Enter your Gemini API key in the app's text field"
echo ""
echo "🎨 App icon has been created from icon.png"
echo "💾 Config will be saved to config.json in the app bundle"
echo "🌐 Unicode support enabled for Japanese text"
echo "📦 All build artifacts are organized in the dist/ directory"
