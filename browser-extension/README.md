# YouTube URL Poster (ytmusic optional component)

This browser extension integrates with the
[ytmusic](https://github.com/conradpankoff/ytmusic) project, which maintains a
local cache of YouTube videos for offline use on devices without official
YouTube support. The plugin provides a convenient way to send YouTube video URLs
directly from your browser to a configured API endpoint managed by ytmusic.

---

## Features

- **Send YouTube URLs to ytmusic**:
  - A button appears below YouTube videos to send the current video to the API.
  - A popup menu allows sending the active video tab.
  - A keyboard shortcut (`Ctrl+Shift+Y` / `Command+Shift+Y` on macOS) instantly
    submits the current video.

- **Visual feedback**:
  - Notifications indicate success, failure, or errors.
  - Button text updates to show status (`✅ Already sent`, `Sending...`, etc.).

- **Duplicate prevention**:
  - Tracks previously sent URLs to avoid resubmission.

- **Configurable API endpoint**:
  - Options page allows setting the target API endpoint used by ytmusic.

---

## Installation

1. Clone or download this repository.
2. Open Chrome (or a Chromium-based browser) and navigate to
   `chrome://extensions/`.
3. Enable **Developer mode**.
4. Click **Load unpacked** and select the plugin folder.
5. The extension will now be available in your toolbar.

---

## Configuration

1. Click the extension icon in the toolbar.
2. Open **Options**.
3. Enter the API endpoint URL provided by your ytmusic instance.
4. Save the configuration.

---

## Usage

- **On YouTube video pages**:  
  Click the red **Send to API** button below the video to submit the URL.
  - If already sent, the button shows `✅ Already sent`.
  - While sending, a spinner and "Sending..." message are displayed.

- **From the popup**:  
  Use **Send Current Video** to submit the active tab’s YouTube URL.

- **Keyboard shortcut**:  
  Press `Ctrl+Shift+Y` (Windows/Linux) or `Command+Shift+Y` (macOS) to send the
  current video URL.

---

## Notifications

After sending a URL, a Chrome notification confirms the result. The button text
also updates temporarily to reflect the outcome.

---

## Permissions

The extension requires:

- `activeTab` – access the current YouTube tab.
- `scripting` – inject the content script.
- `storage` – save API endpoint and sent URLs.
- `notifications` – display feedback messages.

---

## Development Notes

- **background.js**: Handles API requests, notifications, and keyboard
  shortcuts.
- **content.js**: Manages the YouTube page button and UI feedback.
- **options.html/js**: Provides configuration for the API endpoint.
- **popup.html/js**: Offers quick access to sending and configuration.
- **manifest.json**: Defines permissions, scripts, and commands.

---

## Version

Current version: **1.2**

---

## License

MIT License.
