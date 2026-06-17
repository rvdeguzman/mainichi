package ui

func promptSourceLabel(source string) string {
	switch source {
	case "ai":
		return "ai (deprecated)"
	default:
		return source
	}
}

func promptSourceDescription(source string) string {
	switch source {
	case "stoic":
		return "use the Daily Stoic heading"
	case "deck":
		return "draw from the prompt deck"
	case "ai":
		return "generate with AI (deprecated)"
	default:
		return source
	}
}
