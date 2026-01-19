package renamer

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/tidwall/gjson"
)

func GetExifData(file string) (jsonString string) {
	imgFile, err := os.Open(file)
	if err != nil {
		return "error"
	}

	metaData, err := exif.Decode(imgFile)
	if err != nil {
		err = imgFile.Close()
		if err != nil {
			return "error"
		}
		return "error"
	}

	jsonByte, err := metaData.MarshalJSON()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = imgFile.Close()
	if err != nil {
		return "error"
	}
	jsonString = string(jsonByte)

	return jsonString
}

func date(metadata string) (date string) {
	dateTime := gjson.Get(metadata, "DateTimeOriginal").String()
	if dateTime == "" {
		dateTime = gjson.Get(metadata, "DateTimeDigitized").String()
	}
	if dateTime == "" {
		dateTime = gjson.Get(metadata, "DateTime").String()
	}
	if dateTime == "" {
		return ""
	}
	replacer := strings.NewReplacer(":", "-", " ", "_")
	return replacer.Replace(dateTime)
}

func model(metadata string) (model string) {
	model = gjson.Get(metadata, "Model").String()
	if model == "" {
		return "Unknown"
	}
	if strings.Contains(model, "(") {
		model = model[:strings.Index(model, "(")]
	}
	return model
}

func maker(metadata string) (maker string) {
	maker = gjson.Get(metadata, "Make").String()
	if maker == "" {
		return "Unknown"
	}
	return maker
}

func edited(metadata string) (edited string) {
	model := model(metadata)
	software := gjson.Get(metadata, "Software").String()

	if strings.Contains(software, model) {
		return model
	}
	if strings.Contains(software, "Lightroom") {
		return "Lightroom"
	}
	if strings.Contains(software, "Adobe Photoshop Lightroom Classic") {
		return "Lightroom"
	}
	if strings.Contains(software, "Adobe Photoshop") {
		return "Photoshop"
	}
	if strings.Contains(software, "Photoshop") {
		return "Photoshop"
	}
	if strings.Contains(software, "Photomator") {
		return "Photomator"
	}
	if strings.Contains(software, "Ver.1.0") {
		return model
	}
	return ""
}

func Image(file string) (newFileName string) {
	metadata := GetExifData(file)
	if metadata == "error" {
		return "METADATA_error"
	}

	fileExt := filepath.Ext(file)
	if fileExt == "" {
		return "FILEEXT_error"
	}

	date := date(metadata)
	if date == "" {
		return "DATE_error"
	}

	model := model(metadata)
	maker := maker(metadata)

	edited := edited(metadata)
	if edited == "" {
		edited = "Original"
	}

	if edited != model {
		newFileName = date + "_" + maker + "-" + model + "_" + edited + fileExt
		return newFileName
	}
	if edited == model {
		newFileName = date + "_" + maker + "-" + model + fileExt
		return newFileName
	}
	return "ERROR_error"
}

func OpenOutputFolder(folder string) (err error) {
	switch runtime.GOOS {
	case "darwin":
		err := exec.Command("open", "-R", folder).Run()
		if err != nil {
			log.Fatal(err)
		}
	case "windows":
		err := exec.Command("explorer", "/select,", folder).Run()
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("Unsupported operating system")
	}
	return err
}
