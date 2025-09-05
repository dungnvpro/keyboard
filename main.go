package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/joho/godotenv"
	hook "github.com/robotn/gohook"
)

// Config struct for storing API key and model
type Config struct {
	GeminiAPIKey      string   `json:"gemini_api_key"`
	Model             string   `json:"model"`
	SelectedLanguages []string `json:"selected_languages"`
	IncludePrefix     bool     `json:"include_prefix"`
}

// Global config variable
var appConfig Config

// Global variable for selected output languages
var selectedLanguages []string

// Available Gemini models
var geminiModels = []string{
	"gemini-2.0-flash-lite",
	"gemini-1.5-flash",
	"gemini-1.5-pro",
	"gemini-1.0-pro",
}

// Function to update selected languages
func updateSelectedLanguages() {
	fmt.Printf("🌐 Selected output languages: %v\n", selectedLanguages)
	// Update appConfig and save
	appConfig.SelectedLanguages = selectedLanguages
	if err := saveConfig(appConfig); err != nil {
		fmt.Printf("❌ Error saving selected languages: %v\n", err)
	} else {
		fmt.Printf("✅ Selected languages saved to config\n")
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to remove item from slice
func removeFromSlice(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// Get config file path (same directory as executable)
func getConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("⚠️ Cannot get executable path, using current directory: %v\n", err)
		return "config.json"
	}
	execDir := filepath.Dir(execPath)
	configPath := filepath.Join(execDir, "config.json")
	fmt.Printf("🔍 DEBUG: Executable path: %s\n", execPath)
	fmt.Printf("🔍 DEBUG: Executable directory: %s\n", execDir)
	fmt.Printf("🔍 DEBUG: Config path: %s\n", configPath)
	return configPath
}

// Load config from file
func loadConfig() Config {
	configPath := getConfigPath()

	// Try to load from config.json first
	if data, err := os.ReadFile(configPath); err == nil {
		var config Config
		if json.Unmarshal(data, &config) == nil {
			// Set default model if not specified
			if config.Model == "" {
				config.Model = "gemini-2.0-flash-lite"
			}
			// Set default selected languages if not specified
			if config.SelectedLanguages == nil {
				config.SelectedLanguages = []string{"EN"} // Default to English only
			}
			// Set default include prefix if not specified
			// includePrefix defaults to false (zero value)
			fmt.Printf("✅ Loaded config from: %s\n", configPath)
			fmt.Printf("🌐 Loaded selected languages: %v\n", config.SelectedLanguages)
			fmt.Printf("️ Include prefix: %v\n", config.IncludePrefix)
			return config
		}
	}

	// Fallback to .env file (for development)
	err := godotenv.Load()
	if err == nil {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey != "" {
			fmt.Println("✅ Loaded API key from .env file")
			return Config{
				GeminiAPIKey:      apiKey,
				Model:             "gemini-2.0-flash-lite",
				SelectedLanguages: []string{"EN"}, // Default to English only
				IncludePrefix:     false,          // Default to false
			}
		}
	}

	// Fallback to environment variable
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey != "" {
		fmt.Println("✅ Loaded API key from environment variable")
		return Config{
			GeminiAPIKey:      apiKey,
			Model:             "gemini-2.0-flash-lite",
			SelectedLanguages: []string{"EN"}, // Default to English only
			IncludePrefix:     false,          // Default to false
		}
	}

	fmt.Println("⚠️  No API key found in config.json, .env, or environment variables")
	return Config{
		GeminiAPIKey:      "",
		Model:             "gemini-2.0-flash-lite",
		SelectedLanguages: []string{"EN"}, // Default to English only
		IncludePrefix:     false,          // Default to false
	}
}

// Auto-start hotkey listener if API key is available
func autoStartIfReady() bool {
	if appConfig.GeminiAPIKey != "" && appConfig.GeminiAPIKey != "YOUR_GEMINI_API_KEY_HERE" {
		fmt.Println("🚀 Auto-starting hotkey listener (API key found)")
		go startHotkeyListener()
		return true
	}
	return false
}

// Save config to file
func saveConfig(config Config) error {
	configPath := getConfigPath()
	fmt.Printf("🔍 DEBUG: Saving config to: %s\n", configPath)
	fmt.Printf(" DEBUG: Config data: %+v\n", config)

	// Check if directory exists, create if not
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("❌ Error creating directory %s: %v\n", dir, err)
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling config: %v\n", err)
		return err
	}

	fmt.Printf("🔍 DEBUG: JSON data: %s\n", string(data))

	// Try to write file
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing config file: %v\n", err)
		fmt.Printf("🔍 DEBUG: Current working directory: %s\n", func() string {
			wd, _ := os.Getwd()
			return wd
		}())
		return err
	}

	// Verify file was written
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("❌ File was not created: %s\n", configPath)
		return fmt.Errorf("file was not created")
	}

	fmt.Printf("✅ Config saved successfully to: %s\n", configPath)
	return nil
}

