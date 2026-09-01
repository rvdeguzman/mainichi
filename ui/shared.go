package ui

func promptSourceDescription(source string) string {
	switch source {
	case "stoic":
		return "use the Daily Stoic heading"
	case "deck":
		return "draw from the prompt deck"
	default:
		return source
	}
}
