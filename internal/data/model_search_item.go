package data

import "bactic/internal"

type SearchItem struct {
	Name     string `json:"name"`
	ItemType string `json:"item_type"`
	Link     string `json:"link"`
}

// AssertSearchItemRequired checks if the required fields are not zero-ed
func AssertSearchItemRequried(obj SearchItem) error {
	elements := map[string]interface{}{
		"name":      obj.Name,
		"item_type": obj.ItemType,
		"link":      obj.Link,
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
