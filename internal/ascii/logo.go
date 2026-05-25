package ascii

func Logo() string {
	return `
  ┌────────────────────────────────┐
  │       SWAGGER-EXP-KNIFE4J      │
  └────────────────────────────────┘
	`
}

func LogoHelp(s string) string {
	return Logo() + "\n\n" + s
}
