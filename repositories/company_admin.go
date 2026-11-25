package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Companies screen (Administration) queries.

// oabiCertificateIssuedFilter means "an OABI certificate was issued": the
// generated certificate PDF is stored on the shipment control. Rejected
// controls have it cleared, so they never count.
var oabiCertificateIssuedFilter = bson.M{
	"oabi_certificate_url": bson.M{"$exists": true, "$nin": bson.A{"", nil}},
}

// GetByClientID returns a company using clientID (trimmed, exact match),
// ignoring exclude (the company being edited).
func (r *companyRepository) GetByClientID(clientID string, exclude primitive.ObjectID) (*responses.Company, error) {
	filter := bson.M{"client_id": strings.TrimSpace(clientID)}
	if !exclude.IsZero() {
		filter["_id"] = bson.M{"$ne": exclude}
	}
	var c responses.Company
	err := companyCollection.FindOne(context.TODO(), filter).Decode(&c)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UsersCountByCompany counts, per company, the users whose company or clients
// include it (a user counts once per company).
func (r *companyRepository) UsersCountByCompany() (map[primitive.ObjectID]int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$project", Value: bson.M{
			"companies": bson.M{"$setUnion": bson.A{
				bson.M{"$cond": bson.A{bson.M{"$ifNull": bson.A{"$company", false}}, bson.A{"$company"}, bson.A{}}},
				bson.M{"$ifNull": bson.A{"$clients", bson.A{}}},
			}},
		}}},
		{{Key: "$unwind", Value: "$companies"}},
		{{Key: "$group", Value: bson.M{"_id": "$companies", "count": bson.M{"$sum": 1}}}},
	}
	return countByID(userCollection, pipeline)
}

// IssuedCertificatesByCompany counts issued OABI certificates per company.
func (r *companyRepository) IssuedCertificatesByCompany() (map[primitive.ObjectID]int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: oabiCertificateIssuedFilter}},
		{{Key: "$group", Value: bson.M{"_id": "$company", "count": bson.M{"$sum": 1}}}},
	}
	return countByID(shipmentControlCollection, pipeline)
}

// CountIssuedCertificates counts issued OABI certificates of one company.
func (r *companyRepository) CountIssuedCertificates(companyID primitive.ObjectID) (int64, error) {
	filter := bson.M{"company": companyID}
	for k, v := range oabiCertificateIssuedFilter {
		filter[k] = v
	}
	return shipmentControlCollection.CountDocuments(context.TODO(), filter)
}

// GetCompanyUsers lists the users whose company or clients include it, with
// their profile name.
func (r *companyRepository) GetCompanyUsers(companyID primitive.ObjectID) ([]responses.CompanyUser, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"$or": bson.A{
			bson.M{"company": companyID},
			bson.M{"clients": companyID},
		}}}},
		{{Key: "$lookup", Value: bson.M{
			"from": "profiles", "localField": "profile", "foreignField": "_id", "as": "profile_doc",
		}}},
		{{Key: "$project", Value: bson.M{
			"name": 1, "lastName": 1, "email": 1, "is_internal": 1,
			"profile_name": bson.M{"$ifNull": bson.A{bson.M{"$arrayElemAt": bson.A{"$profile_doc.name", 0}}, ""}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "name", Value: 1}, {Key: "lastName", Value: 1}}}},
	}
	cur, err := userCollection.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	var rows []struct {
		ID          primitive.ObjectID `bson:"_id"`
		Name        string             `bson:"name"`
		LastName    string             `bson:"lastName"`
		Email       string             `bson:"email"`
		IsInternal  bool               `bson:"is_internal"`
		ProfileName string             `bson:"profile_name"`
	}
	if err := cur.All(context.TODO(), &rows); err != nil {
		return nil, err
	}
	users := make([]responses.CompanyUser, 0, len(rows))
	for _, u := range rows {
		users = append(users, responses.CompanyUser{
			ID: u.ID.Hex(), Name: u.Name, LastName: u.LastName, Email: u.Email,
			ProfileName: u.ProfileName, IsInternal: u.IsInternal,
		})
	}
	return users, nil
}

func countByID(coll *mongo.Collection, pipeline mongo.Pipeline) (map[primitive.ObjectID]int, error) {
	cur, err := coll.Aggregate(context.TODO(), pipeline, options.Aggregate())
	if err != nil {
		return nil, err
	}
	var rows []struct {
		ID    primitive.ObjectID `bson:"_id"`
		Count int                `bson:"count"`
	}
	if err := cur.All(context.TODO(), &rows); err != nil {
		return nil, err
	}
	out := make(map[primitive.ObjectID]int, len(rows))
	for _, row := range rows {
		out[row.ID] = row.Count
	}
	return out, nil
}
