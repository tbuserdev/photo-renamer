# Testing Photo Renamer

This project uses standard Go testing practices. We have unit tests for the core logic and GUI presentation helpers, plus manual testing procedures for the desktop workflow.

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
- **renamer/metadata_test.go**: Tests metadata-driven filename generation.
- **renamer/exiftool_test.go**: Tests grouped JSON normalization, timestamp precedence and offsets, safe batch arguments, dependency/process errors, malformed and partial output, and cancellation with a fake ExifTool process.
- **renamer/renamer_test.go**: Tests the `ScanFiles` function by creating temporary directories with dummy files to ensure the file walker correctly identifies valid images and ignores excluded files.
- **gui/app_test.go**: Tests preview summaries and action presentation without opening a window.

## Manual Desktop Testing

The `_test` directory contains sample images that can be used for manual verification.

1. Copy representative photos into a disposable `_test` directory.
2. Run the application with `go run .`.
3. Choose `_test` in the native folder picker.
4. Verify that scanning can be cancelled without changing files.
5. Review the proposed names and the duplicate/error destinations.
6. Confirm the batch and keep the app open through completion.
7. Use **Show in File Manager** and verify renamed files plus the `DUPLICATES` and `ERROR-OUTPUT` folders.
8. Repeat once with the operating system in dark appearance.

## CI/CD

The GitHub Actions workflow `.github/workflows/build-and-release.yml` is configured to run `go test ./...` before building the application. This ensures that no broken code is released.
