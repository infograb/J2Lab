package adf

func FindMentionIDs(adf *ADF) []string {
	var mentionIDs []string

	for _, block := range adf.Content {
		mentionIDs = append(mentionIDs, findMentionIDsFromBlock(block)...)
	}

	return mentionIDs
}

func findMentionIDsFromBlock(block *ADFBlock) []string {
	var mentionIDs []string

	if block.Type == "mention" {
		if idNode, found := block.Attrs["id"]; found {
			if id, ok := idNode.(string); ok {
				mentionIDs = append(mentionIDs, id)
			}
		}
	}

	for _, contentNode := range block.Content {
		mentionIDs = append(mentionIDs, findMentionIDsFromBlock(contentNode)...)
	}

	return mentionIDs
}
