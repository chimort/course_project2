package handlers

import "github.com/chimort/course_project2/api/proto/sharedpb"

type UserProfileResponse struct {
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	Age       int32                  `json:"age"`
	Gender    string                 `json:"gender"`
	Languages []*sharedpb.Language   `json:"languages"`
	Interests []*sharedpb.Interests  `json:"interests"`
}

