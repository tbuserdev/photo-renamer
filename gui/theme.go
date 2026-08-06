package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// appTheme keeps Fyne's platform font and icons while using a restrained,
// adaptive palette and softer geometry suited to a focused desktop utility.
type appTheme struct {
	base fyne.Theme
}

func newAppTheme() fyne.Theme {
	return &appTheme{base: theme.DefaultTheme()}
}

func (t *appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	dark := variant == theme.VariantDark
	switch name {
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 10, G: 132, B: 255, A: 255}
	case theme.ColorNameBackground:
		if dark {
			return color.NRGBA{R: 28, G: 28, B: 30, A: 255}
		}
		return color.NRGBA{R: 245, G: 245, B: 247, A: 255}
	case theme.ColorNameInputBackground, theme.ColorNameMenuBackground:
		if dark {
			return color.NRGBA{R: 44, G: 44, B: 46, A: 255}
		}
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameSeparator:
		if dark {
			return color.NRGBA{R: 84, G: 84, B: 88, A: 180}
		}
		return color.NRGBA{R: 60, G: 60, B: 67, A: 45}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 52, G: 199, B: 89, A: 255}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 159, B: 10, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 255, G: 69, B: 58, A: 255}
	}
	return t.base.Color(name, variant)
}

func (t *appTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *appTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *appTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 10
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameButtonRadius, theme.SizeNameCardRadius, theme.SizeNameDialogRadius, theme.SizeNameInputRadius:
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 17
	}
	return t.base.Size(name)
}
