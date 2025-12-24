# Photo Renamer & Copy Tool

A powerful and simple TUI (Terminal User Interface) tool built in Go for automated photo renaming and organization. This tool extracts EXIF metadata from your images to create descriptive, standardized filenames and organizes them into your preferred directory structure.

## 🚀 Features

- **EXIF-Based Renaming**: Automatically renames files using the original capture date, camera make, and model.
- **Smart Metadata Detection**: Detects if an image has been edited (e.g., via Lightroom) and includes that in the filename.
- **Duplicate Handling**: Automatically detects and moves duplicate files to a dedicated `DUPLICATES` folder to prevent data loss or overwriting.
- **Error Management**: Moves files with missing or corrupt metadata to an `ERROR-OUTPUT` folder for manual review.
- **Progress Tracking**: Real-time progress bar shows the status of your renaming task.
- **Multi-Format Support**: Supports a wide range of formats including JPG, PNG, GIF, BMP, TIFF, WebP, HEIF/HEIC, and various RAW formats (ARW, CR2, CR3, DNG, NEF, RW2, SR2, SRW).
- **Cross-Platform**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), ensuring a beautiful terminal experience on macOS, Linux, and Windows.

## 📸 Renaming Logic

The tool generates filenames based on the following pattern:
- **Standard**: `YYYY-MM-DD_HH-MM-SS_Maker-Model.ext`
- **Edited**: `YYYY-MM-DD_HH-MM-SS_Maker-Model_Software.ext` (e.g., including "Lightroom")

## 🛠 Prerequisites

- [Go](https://go.dev/doc/install) (version 1.25 or higher)

## 📦 Installation

### From Binary (Recommended)

1. Go to the [Releases](https://github.com/yourusername/photo-renamer/releases) page.
2. Download the archive for your operating system.
3. Extract the archive.

**macOS / Linux users:**
You may need to make the binary executable:
```bash
chmod +x photo-renamer
```

**macOS Note:**
If you receive a "Developer cannot be verified" error:
1. Open System Settings > Privacy & Security.
2. Scroll down to the Security section and click "Open Anyway".
Alternatively, run this command on the downloaded binary:
```bash
xattr -d com.apple.quarantine photo-renamer
```

### From Source

   ```bash
   git clone https://github.com/yourusername/photo-renamer.git
   cd photo-renamer
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## 🖥 Usage

### Running the Application

**1. Using the Downloaded Binary:**
   - Double-click the `photo-renamer` executable.
   - On **macOS/Linux**, you can also run it from the terminal: `./photo-renamer`

**2. Running from Source (Developers):**
   ```bash
   go run .
   ```

### Controls & Workflow
   - **Select Input Folder**: Navigate through directories using the **Arrow Keys**. Press **Enter** to select the currently highlighted directory as your source folder.
   - **Review Preview**: A table will appear showing the proposed filename changes and identifying any duplicates or errors.
   - **Confirm Rename**: Press **Enter** to confirm and start the renaming process.
   - **Exit**: Press **Esc** or **Ctrl+C** to quit the application at any time.

## 📂 Project Structure

- `main.go`: Entry point and UI definition using Fyne.
- `renamer/`: Core logic for file traversal and renaming.
- `renamer/metadata.go`: Metadata extraction (EXIF) and system utility functions.

## 📜 Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea): A powerful, functional TUI framework.
- [Lip Gloss](https://github.com/charmbracelet/lipgloss): Style definitions for nice terminal layouts.
- [goexif](https://github.com/rwcarlsen/goexif): EXIF metadata decoding.
- [gjson](https://github.com/tidwall/gjson): Fast JSON parsing for metadata handling.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
