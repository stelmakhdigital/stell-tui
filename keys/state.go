package keys

var kittyProtocolActive bool

// SetKittyProtocolActive фиксирует, согласован ли протокол клавиатуры Kitty.
func SetKittyProtocolActive(active bool) {
	kittyProtocolActive = active
}

// KittyProtocolActive сообщает, активен ли протокол клавиатуры Kitty.
func KittyProtocolActive() bool {
	return kittyProtocolActive
}
