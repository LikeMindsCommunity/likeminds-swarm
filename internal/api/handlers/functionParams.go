package handlers

// Params Structure for Logged In User
type LoggedInUserParams struct {
	UserId           string
	CommunityId      int
	IsCm             bool
	PlatformCode     string
	VersionCode      string
	ApiRevampCheckV1 bool
	MemberRole       string
}

type PostImpressionsData struct {
	ImpressionsCount int `json:"impressions_count"`
	ReachCount       int `json:"reach_count"`
}

// Params Structure for Secondary Post Data
type PostSecondaryDataParams struct {
	LikesCount       int
	RepliesCount     int
	RepostCount      int
	IsRepostedByUser bool
	IsLikedByUser    bool
	IsSavedByUser    bool
	PostImpressions  PostImpressionsData
}
