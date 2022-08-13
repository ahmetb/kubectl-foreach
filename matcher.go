package main

func matchContexts(in []string, f []filter) []string {
	var additive, subtractive []filter
	for _, ff := range f {
		if ff.additive() {
			additive = append(additive, ff)
		} else {
			subtractive = append(subtractive, ff)
		}
	}

	var out []string
	for _, ctx := range in {
		add, remove := len(additive) == 0, false

		for _, af := range additive {
			if af.match(ctx) {
				add = true
				break
			}
		}

		for _, sf := range subtractive {
			if sf.match(ctx) {
				remove = true
				break
			}
		}

		if add && !remove {
			out = append(out, ctx)
		}
	}
	return out
}
