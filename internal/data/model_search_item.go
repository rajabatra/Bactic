package data

import "bactic/internal"

type SearchItemType uint32

const (
	ATHLETE SearchItemType = 1 << iota
	SCHOOL
	REGION
)

func (s SearchItemType) Validate() bool {
	return s >= ATHLETE && s <= REGION
}

type SearchItem struct {
	Name     string         `json:"name"`
	ItemType uint32         `json:"item_type"`
	Id       SearchItemType `json:"id"`
}

// AssertSearchItemRequired checks if the required fields are not zero-ed
func AssertSearchItemRequried(obj SearchItem) error {
	elements := map[string]interface{}{
		"name":      obj.Name,
		"item_type": obj.ItemType,
		"link":      obj.Id,
	}
	for name, el := range elements {
		if isZero := internal.IsZeroValue(el); isZero {
			return &internal.RequiredError{Field: name}
		}
	}

	return nil
}

// AssertSearchItemConstraints checks if the values respects the defined constraints
func AssertSearchItemConstraints(obj SearchItem) error {
	return nil
}
