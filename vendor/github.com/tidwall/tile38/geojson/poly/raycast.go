package poly

import "math"

type rayres struct {
	in, on bool
}

func raycast(p, a, b Point) rayres {
	// make sure that the point is inside the segment bounds
	if a.Y < b.Y && (p.Y < a.Y || p.Y > b.Y) {
		return rayres{false, false}
	} else if a.Y > b.Y && (p.Y < b.Y || p.Y > a.Y) {
		return rayres{false, false}
	}

	// test if point is in on the segment
	if a.Y == b.Y {
		if a.X == b.X {
			if p == a {
				return rayres{false, true}
			} else {
				return rayres{false, false}
			}
		}
		if p.Y == b.Y {
			// horizontal segment
			// check if the point in on the line
			if a.X < b.X {
				if p.X >= a.X && p.X <= b.X {
					return rayres{false, true}
				}
			} else {
				if p.X >= b.X && p.X <= a.X {
					return rayres{false, true}
				}
			}
		}
	}
	if a.X == b.X && p.X == b.X {
		// vertical segment
		// check if the point in on the line
		if a.Y < b.Y {
			if p.Y >= a.Y && p.Y <= b.Y {
				return rayres{false, true}
			}
		} else {
			if p.Y >= b.Y && p.Y <= a.Y {
				return rayres{false, true}
			}
		}
	}
	if (p.X-a.X)/(b.X-a.X) == (p.Y-a.Y)/(b.Y-a.Y) {
		return rayres{false, true}
	}

	// do the actual raycast here.
	for p.Y == a.Y || p.Y == b.Y {
		p.Y = math.Nextafter(p.Y, math.Inf(1))
	}
	if a.Y < b.Y {
		if p.Y < a.Y || p.Y > b.Y {
			return rayres{false, false}
		}
	} else {
		if p.Y < b.Y || p.Y > a.Y {
			return rayres{false, false}
		}
	}
	if a.X > b.X {
		if p.X > a.X {
			return rayres{false, false}
		}
		if p.X < b.X {
			return rayres{true, false}
		}
	} else {
		if p.X > b.X {
			return rayres{false, false}
		}
		if p.X < a.X {
			return rayres{true, false}
		}
	}
	if a.Y < b.Y {
		if (p.Y-a.Y)/(p.X-a.X) >= (b.Y-a.Y)/(b.X-a.X) {
			return rayres{true, false}
		}
	} else {
		if (p.Y-b.Y)/(p.X-b.X) >= (a.Y-b.Y)/(a.X-b.X) {
			return rayres{true, false}
		}
	}
	return rayres{false, false}
}
