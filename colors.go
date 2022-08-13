package main

var colors = []func(string, ...interface{}) string{
	// foreground only
	chalk.WithRed().Sprintf,
	chalk.WithBlue().Sprintf,
	chalk.WithGreen().Sprintf,
	chalk.WithYellow().WithBgBlack().Sprintf,
	chalk.WithGray().Sprintf,
	chalk.WithMagenta().Sprintf,
	chalk.WithCyan().Sprintf,
	chalk.WithBrightRed().Sprintf,

	chalk.WithBrightBlue().Sprintf,
	chalk.WithBrightGreen().Sprintf,
	chalk.WithBrightMagenta().Sprintf,
	chalk.WithBrightYellow().WithBgBlack().Sprintf,
	chalk.WithBrightCyan().Sprintf,

	// inverse
	chalk.WithBgRed().WithWhite().Sprintf,
	chalk.WithBgBlue().WithWhite().Sprintf,
	chalk.WithBgCyan().WithBlack().Sprintf,
	chalk.WithBgGreen().WithBlack().Sprintf,
	chalk.WithBgMagenta().WithBrightWhite().Sprintf,
	chalk.WithBgYellow().WithBlack().Sprintf,
	chalk.WithBgGray().WithWhite().Sprintf,
	chalk.WithBgBrightRed().WithWhite().Sprintf,
	chalk.WithBgBrightBlue().WithWhite().Sprintf,
	chalk.WithBgBrightCyan().WithBlack().Sprintf,
	chalk.WithBgBrightGreen().WithBlack().Sprintf,
	chalk.WithBgBrightMagenta().WithBlack().Sprintf,
	chalk.WithBgBrightYellow().WithBlack().Sprintf,

	// mixes+inverses
	chalk.WithBgRed().WithYellow().Sprintf,
	chalk.WithBgYellow().WithRed().Sprintf,
	chalk.WithBgBlue().WithYellow().Sprintf,
	chalk.WithBgYellow().WithBlue().Sprintf,
	chalk.WithBgBlack().WithBrightWhite().Sprintf,
	chalk.WithBgBrightWhite().WithBlack().Sprintf,
}
