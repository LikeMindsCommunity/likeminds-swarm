package searchElastic

const _ = `
{
	"from": 1,
	"size": 10,
	"sort": {
		"_score": {
			"order": "desc"
		},
		"updated_at": {
			"order": "desc"
		}
	},
	"query": {
		"bool": {
			"must": [
				{
					"term": {"member.id": self.get_member_id()}
				},
				{
					"term": {"follow_status": self.get_follow_status()}
				},
				{
					"bool": {
						"should": [
							{
								"match": {
									f"chatroom.{self.get_search_field()}": {
										"query": self.get_search_term(),
										"analyzer": "standard"
									}
								}
							}
						]
					}
				}
			],
			"must_not": [
				{
					"term": {"chatroom.type": card_types.CARD_DIRECT_MESSAGE}
				}
			]
		}
	}
}
`
