package repositories

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Countries screen (Administration) queries.

// Modules whose records store a country id, and where.
const (
	CountryModuleHomologation    = "homologation"
	CountryModuleTracking        = "tracking"
	CountryModuleShipmentControl = "shipment_control"
)

// byName marks references stored as a country name instead of an id:
// tracking logs embed a snapshot of the country without its _id.
var countryUsageFields = []struct {
	module     string
	collection string
	field      string
	byName     bool
}{
	{CountryModuleHomologation, "homologations", "country", false},
	{CountryModuleTracking, "device_trackings", "tracking_logs.country.name", true},
	{CountryModuleShipmentControl, "shipment_controls", "country", false},
}

func countryNameKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// UpdateFields applies a partial $set (only the given fields change).
func (r *countryRepository) UpdateFields(id primitive.ObjectID, set primitive.M) error {
	res, err := countryCollection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$set": set})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

// FindByNameExcluding finds a country by name (trimmed, case-insensitive,
// exact), ignoring exclude (the country being edited).
func (r *countryRepository) FindByNameExcluding(name string, exclude primitive.ObjectID) (*responses.Country, error) {
	filter := bson.M{"name": primitive.Regex{
		Pattern: "^\\s*" + regexp.QuoteMeta(strings.TrimSpace(name)) + "\\s*$",
		Options: "i",
	}}
	if !exclude.IsZero() {
		filter["_id"] = bson.M{"$ne": exclude}
	}
	var c responses.Country
	err := countryCollection.FindOne(context.TODO(), filter).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CountryUsage maps each country id to the modules that reference it (in the
// order homologation, tracking, shipment_control).
func (r *countryRepository) CountryUsage() (map[primitive.ObjectID][]string, error) {
	db := countryCollection.Database()

	// Name → id, for the references stored by name.
	countries, err := r.Get()
	if err != nil {
		return nil, err
	}
	idByName := make(map[string]primitive.ObjectID, len(countries))
	for _, c := range countries {
		idByName[countryNameKey(c.Name)] = c.ID
	}

	out := map[primitive.ObjectID][]string{}
	for _, u := range countryUsageFields {
		values, err := db.Collection(u.collection).Distinct(context.TODO(), u.field, bson.M{})
		if err != nil {
			return nil, err
		}
		seen := map[primitive.ObjectID]bool{}
		for _, v := range values {
			var id primitive.ObjectID
			if u.byName {
				name, _ := v.(string)
				id = idByName[countryNameKey(name)]
			} else {
				id, _ = v.(primitive.ObjectID)
			}
			if id.IsZero() || seen[id] {
				continue
			}
			seen[id] = true
			out[id] = append(out[id], u.module)
		}
	}
	return out, nil
}

// UsersCountByCountry counts the users whose countries list includes each
// country. Internal users with an empty list (all countries) do not count.
func (r *countryRepository) UsersCountByCountry() (map[primitive.ObjectID]int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$project", Value: bson.M{"countries": bson.M{"$setUnion": bson.A{bson.M{"$ifNull": bson.A{"$countries", bson.A{}}}}}}}},
		{{Key: "$unwind", Value: "$countries"}},
		{{Key: "$group", Value: bson.M{"_id": "$countries", "count": bson.M{"$sum": 1}}}},
	}
	return countByID(userCollection, pipeline)
}