// Get Gemini API key from config (load from file each time)
func getGeminiAPIKey() string {
	// Load config from file each time instead of using global variable
	config := loadConfig()
	if config.GeminiAPIKey == "" {
		return "YOUR_GEMINI_API_KEY_HERE"
	}
	return config.GeminiAPIKey
}

// Get Gemini model from config (load from file each time)
func getGeminiModel() string {
	// Load config from file each time instead of using global variable
	config := loadConfig()
	if config.Model == "" {
		return "gemini-2.0-flash-lite"
	}
	return config.Model
}

// Request struct for Gemini API
type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

// Response struct for Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// Global channels for translation requests
var translationChan = make(chan bool, 1)
var dualTranslationChan = make(chan bool, 1)

type smallTheme struct {
	fyne.Theme
}
type defaultTheme struct {
	fyne.Theme
}

func (s smallTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 10 // giảm fontsize mặc định
	}
	return s.Theme.Size(name)
}
func (s defaultTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 12 // giảm fontsize mặc định
	}
	return s.Theme.Size(name)
}

// Check if app has accessibility permissions
func checkAccessibilityPermission() bool {
	if runtime.GOOS != "darwin" {
		return true // Skip check on non-macOS systems
	}

	// Use AppleScript to check accessibility permissions
	cmd := exec.Command("osascript", "-e", `
		tell application "System Events"
			try
				-- Try to perform an action that requires accessibility permission
				-- Use a safe key combination that won't trigger system shortcuts
				keystroke "a" using {command down, option down}
				return "true"
			on error
				return "false"
			end try
		end tell
	`)

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Error checking accessibility permission: %v\n", err)
		return false
	}

	result := strings.TrimSpace(string(output))
	return result == "true"
}

