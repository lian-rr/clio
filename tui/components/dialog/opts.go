package dialog

// OptFunc optional configs for the Dialog.
type OptFunc func(*Dialog)

// WithButtonNames used for setting custom names for the buttons.
func WithButtonNames(accept, cancel string) OptFunc {
	return func(d *Dialog) {
		d.acceptButtonLabel = accept
		d.cancelButtonLabel = cancel
	}
}
