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
	LikesCount       int64
	RepliesCount     int64
	RepostCount      int32
	IsRepostedByUser bool
	IsLikedByUser    bool
	IsSavedByUser    bool
}
