package responses

type StackedSerieChart struct {
	Group       string `json:"group" bson:"group"`
	XAxis      []int  `json:"x_axis" bson:"x_axis"`
	Descripcion string `json:"description" bson:"description"`
}
