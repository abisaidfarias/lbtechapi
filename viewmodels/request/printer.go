package request

// User model
type Printer struct {
	Modelo       string `bson:"modelo" json:"modelo,omitempty"`
	Serial       string `bson:"serial" json:"serial,omitempty"`
	Pages        string `bson:"paginas" json:"paginas,omitempty"`
	Location     string `bson:"ubicacion" json:"ubicacion,omitempty"`
	MaxToner     string `bson:"maxtoner" json:"maxtoner,omitempty"`
	RemToner     string `bson:"remtoner" json:"remtoner,omitempty"`
	SNconsumible string `bson:"SNconsumible" json:"SNconsumible,omitempty"`
	PNconsumible string `bson:"PNconsumible" json:"PNconsumible,omitempty"`
	Rat          string `bson:"rat" json:"rat,omitempty"`
	Level        string `bson:"level" json:"level,omitempty"`
}