func main() {
	// Load config at startup
	appConfig = loadConfig()

	// Initialize selectedLanguages from config
	selectedLanguages = appConfig.SelectedLanguages
	fmt.Printf("🌐 Initialized selected languages: %v\n", selectedLanguages)

	myApp := app.New()
	myApp.Settings().SetTheme(&smallTheme{theme.DefaultTheme()})
	myWindow := myApp.NewWindow("Hotkey Translator")
	myWindow.SetIcon(theme.ComputerIcon())

	// Check accessibility permission first
	if !checkAccessibilityPermission() {
		myApp.Settings().SetTheme(&defaultTheme{theme.DefaultTheme()})
		fmt.Println("⚠️ Accessibility permission not granted")
		// Show instruction dialog with detailed guidance
		instructionText := `🔒 Accessibility Permission Required

This app needs Accessibility permission to work properly.

How to grant permission:
1. Click "Open System Preferences" below
2. In the Accessibility window, click the lock icon 🔒
3. Enter your password to make changes
4. Click the "+" button to add this app
5. Navigate to and select this app from Applications
6. Make sure the checkbox next to this app is checked ✅
7. Click "Quit" below and restart this app

The app will appear in the list once you add it.`

		content := container.NewVBox(
			widget.NewLabel(instructionText),
			widget.NewButton("Open System Preferences", func() {
				exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Start()
			}),
			widget.NewButton("Quit", func() {
				os.Exit(0)
			}),
		)

		myWindow.SetContent(content)
		myWindow.Resize(fyne.NewSize(300, 200))
		myWindow.CenterOnScreen()
		myWindow.SetFixedSize(true)
		myWindow.ShowAndRun()
		return
	} else {
		fmt.Println("✅ Accessibility permission granted")
	}

	// Auto-start if API key is available
	autoStarted := autoStartIfReady()

	// Create title section
	// titleLabel := widget.NewLabel("🌐 Hotkey Translator")
	// titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	// titleLabel.Alignment = fyne.TextAlignCenter

	subtitleLabel := widget.NewLabel("AI-powered translation with global hotkeys")
	subtitleLabel.Alignment = fyne.TextAlignCenter

	// Create API key section
	apiKeyLabel := widget.NewLabel("🔑 Gemini API Key")
	apiKeyLabel.TextStyle = fyne.TextStyle{Bold: true}

	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.SetText(appConfig.GeminiAPIKey)
	apiKeyEntry.SetPlaceHolder("Enter your Gemini API key...")
	// apiKeyEntry.Password = true // Hide API key for security

	// Auto-save when API key changes
	apiKeyEntry.OnChanged = func(text string) {
		fmt.Printf("🔍 DEBUG: API key changed to: %s\n", text)
		appConfig.GeminiAPIKey = text
		if err := saveConfig(appConfig); err != nil {
			fmt.Printf("❌ Error auto-saving config: %v\n", err)
		} else {
			fmt.Printf("✅ API key auto-saved successfully\n")
		}
	}

	// Auto-save when API key field is submitted (Enter key pressed)
	apiKeyEntry.OnSubmitted = func(text string) {
		fmt.Printf("🔍 DEBUG: API key submitted: %s\n", text)
		appConfig.GeminiAPIKey = text
		if err := saveConfig(appConfig); err != nil {
			fmt.Printf("❌ Error auto-saving config on submit: %v\n", err)
		} else {
			fmt.Printf("✅ API key auto-saved on submit\n")
		}
	}

	// Create model selection section
	modelLabel := widget.NewLabel("🤖 AI Model")
	modelLabel.TextStyle = fyne.TextStyle{Bold: true}

	modelSelect := widget.NewSelect(geminiModels, func(value string) {
		appConfig.Model = value
		// Auto-save when model changes
		if err := saveConfig(appConfig); err != nil {
			fmt.Printf("❌ Error auto-saving config: %v\n", err)
		} else {
			fmt.Printf("✅ Model auto-saved\n")
		}
	})
	modelSelect.SetSelected(appConfig.Model)

	// Remove the OnFocusChanged for modelSelect since it doesn't exist
	// The OnChanged callback in NewSelect is sufficient for auto-saving

	// Create start button (removed save button)
	var startButton *widget.Button
	startButton = widget.NewButton("🚀 Start Hotkey Listener", func() {
		// Update config before starting
		appConfig.GeminiAPIKey = apiKeyEntry.Text
		appConfig.Model = modelSelect.Selected

		// Validate API key
		if appConfig.GeminiAPIKey == "" || appConfig.GeminiAPIKey == "YOUR_GEMINI_API_KEY_HERE" {
			dialog.ShowInformation("⚠️ Configuration Required", "Please enter a valid Gemini API key before starting!", myWindow)
			return
		}

		// Save config
		if err := saveConfig(appConfig); err != nil {
			fmt.Printf("❌ Error saving config: %v\n", err)
		}

		go startHotkeyListener()
		startButton.Disable() // Disable button after starting
		// Keep apiKeyEntry and modelSelect enabled for editing

		// Update button text to show status
		startButton.SetText("✅ Hotkey Listener Active")
	})
	startButton.Importance = widget.HighImportance

	// If auto-started, update UI to reflect the status
	if autoStarted {
		startButton.Disable() // Disable button after starting
		// Keep apiKeyEntry and modelSelect enabled for editing
		startButton.SetText("✅ Hotkey Listener Active")

	}

	// Create hotkey instructions section
	instructionsLabel := widget.NewLabel("🎯 Hotkey Instructions")
	instructionsLabel.TextStyle = fyne.TextStyle{Bold: true}
	// instructionsLabel.Alignment = fyne.TextLe

	hotkeyHLabel := widget.NewLabel("⌨️  Control + Option + H: Translate selected text to English only \n⌨️  Control + Option + J: Select all text and translate to English + Japanese")

	// Create warning section
	warningLabel := widget.NewLabel("⚠️  Important")
	warningLabel.TextStyle = fyne.TextStyle{Bold: true}
	// warningLabel.Alignment = fyne.TextAlignCenter

	warningText := widget.NewLabel("Make sure to grant Accessibility permissions in System Preferences > Security & Privacy > Privacy > Accessibility")
	warningText.Wrapping = fyne.TextWrapWord
	warningText.Alignment = fyne.TextAlignCenter
	// create an other button to open System Preferences
	warningTextButton := widget.NewButton("Open System Preferences", func() {
		exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Start()
	})
	warningTextButton.Importance = widget.LowImportance
	// warningTextButton.Alignment = fyne.TextAlignCenter

	// Create main content layout
	configSection := container.NewVBox(
		apiKeyLabel,
		apiKeyEntry,
		// widget.NewLabel(""), // Spacer
		modelLabel,
		modelSelect,
	)

	// tạo 1 selection gồm có 3 checkbox [EN] [VN] [JP]
	outputLanguageLabel := widget.NewLabel("Output Language:")
	outputLanguageLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create individual checkboxes for each language
	enCheck := widget.NewCheck("EN", func(value bool) {
		if value {
			// Add to selected languages if not already present
			if !contains(selectedLanguages, "EN") {
				selectedLanguages = append(selectedLanguages, "EN")
			}
		} else {
			// Remove from selected languages
			selectedLanguages = removeFromSlice(selectedLanguages, "EN")
		}
		updateSelectedLanguages()
	})
	enCheck.SetChecked(contains(appConfig.SelectedLanguages, "EN"))

	vnCheck := widget.NewCheck("VN", func(value bool) {
		if value {
			if !contains(selectedLanguages, "VN") {
				selectedLanguages = append(selectedLanguages, "VN")
			}
		} else {
			selectedLanguages = removeFromSlice(selectedLanguages, "VN")
		}
		updateSelectedLanguages()
	})
	vnCheck.SetChecked(contains(appConfig.SelectedLanguages, "VN"))

	jpCheck := widget.NewCheck("JP", func(value bool) {
		if value {
			if !contains(selectedLanguages, "JP") {
				selectedLanguages = append(selectedLanguages, "JP")
			}
		} else {
			selectedLanguages = removeFromSlice(selectedLanguages, "JP")
		}
		updateSelectedLanguages()
	})
	jpCheck.SetChecked(contains(appConfig.SelectedLanguages, "JP"))

	languageSelection := container.NewHBox(
		outputLanguageLabel,
		container.NewHBox(enCheck, vnCheck, jpCheck),
	)
	// Create prefix checkbox
	prefixCheck := widget.NewCheck("Include [LANG] prefix, ex: [EN]: text_text", func(value bool) {
		appConfig.IncludePrefix = value
		if err := saveConfig(appConfig); err != nil {
			fmt.Printf("❌ Error saving prefix setting: %v\n", err)
		} else {
			fmt.Printf("✅ Prefix setting saved: %v\n", value)
		}
	})
	prefixCheck.SetChecked(appConfig.IncludePrefix)
	prefixIncludeLabel := widget.NewLabel("Prefix: ")
	prefixIncludeLabel.TextStyle = fyne.TextStyle{Bold: true}
	includePrefixSection := container.NewHBox(
		prefixIncludeLabel,
		prefixCheck,
	)

	buttonSection := container.NewVBox(
		widget.NewLabel(""), // Spacer
		startButton,
		widget.NewLabel(""), // Spacer
	)
	// set width 100% for buttonSection

	hotkeySection := container.NewVBox(
		instructionsLabel,
		// widget.NewLabel(""), // Spacer
		hotkeyHLabel,
		// hotkeyHDesc,
		// widget.NewLabel(""), // Spacer
		// hotkeyJLabel,
	)

	warningSection := container.NewVBox(
		warningLabel,
		warningText,
		warningTextButton,
	)

	// Main content with proper spacing
	content := container.NewVBox(
		// Header
		// titleLabel,
		// subtitleLabel,
		// widget.NewSeparator(),

		// Configuration section
		configSection,
		// widget.NewSeparator(),

		// Buttons
		buttonSection,
		// widget.NewSeparator(),
		languageSelection,
		includePrefixSection,
		widget.NewSeparator(),

		// Hotkey instructions
		hotkeySection,
		widget.NewSeparator(),

		// Warning
		warningSection,
	)

	// Set window properties
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(300, 250))
	myWindow.CenterOnScreen()
	myWindow.SetFixedSize(true) // Prevent resizing for consistent layout

	// Start translation handler
	go handleTranslationRequests()

	myWindow.ShowAndRun()
}

