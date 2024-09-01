package titleentry

import "fyne.io/fyne/v2/widget"

func CreateEntry(placeholder string) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	return entry
}
