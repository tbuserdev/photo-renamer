# Photo Renamer

A focused desktop app built in Go for automated photo renaming and organization. It previews every proposed change, extracts EXIF metadata to create standardized filenames, and keeps duplicates and metadata errors in dedicated review folders.

## 🚀 Features

- **EXIF-Based Renaming**: Automatically renames files using the original capture date, camera make, and model.
- **Smart Metadata Detection**: Detects if an image has been edited (e.g., via Lightroom) and includes that in the filename.
- **Content-Based Duplicate Handling**: Uses SHA-256 when filenames collide. Byte-identical files move to `DUPLICATES`; different files are preserved with `_2`, `_3`, and later suffixes.
- **Error Management**: Moves files with missing or corrupt metadata to an `ERROR-OUTPUT` folder for manual review.
- **Guided Desktop Workflow**: Choose a folder, review every action, then confirm once.
- **Flexible Review Table**: Resize the window and drag column dividers to fit long filenames.
- **Native Dialogs**: Uses the operating system's folder picker and confirmation dialogs.
- **Progress Tracking**: Real-time progress shows the status of your renaming task.
- **Multi-Format Support**: Supports a wide range of formats including JPG, PNG, GIF, BMP, TIFF, WebP, HEIF/HEIC, MOV, MP4, and various RAW formats (ARW, CR2, CR3, DNG, NEF, RW2, SR2, SRW).
- **Cross-Platform**: Built with [Fyne](https://fyne.io/) for macOS, Linux, and Windows. The original Bubble Tea interface remains available for terminal users.

## 📸 Renaming Logic

The tool generates filenames based on the following pattern:
- **Standard**: `YYYY-MM-DD_HH-MM-SS_Maker-Model.ext`
- **Edited**: `YYYY-MM-DD_HH-MM-SS_Maker-Model_Editor.ext` when Lightroom, Photoshop, or Photomator is explicitly detected

Missing camera metadata and unrecognized software never add `Unknown` or `Original` placeholders.

## 📦 Installation

Photo Renamer requires [ExifTool](https://exiftool.org/) to be installed. Follow the [ExifTool installation guide](https://exiftool.org/install.html) and verify the dependency with `exiftool -ver`. On macOS, the app also checks the standard Homebrew and MacPorts install locations because apps launched from Finder do not inherit your shell's `PATH`. See [the dependency and timestamp policy](docs/exiftool.md) for details.

### From Binary (Recommended)

1. Go to the [Releases](https://github.com/tbuserdev/photo-renamer/releases) page.
2. Download the archive for your operating system.
3. Extract the archive.

#### Linux users

You may need to make the binary executable:

```bash
chmod +x photo-renamer
```

### macOS installation

1. Install ExifTool with [Homebrew](https://brew.sh/):

   ```bash
   brew install exiftool
   exiftool -ver
   ```

   The app automatically checks the standard Homebrew locations on both Apple Silicon and Intel Macs, as well as the MacPorts location. You do not need to launch the app from Terminal.

2. Download the macOS ZIP for your processor from [Releases](https://github.com/tbuserdev/photo-renamer/releases):

   - Apple Silicon (`M1`, `M2`, `M3`, `M4`, or newer): `darwin-arm64`
   - Intel Mac: `darwin-amd64`

3. Extract the ZIP and drag `PhotoRenamer.app` into your `/Applications` folder.

4. Because current release builds are not notarized by Apple, macOS may block the first launch. Try to open the app once, then go to **System Settings → Privacy & Security**, scroll to **Security**, and click **Open Anyway** for Photo Renamer. Confirm by clicking **Open**. This grants an exception only for this app and keeps Gatekeeper enabled.

If **Open Anyway** is unavailable, you can remove the quarantine attribute from this specific app in Terminal:

```bash
xattr -dr com.apple.quarantine /Applications/PhotoRenamer.app
```

Only run that command after confirming that the app came from this project's GitHub Releases page. It bypasses Apple's quarantine check for Photo Renamer; do not disable Gatekeeper globally.

When selecting protected locations such as Desktop, Documents, Downloads, removable volumes, or network drives, macOS may ask for file-access permission. Choose **Allow**. You can review or revoke access later under **System Settings → Privacy & Security → Files & Folders**.

### From Source

```bash
git clone https://github.com/tbuserdev/photo-renamer.git
cd photo-renamer
go mod download
```

## 🖥 Usage

### Running the Desktop Application

**1. Using the Downloaded Binary:**
   - Open the packaged application or double-click `photo-renamer.exe` on Windows.
   - On **macOS/Linux**, an unpackaged binary can also be run from a terminal: `./photo-renamer`

**2. Running from Source (Developers):**
```bash
go run .
```

### Workflow

1. Choose a photo folder with the native system picker.
2. Review the proposed names and the destination of duplicates or errors.
3. Confirm the batch in the native system dialog.
4. Keep the app open while the progress indicator completes.

No files are changed during scanning or preview. During processing, errors move to `ERROR-OUTPUT` and exact duplicates move to `DUPLICATES` inside the selected folder.

### Terminal Interface

The original terminal interface remains available from source:

```bash
go run ./cmd/photo-renamer-tui
```

## 📜 Dependencies

- [Fyne](https://fyne.io/): Cross-platform desktop UI framework.
- [sqweek/dialog](https://github.com/sqweek/dialog): Native folder and confirmation dialogs.
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss): Optional terminal interface.
- [ExifTool](https://exiftool.org/): Group-aware metadata extraction for still images, RAW files, HEIC/HEIF, MOV, and MP4. It is an external runtime dependency and is not bundled in releases.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
