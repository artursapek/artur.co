package photos

type ItemsByTimestamp []ContentItem

func (items ItemsByTimestamp) Len() int {
	return len(items)
}

func (items ItemsByTimestamp) Less(i, j int) bool {
	return items[i].Timestamp().Before(items[j].Timestamp())
}

func (items ItemsByTimestamp) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}
