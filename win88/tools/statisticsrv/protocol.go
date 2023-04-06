package main

type FileReqMsg struct {
	FileName   string `json:"filename"`
	Size       int64  `json:"size"`
	Url        string `json:"url"`
	StatusCode string `json:"statuscode"`
	ErrorCode  string `json:"errorcode"`
	ErrorMsg   string `json:"errormsg"`
}

type TopNFileMsg struct {
	Id    string `bson:"_id"`
	Total int64  `bson:"total"`
	Size  int64  `bson:"size"`
}
