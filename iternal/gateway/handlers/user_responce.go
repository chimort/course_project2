package handlers

import "github.com/chimort/course_project2/api/proto/sharedpb"

type UserProfileResponce struct {
	Username string `json:"username"`
	Language  []*sharedpb.Language  `json:"language"`
	Interests []*sharedpb.Interests `json:"interests"`
}