// Handle translation requests
func handleTranslationRequests() {
	for {
		select {
		case <-translationChan:
			fmt.Println("🎯 Translation request received (English only)")
			// Use a goroutine with proper error handling
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("❌ Panic in translation: %v\n", r)
					}
				}()
				performTranslation()
			}()
		case <-dualTranslationChan:
			fmt.Println("🎯 Dual translation request received (English + Japanese)")
			// Use a goroutine with proper error handling
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("❌ Panic in dual translation: %v\n", r)
					}
				}()
				performDualTranslation()
			}()
		}
	}
}

// startHotkeyListener bắt các sự kiện hotkey Control+Option+H/J
func startHotkeyListener() {
	fmt.Println("Hotkey listener started.")
	fmt.Println("Nhấn Control+Option+H để dịch sang tiếng Anh.")
	fmt.Println("Nhấn Control+Option+J để dịch sang cả tiếng Anh và Nhật.")
	fmt.Printf("Sử dụng model: %s\n", getGeminiModel())
	fmt.Println("Đang lắng nghe sự kiện hotkey...")

	// Bắt đầu hook
	evChan := hook.Start()
	if evChan == nil {
		fmt.Println("Lỗi: Không thể khởi động hook (nil channel)")
		return
	}
	defer hook.End()

	// Thời gian debouncing
	var lastEvent time.Time

	for ev := range evChan {
		// Log sự kiện để debug (có thể xóa sau khi xác nhận hoạt động)
		// fmt.Printf("Sự kiện: Kind=%v, Keycode=%d (0x%x), Mask=%d (0x%x), Keychar=%q\n",
		// 	ev.Kind, ev.Keycode, ev.Keycode, ev.Mask, ev.Mask, ev.Keychar)

		// Chỉ xử lý sự kiện KeyDown
		if ev.Kind == hook.KeyDown {
			// Mask cho Control + Option (từ log)
			const requiredModifiers = 0xa00a // 40970

			// Kiểm tra nếu đúng tổ hợp Control + Option
			if ev.Mask == requiredModifiers {
				// Debouncing
				if time.Since(lastEvent) < 200*time.Millisecond {
					continue
				}
				lastEvent = time.Now()

				switch ev.Keycode {
				case 0x23: // Keycode cho 'H' (từ log)
					fmt.Printf("🎯 Phát hiện hotkey: Control+Option+H (chỉ tiếng Anh)\n")
					fmt.Printf("   Keycode: %d (0x%x), Mask: %d (0x%x)\n", ev.Keycode, ev.Keycode, ev.Mask, ev.Mask)
					select {
					case translationChan <- true:
						fmt.Println("Yêu cầu dịch tiếng Anh đã gửi")
					default:
						fmt.Println("⚠️ Yêu cầu dịch tiếng Anh bị bỏ qua (channel đầy)")
					}

				case 0x24: // Keycode cho 'J' (từ log)
					fmt.Printf("🎯 Phát hiện hotkey: Control+Option+J (tiếng Anh + Nhật)\n")
					fmt.Printf("   Keycode: %d (0x%x), Mask: %d (0x%x)\n", ev.Keycode, ev.Keycode, ev.Mask, ev.Mask)
					fmt.Printf("   Keycode: %d (0x%x), Mask: %d (0x%x)\n", ev.Keycode, ev.Keycode, ev.Mask, ev.Mask)
					select {
					case dualTranslationChan <- true:
						fmt.Println("Yêu cầu dịch tiếng Anh + Nhật đã gửi")
					default:
						fmt.Println("⚠️ Yêu cầu dịch tiếng Anh + Nhật bị bỏ qua (channel đầy)")
					}
				}
			}
		}
	}
}

