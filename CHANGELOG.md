# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.1] - 2025-12-25

### Added
- **UI:** Added a loading spinner to the preview screen to indicate progress while scanning files.

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