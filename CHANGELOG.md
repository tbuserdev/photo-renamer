# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.4] - 2026-08-06

### Added
- **Metadata:** Added ExifTool-based, group-aware metadata extraction for supported photos, RAW files, HEIC/HEIF, MOV, and MP4 media.
- **Metadata:** Added capture-time selection for video metadata, including timezone-aware QuickTime timestamps.
- **Duplicates:** Added SHA-256 content checks for destination filename collisions.
- **Testing:** Added coverage for ExifTool normalization, batch execution, timestamp parsing, dependency failures, partial or malformed output, and cancellation.

### Changed
- **Metadata:** Metadata scanning now processes supported files in a single ExifTool batch and records the source metadata field used for each capture time.
- **Dependencies:** Replaced the bundled Go EXIF/JSON parsing dependencies with the external ExifTool runtime.

### Fixed
- **Metadata:** Preserve metadata failures for individual files rather than failing the entire scan when ExifTool returns usable per-file results.
- **Duplicates:** Preserve different files with `_2`, `_3`, and later suffixes instead of treating matching generated filenames as proof of duplication.
- **UI:** Show the underlying metadata error in the rename preview for files that cannot be processed.



## [v0.1.3] - 2026-01-19

### Fixed
- **Metadata:** Relaxed strict requirement for "Model" and "Make" metadata tags. Files with missing metadata now use "Unknown" fallback instead of failing with an error.
- **Metadata:** Added fallback to "Original" when software information is missing during rename.



## [v0.1.2] - 2026-01-19

### Added
- **UI:** Support for Flexoki color scheme (Dark and Light modes).
- **UI:** Dynamic theme toggling with 'T' key.
- **UI:** Auto-detect system/terminal theme at startup.

### Changed
- **UI:** Refactored styling system to support dynamic themes across all components.



## [v0.1.1] - 2025-12-25

### Added
- **UI:** Added a loading spinner to the preview screen to indicate progress while scanning files.
- **UI:** Added display of original file count in the preview table header.

### Changed
- **Metadata:** Improved software extraction logic to better support Lightroom, Photoshop, and Photomator.
- **Metadata:** Enhanced datetime extraction with fallback strategy (DateTimeOriginal → DateTimeDigitized → DateTime) for better camera compatibility.



## [v0.1.0] - 2025-12-25

### Added
- **CLI:** Added a `--version` flag to display application version.
- **Build:** Updated build process to inject versioning information.

### Fixed
- **CI/CD:** Updated binary naming format to use `github.ref_name` for accurate version tracking.



## [v0.0.2] - 2025-12-25

### Added
- **Features:** Added support for skipping file actions during the renaming process.
- **Metadata:** Enhanced image metadata handling with EXIF data display and a new debug view.
- **Testing:** Added unit tests for metadata extraction and file scanning functionality.
- **CI/CD:** Added a dedicated testing job to the GitHub Actions workflow.

### Changed
- **UI:** Improved table styling with padding and borders for better readability.
- **Documentation:** Updated README to simplify installation instructions and improve formatting.

### Fixed
- **UI:** Removed background color from table header style for improved visibility.



## [v0.0.1] - 2025-12-25

### Added
- **UI:** Migrated to a TUI (Text User Interface) using Bubble Tea.
- **UI:** Added theming support with Lipgloss.
- **Features:** Implemented file picker for input folder selection.
- **Features:** Added preview functionality for renaming images.
- **CI/CD:** Added GitHub Actions workflow for build and release automation.
- **Project:** Initial upload of the `photo-renamer` tool.

### Changed
- **Refactor:** Renamed module from `ImageRenamer` to `photo-renamer`.
- **Refactor:** Moved utility package directly into the renamer package.
- **Documentation:** Updated README to reflect the transition from GUI to TUI.
- **Style:** Updated color scheme and title.