// Safe version using system commands instead of robotgo
func performTranslation() {
	fmt.Println("�� Copying selected text...")

	// Add a small delay to ensure hotkey processing is complete
	time.Sleep(150 * time.Millisecond)

	// Copy selected text using system command
	var copyCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// On macOS, use osascript to simulate Cmd+C
		copyCmd = exec.Command("osascript", "-e", "tell application \"System Events\" to keystroke \"c\" using command down")
	} else {
		// On other systems, you might need different commands
		fmt.Println("❌ Unsupported operating system for copy command")
		return
	}

	err := copyCmd.Run()
	if err != nil {
		fmt.Printf("❌ Error copying text: %v\n", err)
		return
	}

	time.Sleep(300 * time.Millisecond) // Wait for copy to complete

	// Read from clipboard using system command
	var clipboardCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		clipboardCmd = exec.Command("pbpaste")
	} else {
		fmt.Println("❌ Unsupported operating system for clipboard access")
		return
	}

	clipboardOutput, err := clipboardCmd.Output()
	if err != nil {
		fmt.Printf("❌ Error reading clipboard: %v\n", err)
		return
	}

	text := string(clipboardOutput)
	if text == "" {
		fmt.Println("⚠️  No text in clipboard")
		return
	}

	fmt.Printf("📝 Copied text: \"%s\"\n", text)
	fmt.Printf("📏 Text length: %d characters\n", len(text))

	// Translate using Gemini API
	fmt.Println("🌐 Translating with Gemini API...")
	translatedText, err := translateWithGemini(text, "English")
	if err != nil {
		fmt.Printf("❌ Translation error: %v\n", err)
		return
	}

	fmt.Printf("✅ Translated text: \"%s\"\n", translatedText)

	// Write back to clipboard and paste
	fmt.Println("📋 Writing translated text to clipboard...")
	var writeCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		writeCmd = exec.Command("pbcopy")
	} else {
		fmt.Println("❌ Unsupported operating system for clipboard write")
		return
	}

	writeCmd.Stdin = bytes.NewReader([]byte(translatedText))
	err = writeCmd.Run()
	if err != nil {
		fmt.Printf("❌ Error writing to clipboard: %v\n", err)
		return
	}

	// Paste the translated text
	fmt.Println("📝 Pasting translated text...")
	var pasteCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		pasteCmd = exec.Command("osascript", "-e", "tell application \"System Events\" to keystroke \"v\" using command down")
	} else {
		fmt.Println("❌ Unsupported operating system for paste command")
		return
	}

	err = pasteCmd.Run()
	if err != nil {
		fmt.Printf("❌ Error pasting text: %v\n", err)
		return
	}

	fmt.Println("✨ Translation completed!")
}

