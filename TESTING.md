# Testing Photo Renamer

This project uses standard Go testing practices. We have unit tests for the core logic and manual testing procedures for the TUI.

## Automated Tests

We use the built-in `testing` package.

### Running Tests
To run all tests in the project:
```bash
go test ./...
```

To run tests with verbose output:
```bash
go test -v ./...
```

### Test Coverage
- **renamer/metadata_test.go**: Tests the logic for extracting metadata and generating new filenames. It uses mock JSON strings to simulate Exif data.
- **renamer/renamer_test.go**: Tests the `ScanFiles` function by creating temporary directories with dummy files to ensure the file walker correctly identifies valid images and ignores excluded files.

## Manual Testing

The `_test` directory contains sample images that can be used for manual verification.

1. Run the application: `go run .`
2. In the TUI, navigate to the `_test` directory as the input folder.
3. Select an output folder (e.g., create a new `output` folder inside `_test`).
4. Press `Enter` to view the preview table.
5. Verify that the "New Name" column looks correct based on the file metadata.
6. Press `Enter` again to execute the rename.
7. Check the output folder to verify files were moved and renamed correctly.

## CI/CD

The GitHub Actions workflow `.github/workflows/build-and-release.yml` is configured to run `go test ./...` before building the application. This ensures that no broken code is released.
