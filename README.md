# 🌐 Hotkey Translator

A powerful macOS application that provides instant AI-powered translation using global hotkeys. Select any text and translate it to English or both English and Japanese with simple keyboard shortcuts.

[![Buy me a coffee](https://img.shields.io/badge/Buy%20me%20a%20coffee-☕-yellow.svg)](https://buymeacoffee.com/dngteam)

## ✨ Features

- **Global Hotkeys**: Translate text from any application using `Control + Option + H` or `Control + Option + J`
- **Dual Translation**: Get both English and Japanese translations simultaneously
- **AI-Powered**: Uses Google Gemini API for high-quality translations
- **Auto-Improvement**: AI not only translates but also improves and rephrases text for better clarity
- **Multiple Models**: Support for various Gemini models (2.0 Flash Lite, 1.5 Flash, 1.5 Pro, 1.0 Pro)
- **Auto-Configuration**: Automatically saves settings and starts listening when API key is provided
- **Cross-Application**: Works with any text field or application on macOS

## 🚀 Quick Start

### Download app

[![Download App](https://img.shields.io/badge/Download-App-blue.svg)](https://github.com/dungnvpro/keyboard/raw/refs/heads/main/dist/app.zip)

### Prerequisites

- macOS (tested on macOS 10.15+)
- Go 1.24.1 or later
- Google Gemini API key

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/assistant-keyboard.git
   cd assistant-keyboard
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Build the application**
   ```bash
   go build -o disopen dist/translator.app main.go
   ```

4. **Run the application**
   ```bash
   open dist/translator.app
   ```

### Configuration

1. **Get a Gemini API Key**
   - Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
   - Create a new API key
   - Copy the API key

2. **Configure the Application**
   - Launch the application
   - Enter your Gemini API key in the text field
   - Select your preferred AI model
   - Click "🚀 Start Hotkey Listener"

3. **Grant Accessibility Permissions**
   - Go to System Preferences > Security & Privacy > Privacy > Accessibility
   - Add the `hotkey-translator` application
   - Or click "Open System Preferences" button in the app

## 🎯 Usage

### Hotkeys

- **`Control + Option + H`**: Translate selected text to English only
- **`Control + Option + J`**: Select all text and translate to both English and Japanese

### How to Use

1. **Select text** in any application (browser, text editor, etc.)
2. **Press the hotkey** (`Control + Option + H` or `Control + Option + J`)
3. **Wait for translation** (usually 1-3 seconds)
4. **Translated text** will automatically replace the selected text

### Example

**Original text**: "こんにちは、元気ですか？"

**After `Control + Option + H`**: "Hello, how are you?"

**After `Control + Option + J`**:
```
[EN]: Hello, how are you?
----------------
[JP]: こんにちは、元気ですか？
```

## 🛠️ Development

### Project Structure

```
assistant-keyboard/
├── main.go              # Main application code
├── go.mod              # Go module dependencies
├── go.sum              # Go module checksums
├── config.json         # Configuration file (auto-generated)
├── .env                # Environment variables (optional)
└── README.md           # This file
```

### Dependencies

- **Fyne**: Cross-platform GUI framework
- **gohook**: Global hotkey detection
- **godotenv**: Environment variable loading
- **Standard Go libraries**: HTTP, JSON, OS operations

### Configuration System

The application supports multiple configuration methods (in order of priority):

1. **config.json** (same directory as executable)
2. **.env file** (for development)
3. **Environment variables**

Example `config.json`:
```json
{
  "gemini_api_key": "your-api-key-here",
  "model": "gemini-2.0-flash-lite"
}
```

Example `.env`:
```env
GEMINI_API_KEY=your-api-key-here
```

### Building for Distribution

1. **Build for current platform**
   ```bash
   go build -o disopen dist/translator.app main.go
   ```

2. **Build for macOS (universal binary)**
   ```bash
   GOOS=darwin GOARCH=amd64 go build -o hotkey-translator-amd64 main.go
   GOOS=darwin GOARCH=arm64 go build -o hotkey-translator-arm64 main.go
   lipo -create -output hotkey-translator hotkey-translator-amd64 hotkey-translator-arm64
   ```

3. **Create a macOS app bundle**
   ```bash
   mkdir -p HotkeyTranslator.app/Contents/MacOS
   cp hotkey-translator HotkeyTranslator.app/Contents/MacOS/
   ```

### Development Guidelines

#### Code Style
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions focused and small

#### Error Handling
- Always handle errors explicitly
- Use descriptive error messages
- Log errors with context
- Graceful degradation when possible

#### Testing
- Test hotkey detection thoroughly
- Verify API integration
- Test with various text lengths and languages
- Test accessibility permissions

#### Performance
- Use goroutines for non-blocking operations
- Implement proper debouncing for hotkeys
- Cache API responses when appropriate
- Monitor memory usage

### Debugging

Enable debug logging by adding print statements or using a logging library:

```go
fmt.Printf("Debug: %s\n", debugInfo)
```

Common issues:
- **Hotkeys not working**: Check accessibility permissions
- **API errors**: Verify API key and network connection
- **Clipboard issues**: Ensure proper macOS permissions

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Roadmap

- [ ] Support for more languages
- [ ] Custom hotkey configuration
- [ ] Translation history
- [ ] Batch translation
- [ ] Windows/Linux support
- [ ] Plugin system

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/yourusername/assistant-keyboard/issues) page
2. Create a new issue with detailed description
3. Include your macOS version and Go version

## 🙏 Acknowledgments

- [Fyne](https://fyne.io/) for the GUI framework
- [gohook](https://github.com/robotn/gohook) for global hotkey detection
- [Google Gemini](https://ai.google.dev/) for AI translation capabilities

### Build Script

The project includes a convenient build script (`app.sh`) that creates a proper macOS application bundle:

```bash
# Make the script executable
chmod +x app.sh

# Build the application
./app.sh
```

This will create:
- `dist/translator` - Standalone binary
- `dist/translator.app` - macOS application bundle

The build script automatically:
- Creates a `dist/` directory for all build artifacts
- Builds the Go binary with optimizations
- Creates a proper macOS app bundle structure
- Generates app icons from `icon.png` (if present)
- Sets up proper Info.plist with Unicode support
- Copies configuration files to the app bundle