// Dual translation function for English + Japanese with Select All
func performDualTranslation() {
	fmt.Println("📋 Selecting all text and copying...")

	// Add a small delay to ensure hotkey processing is complete
	time.Sleep(150 * time.Millisecond)

	// First, select all text using Cmd+A
	var selectAllCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// On macOS, use osascript to simulate Cmd+A
		selectAllCmd = exec.Command("osascript", "-e", "tell application \"System Events\" to keystroke \"a\" using command down")
	} else {
		// On other systems, you might need different commands
		fmt.Println("❌ Unsupported operating system for select all command")
		return
	}

	err := selectAllCmd.Run()
	if err != nil {
		fmt.Printf("❌ Error selecting all text: %v\n", err)
		return
	}

	time.Sleep(200 * time.Millisecond) // Wait for select all to complete

	// Then copy selected text using system command
	var copyCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// On macOS, use osascript to simulate Cmd+C
		copyCmd = exec.Command("osascript", "-e", "tell application \"System Events\" to keystroke \"c\" using command down")
	} else {
		// On other systems, you might need different commands
		fmt.Println("❌ Unsupported operating system for copy command")
		return
	}

	err = copyCmd.Run()
	if err != nil {
		fmt.Printf("❌ Error copying text: %v\n", err)
		return
	}

	time.Sleep(300 * time.Millisecond) // Wait for copy to complete

	// Read from clipboard using system command
	var clipboardCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		clipboardCmd = exec.Command("pbpaste")
	} else {
		fmt.Println("❌ Unsupported operating system for clipboard access")
		return
	}

	clipboardOutput, err := clipboardCmd.Output()
	if err != nil {
		fmt.Printf("❌ Error reading clipboard: %v\n", err)
		return
	}

	text := string(clipboardOutput)
	if text == "" {
		fmt.Println("⚠️  No text in clipboard")
		return
	}

	fmt.Printf("📝 Copied text: \"%s\"\n", text)
	fmt.Printf("📏 Text length: %d characters\n", len(text))

	// lay danh sách languages từ appConfig
	selectedLanguages = appConfig.SelectedLanguages

	// Map language codes to full names
	languageMap := map[string]string{
		"EN": "English",
		"VN": "Vietnamese",
		"JP": "Japanese",
	}

	var translations []string
	var combinedText string

	// Translate to each selected language
	for _, langCode := range selectedLanguages {
		if fullName, exists := languageMap[langCode]; exists {
			fmt.Printf("🌐 Translating to %s...\n", fullName)
			translatedText, err := translateWithGemini(text, fullName)
			if err != nil {
				fmt.Printf("❌ %s translation error: %v\n", fullName, err)
				continue // Skip this language if translation fails
			}

			// Format with or without prefix based on setting
			var formattedText string
			if appConfig.IncludePrefix {
				formattedText = fmt.Sprintf("[%s]: %s", langCode, translatedText)
			} else {
				formattedText = translatedText
			}

			translations = append(translations, formattedText)
			fmt.Printf("✅ %s: \"%s\"\n", langCode, translatedText)
		}
	}

	// Combine all translations
	if len(translations) > 0 {
		combinedText = strings.Join(translations, "\n----------------\n")
	} else {
		fmt.Println("⚠️ No valid languages selected for translation")
		return
	}

	// Write back to clipboard and paste
	fmt.Println("📋 Writing combined translations to clipboard...")
	var writeCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		writeCmd = exec.Command("pbcopy")
	} else {
		fmt.Println("❌ Unsupported operating system for clipboard write")
		return
	}

	writeCmd.Stdin = bytes.NewReader([]byte(combinedText))
	err = writeCmd.Run()
	if err != nil {
		fmt.Printf("❌ Error writing to clipboard: %v\n", err)
		return
	}

	// Paste the translated text
	fmt.Println("📝 Pasting combined translations...")
	var pasteCmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		pasteCmd = exec.Command("osascript", "-e", "tell application \"System Events\" to keystroke \"v\" using command down")
	} else {
		fmt.Println("❌ Unsupported operating system for paste command")
		return
	}

	err = pasteCmd.Run()
	if err != nil {
		fmt.Printf("❌ Error pasting text: %v\n", err)
		return
	}

	fmt.Println("✨ Dual translation completed!")
}

func translateWithGemini(text string, language string) (string, error) {
	// Get API key and model from config
	apiKey := getGeminiAPIKey()
	model := getGeminiModel()
	geminiAPIURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)

	// Prepare prompt for translation with improvement
	prompt := fmt.Sprintf("Please translate the following text to %s and improve/rephrase it to make it more clear, natural, and easy to understand. Return only the improved translated result without any additional explanation: \"%s\"", language, text)

	reqBody := GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", geminiAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var geminiResp GeminiResponse
	err = json.Unmarshal(body, &geminiResp)
	if err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no translation received")
	}

	// Clean up the response text
	result := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)

	// Remove any potential quotes around the result
	if strings.HasPrefix(result, "\"") && strings.HasSuffix(result, "\"") {
		result = result[1 : len(result)-1]
	}

	return result, nil
}
