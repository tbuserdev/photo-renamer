package renamer

import (
	"ImageRenamer/renamer/utility"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"
	"strings"
)

type Object struct {
	InputFolder     string
	OutputFolder    string
	ErrorFolder     string
	DuplicateFolder string
	ProgressBar     *widget.ProgressBar
}

func Rename(inputFolder string, outputFolder string, errorFolder string, duplicateFolder string, progressBar *widget.ProgressBar) error {

	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(errorFolder, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(duplicateFolder, 0755)
	if err != nil {
		return err
	}

	// EXIF DATA
	err = filepath.Walk(inputFolder, func(path string, info os.FileInfo, err error) error {

		// ARRAY OF FILES
		imageEndings := []string{
			".jpg", ".JPG",
			".jpeg", ".JPEG",
			".png", ".PNG",
			".gif", ".GIF",
			".bmp", ".BMP",
			".tiff", ".TIFF",
			".tif", ".TIF",
			".webp", ".WEBP",
			".heif", ".HEIF",
			".heic", ".HEIC",
			".arw", ".ARW",
			".cr2", ".CR2",
			".cr3", ".CR3",
			".dng", ".DNG",
			".nef", ".NEF",
			".rw2", ".RW2",
			".sr2", ".SR2",
			".srw", ".SRW",
		}

		// RENAME IMAGE
		if !info.IsDir() {
			if !strings.Contains(path, "@eaDir") && !strings.Contains(path, "Thumbs.db") &&
				!strings.Contains(path, "desktop.ini") && !strings.Contains(path, ".DS_Store") &&
				!strings.Contains(path, "._") && !strings.Contains(path, "._.") &&
				!strings.Contains(path, "Syno") && !strings.Contains(path, "syno") &&
				!strings.Contains(path, "SYNO") && !strings.Contains(path, "Synology") &&
				!strings.Contains(path, "thumb") && !strings.Contains(path, "Thumb") &&
				!strings.Contains(path, "THUMB") && !strings.Contains(path, "Thumbnails") {
				for _, ending := range imageEndings {
					if strings.HasSuffix(path, ending) {

						// CREATE NEW FILENAME
						newFileName := utility.Image(path)

						// ERROR OR MOVE
						if strings.Contains(newFileName, "error") {
							if _, err := os.Stat(errorFolder + "/" + info.Name()); os.IsNotExist(err) {
								err := os.Rename(path, errorFolder+"/"+info.Name())
								if err != nil {
									return err
								}
							} else {
								err := os.Rename(path, duplicateFolder+"/"+"DUPLICATE_"+info.Name())
								if err != nil {
									return err
								}
							}

						} else {
							_, err := os.Stat(outputFolder + "/" + newFileName)
							if os.IsNotExist(err) {
								err := os.Rename(path, outputFolder+"/"+newFileName)
								if err != nil {
									return err
								}
							} else {
								err := os.Rename(path, duplicateFolder+"/"+"DUPLICATE_"+newFileName)
								if err != nil {
									return err
								}
							}
						}
						progressBar.SetValue(progressBar.Value + 1)
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
