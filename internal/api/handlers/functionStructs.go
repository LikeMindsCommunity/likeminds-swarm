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

type PostSecondaryDataParams struct {
	LikesCount       int
	RepliesCount     int
	RepostCount      int
	IsRepostedByUser bool
	IsLikedByUser    bool
	IsSavedByUser    bool
}
