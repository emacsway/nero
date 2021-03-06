package sort

// Direction is a sort direction
type Direction int

func (d Direction) String() string {
	switch d {
	case Asc:
		return "Asc"
	case Desc:
		return "Desc"
	}

	return ""
}

// Desc is a sort description
func (d Direction) Desc() string {
	switch d {
	case Asc:
		return "ascending"
	case Desc:
		return "descending"
	}

	return ""
}

const (
	// Asc is an ascending sort direction
	Asc Direction = iota
	// Desc is a descending sort direction
	Desc
)
