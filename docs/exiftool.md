# ExifTool dependency

Photo Renamer requires [ExifTool](https://exiftool.org/) on `PATH`. Releases do not bundle ExifTool. This keeps the application archives small and makes the separately installed dependency and its updates explicit.

Install ExifTool using the [official installation instructions](https://exiftool.org/install.html), then verify it before starting Photo Renamer:

```text
exiftool -ver
```

If the executable is absent or cannot run, Photo Renamer stops the preview and reports how to install it. A file that ExifTool can read but that has no supported capture timestamp is reported as a per-file metadata error; filesystem timestamps are never substituted.

## Supported media and timestamps

Photo Renamer scans its existing still-image and RAW formats, HEIC/HEIF, and the initial video set MOV and MP4. It sends a preview batch to one ExifTool process and requests grouped JSON metadata.

Still timestamps use this precedence:

1. EXIF `DateTimeOriginal`
2. XMP `DateTimeOriginal`
3. QuickTime or Keys `CreationDate`
4. EXIF or XMP `CreateDate`
5. QuickTime `CreateDate`

Video timestamps use this separate precedence:

1. timezone-bearing QuickTime or Keys `CreationDate`
2. QuickTime `CreateDate`
3. QuickTime `MediaCreateDate`
4. QuickTime `TrackCreateDate`

Explicit offsets are preserved in the normalized metadata and the original wall-clock capture time is used in the filename. Photo Renamer invokes ExifTool with `QuickTimeUTC=1`; timezone-less QuickTime timestamps are therefore interpreted as UTC. Some cameras incorrectly store local time in timezone-less QuickTime fields, and that ambiguity cannot be recovered automatically.

## Packaging and licensing

ExifTool is not copied into Photo Renamer release archives. Users install and update it separately. ExifTool is distributed under the same terms as Perl itself (Artistic License or GPL); see the [upstream license notice](https://github.com/exiftool/exiftool#copyright-and-license). If a future release bundles ExifTool, it must pin a version and include the applicable distribution and license notices.
