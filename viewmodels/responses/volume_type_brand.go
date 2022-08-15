package responses

type VolumeTypeBrand struct {
	Type  int    `json:"type" bson:"type"`
	Brand string `json:"brand" bson:"brand"`
}
