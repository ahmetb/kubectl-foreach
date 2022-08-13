// Copyright 2022 Twitter, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
