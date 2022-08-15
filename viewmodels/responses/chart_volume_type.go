package responses

type ChartVolumeType struct {
	Keys  VolumeTypeBrand `json:"_id" bson:"_id"`
	Count int             `json:"count" bson:"count"  `
}
