# Photo Renamer & Copy Tool

A powerful and simple GUI tool built in Go for automated photo renaming and organization. This tool extracts EXIF metadata from your images to create descriptive, standardized filenames and organizes them into your preferred directory structure.

## 🚀 Features

- **EXIF-Based Renaming**: Automatically renames files using the original capture date, camera make, and model.
- **Smart Metadata Detection**: Detects if an image has been edited (e.g., via Lightroom) and includes that in the filename.
- **Duplicate Handling**: Automatically detects and moves duplicate files to a dedicated `DUPLICATES` folder to prevent data loss or overwriting.
- **Error Management**: Moves files with missing or corrupt metadata to an `ERROR-OUTPUT` folder for manual review.
- **Progress Tracking**: Real-time progress bar shows the status of your renaming task.
- **Multi-Format Support**: Supports a wide range of formats including JPG, PNG, GIF, BMP, TIFF, WebP, HEIF/HEIC, and various RAW formats (ARW, CR2, CR3, DNG, NEF, RW2, SR2, SRW).
- **Cross-Platform**: Built with [Fyne](https://fyne.io/), making it compatible with macOS and Windows.

## 📸 Renaming Logic

The tool generates filenames based on the following pattern:
- **Standard**: `YYYY-MM-DD_HH-MM-SS_Maker-Model.ext`
- **Edited**: `YYYY-MM-DD_HH-MM-SS_Maker-Model_Software.ext` (e.g., including "Lightroom")

## 🛠 Prerequisites

- [Go](https://golang.org/doc/install) (version 1.19 or higher)
- C compiler (required for Fyne on some systems)
- System dependencies for Fyne (see [Fyne setup guide](https://developer.fyne.io/started/))

## 📦 Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/photo-renamer.git
   cd photo-renamer
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

## 🖥 Usage

1. Run the application:
   ```bash
   go run main.go
   ```

2. Using the GUI:
   - **Input Path**: Select the folder containing your source images (e.g., your camera's SD card).
   - **Output Path**: Select the destination folder where you want the renamed photos to be saved.
   - Click **Start** to begin the process.
   - Use **Open Output** to quickly view your organized photos once finished.

## 📂 Project Structure

- `main.go`: Entry point and UI definition using Fyne.
- `renamer/`: Core logic for file traversal and renaming.
- `renamer/utility/`: Metadata extraction (EXIF) and system utility functions.
- `progressCounter/`: Logic for calculating total file count for the progress bar.

## 📜 Dependencies

- [Fyne v2](https://fyne.io/): Cross-platform GUI toolkit.
- [goexif](https://github.com/rwcarlsen/goexif): EXIF metadata decoding.
- [gjson](https://github.com/tidwall/gjson): Fast JSON parsing for metadata handling.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
