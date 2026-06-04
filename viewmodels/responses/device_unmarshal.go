package responses

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/abisaidfarias/lbtechapi/utils/bsonutil"
)

var defaultBSONRegistry = bson.DefaultRegistry

// UnmarshalBSON normaliza memory_extended cuando en MongoDB viene como string u otro tipo legacy.
func (d *Device) UnmarshalBSON(data []byte) error {
	var m bson.M
	if err := bson.UnmarshalWithRegistry(defaultBSONRegistry, data, &m); err != nil {
		return err
	}
	if v, ok := m["memory_extended"]; ok {
		m["memory_extended"] = bsonutil.CoerceToBool(v)
	}
	out, err := bson.MarshalWithRegistry(defaultBSONRegistry, m)
	if err != nil {
		return err
	}
	type deviceNoUnmarshaler Device
	return bson.UnmarshalWithRegistry(defaultBSONRegistry, out, (*deviceNoUnmarshaler)(d))
}

// UnmarshalBSON normaliza memory_extended cuando en MongoDB viene como string u otro tipo legacy.
func (d *DeviceExpanded) UnmarshalBSON(data []byte) error {
	var m bson.M
	if err := bson.UnmarshalWithRegistry(defaultBSONRegistry, data, &m); err != nil {
		return err
	}
	if v, ok := m["memory_extended"]; ok {
		m["memory_extended"] = bsonutil.CoerceToBool(v)
	}
	out, err := bson.MarshalWithRegistry(defaultBSONRegistry, m)
	if err != nil {
		return err
	}
	type deviceExpandedNoUnmarshaler DeviceExpanded
	return bson.UnmarshalWithRegistry(defaultBSONRegistry, out, (*deviceExpandedNoUnmarshaler)(d))
}
