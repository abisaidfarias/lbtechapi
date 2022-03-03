package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"

	"github.com/abisaidfarias/lbtechapi/database"
	"github.com/abisaidfarias/lbtechapi/database/queries"
	"github.com/abisaidfarias/lbtechapi/models"
	"github.com/abisaidfarias/lbtechapi/utils/enums"
	"github.com/abisaidfarias/lbtechapi/utils/functions"
	"github.com/abisaidfarias/lbtechapi/viewmodels/request"
	"github.com/abisaidfarias/lbtechapi/viewmodels/responses"
)

type IHomologationRepository interface {
	Create(*models.Homologation) error
	GetPrevious(primitive.ObjectID,
		primitive.ObjectID, primitive.ObjectID, bool) (*responses.Homologation, error)
	GetByInternal([]primitive.ObjectID, []primitive.ObjectID,
		[]primitive.ObjectID) ([]*responses.HomologationExpanded, error)
	GetByExternal(primitive.ObjectID, []primitive.ObjectID,
		[]primitive.ObjectID) ([]*responses.HomologationExpanded, error)
	GetByID(primitive.ObjectID) (*responses.Homologation, error)
	UpdateTestResult(string, request.TestResultResume) error
	CreateFailTestResult(string, *models.TestResult) error
	GetGroupedByTypeCountry(companies []primitive.ObjectID,
		devices []primitive.ObjectID, countries []primitive.ObjectID,
		companyId primitive.ObjectID, isInternal bool) ([]*responses.ChartTypeCountry, error)
	PhaseChange(string, *models.Homologation) error
	GetGroupedByBrandCountry(companies []primitive.ObjectID,
		devices []primitive.ObjectID, countries []primitive.ObjectID) ([]*responses.ChartVolumeCountry, error)
	GetGroupedByBrandType(companies []primitive.ObjectID,
		devices []primitive.ObjectID, countries []primitive.ObjectID) ([]*responses.ChartVolumeType, error)
	GetByTestPlan(testPlan primitive.ObjectID) (*responses.Homologation, error)
	UpdateDocument(string, primitive.ObjectID) error
	GetByCompanyStatusGrouped(primitive.ObjectID) ([]*responses.DashboardTotal, error)
	GetByCountry(primitive.ObjectID) (*responses.Homologation, error)
	GetByCompany(primitive.ObjectID) (*responses.Homologation, error)
	GetByDevice(primitive.ObjectID) (*responses.Homologation, error)
	Update(primitive.ObjectID, *models.Homologation) error
	Delete(primitive.ObjectID) error
	DeleteHierarchy(primitive.ObjectID, primitive.ObjectID, primitive.ObjectID, bool) error
}
type homologationRepository struct {
}

func NewHomologationRepository() IHomologationRepository {
	return &homologationRepository{}
}

var homologationCollection = database.GetInstance().Collection("homologations")

// Create a new tet case
func (r *homologationRepository) Create(homologation *models.Homologation) error {

	_, err := homologationCollection.InsertOne(context.TODO(), homologation)

	if err != nil {
		return err
	}
	return nil
}

// Get returns a list of all test cases
func (r *homologationRepository) Get() ([]*responses.Homologation, error) {

	cursor, err := homologationCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}
	var homologations []*responses.Homologation = []*responses.Homologation{}
	if err = cursor.All(context.TODO(), &homologations); err != nil {
		panic(err)
	}
	cursor.Close(context.TODO())
	return homologations, nil
}

// Get returns a list of all test cases
func (r *homologationRepository) GetPrevious(deviceId primitive.ObjectID,
	countryId primitive.ObjectID, companyId primitive.ObjectID, isIntenal bool) (*responses.Homologation, error) {

	var homologations []*responses.Homologation
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": -1})
	cursor, err := homologationCollection.Find(context.TODO(),
		queries.GetHomologationValidations(deviceId, countryId, companyId, isIntenal), findOptions)
	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	if err = cursor.All(context.TODO(), &homologations); err != nil {
		return nil, err
	}
	if len(homologations) > 0 {
		return homologations[0], nil
	}

	return nil, nil
}
func (r *homologationRepository) GetByInternal(companies []primitive.ObjectID,
	devices []primitive.ObjectID, countries []primitive.ObjectID) ([]*responses.HomologationExpanded, error) {

	cursor, err := homologationCollection.Aggregate(context.TODO(),
		queries.GetHomologations(companies, devices, countries, true, primitive.ObjectID{}))
	if err != nil {
		return nil, err
	}
	var homologations []*responses.HomologationExpanded = []*responses.HomologationExpanded{}
	for cursor.Next(context.TODO()) {
		var homologation *responses.HomologationExpanded
		err := cursor.Decode(&homologation)
		if err != nil {
			return nil, err
		}
		homologation.OsVersion = homologation.Device.PlatformOs + " " + homologation.OsVersion
		homologation.ApprovalType = enums.HomologationType_key[homologation.Type]
		if homologation.Type == enums.HomologationType_value["MAINTENANCE"] {
			homologation.ApprovalType = homologation.ApprovalTypeOption
		}
		homologation.ProjectType = "External"
		if homologation.IsInternalProject {
			homologation.ProjectType = "Internal"
		}
		homologation.StatusView = enums.HomologationStatus_type[homologation.Status]

		functions.SetHomologationDatesToNull(homologation)
		homologations = append(homologations, homologation)
	}

	return homologations, nil
}
func (r *homologationRepository) GetByExternal(companyID primitive.ObjectID,
	devices []primitive.ObjectID, countries []primitive.ObjectID) ([]*responses.HomologationExpanded, error) {

	cursor, err := homologationCollection.Aggregate(context.TODO(),
		queries.GetHomologations([]primitive.ObjectID{}, devices, countries, false, companyID))
	if err != nil {
		return nil, err
	}
	var homologations []*responses.HomologationExpanded = []*responses.HomologationExpanded{}
	for cursor.Next(context.TODO()) {
		var homologation *responses.HomologationExpanded
		err := cursor.Decode(&homologation)
		if err != nil {
			return nil, err
		}
		homologation.OsVersion = homologation.Device.PlatformOs + " " + homologation.OsVersion
		homologation.ApprovalType = enums.HomologationType_key[homologation.Type]
		homologation.ProjectType = "External"
		if homologation.IsInternalProject {
			homologation.ProjectType = "Internal"
		}
		homologation.StatusView = enums.HomologationStatus_type[homologation.Status]

		functions.SetHomologationDatesToNull(homologation)
		homologations = append(homologations, homologation)
	}

	return homologations, nil
}
func (r *homologationRepository) GetByID(homologationID primitive.ObjectID) (*responses.Homologation, error) {

	var homologation *responses.Homologation
	err := homologationCollection.FindOne(context.TODO(),
		queries.GetHomologationById(homologationID)).Decode(&homologation)
	if err != nil {
		return nil, err
	}
	return homologation, nil
}
func (r *homologationRepository) UpdateTestResult(id string, testResult request.TestResultResume) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter, update := queries.UpdateTestResult(testResult, oid)

	_, err = homologationCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *homologationRepository) CreateFailTestResult(id string, testResult *models.TestResult) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter, update := queries.CreateTestResult(testResult, oid)

	_, err = homologationCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *homologationRepository) PhaseChange(id string, homologation *models.Homologation) error {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter, update := queries.UpdatePhaseChange(homologation, oid)

	_, err = homologationCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}

func (r *homologationRepository) GetGroupedByTypeCountry(companies []primitive.ObjectID,
	devices []primitive.ObjectID, countries []primitive.ObjectID,
	companyId primitive.ObjectID, isInternal bool) ([]*responses.ChartTypeCountry, error) {

	homologationQuery := queries.GetHomologations(companies, devices, countries, isInternal, companyId)
	groupedQuery := queries.GetHomologationsGroupedCountryApprovalType()
	match := queries.IntervalTime()
	homologationQuery = append(homologationQuery, match)
	sort := queries.SortGroupedCountryApprovalType()
	homologationQuery = append(homologationQuery, groupedQuery)
	homologationQuery = append(homologationQuery, sort)

	cursor, err := homologationCollection.Aggregate(context.TODO(),
		homologationQuery)
	if err != nil {
		return nil, err
	}
	var charts []*responses.ChartTypeCountry = []*responses.ChartTypeCountry{}
	for cursor.Next(context.TODO()) {
		var chart *responses.ChartTypeCountry
		err := cursor.Decode(&chart)
		if err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}

	return charts, nil
}

func (r *homologationRepository) GetGroupedByBrandCountry(companies []primitive.ObjectID,
	devices []primitive.ObjectID, countries []primitive.ObjectID) ([]*responses.ChartVolumeCountry, error) {

	homologationQuery := queries.GetHomologations(companies, devices, countries, true, primitive.ObjectID{})
	groupedQuery := queries.GetHomologationsGroupedCountryBrand()
	sort := queries.SortGroupedCountryBrand()
	homologationQuery = append(homologationQuery, groupedQuery)
	homologationQuery = append(homologationQuery, sort)

	cursor, err := homologationCollection.Aggregate(context.TODO(),
		homologationQuery)
	if err != nil {
		return nil, err
	}
	var charts []*responses.ChartVolumeCountry = []*responses.ChartVolumeCountry{}
	for cursor.Next(context.TODO()) {
		var chart *responses.ChartVolumeCountry
		err := cursor.Decode(&chart)
		if err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}

	return charts, nil
}
func (r *homologationRepository) GetGroupedByBrandType(companies []primitive.ObjectID,
	devices []primitive.ObjectID, countries []primitive.ObjectID) ([]*responses.ChartVolumeType, error) {

	homologationQuery := queries.GetHomologations(companies, devices, countries, true, primitive.ObjectID{})
	groupedQuery := queries.GetHomologationsGroupedTypeBrand()
	sort := queries.SortGroupedBrandType()
	homologationQuery = append(homologationQuery, groupedQuery)
	homologationQuery = append(homologationQuery, sort)

	cursor, err := homologationCollection.Aggregate(context.TODO(),
		homologationQuery)
	if err != nil {
		return nil, err
	}
	var charts []*responses.ChartVolumeType = []*responses.ChartVolumeType{}
	for cursor.Next(context.TODO()) {
		var chart *responses.ChartVolumeType
		err := cursor.Decode(&chart)
		if err != nil {
			return nil, err
		}
		charts = append(charts, chart)
	}

	return charts, nil
}
func (r *homologationRepository) GetByTestPlan(testPlanId primitive.ObjectID) (*responses.Homologation, error) {

	var homologation *responses.Homologation
	err := homologationCollection.FindOne(context.TODO(),
		queries.GetHomologationByTestPlan(testPlanId)).Decode(&homologation)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	return homologation, nil
}
func (r *homologationRepository) UpdateDocument(documentUrl string, homologationId primitive.ObjectID) error {

	filter, update := queries.UpdateDocument(documentUrl, homologationId)

	_, err := homologationCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *homologationRepository) GetByCompanyStatusGrouped(companyId primitive.ObjectID) ([]*responses.DashboardTotal, error) {

	homologationQuery := queries.GetHomologationsByCompanyD(companyId)
	groupedQuery := queries.GetHomologationsGroupedByStatus()

	homologationQuery = append(homologationQuery, groupedQuery)

	cursor, err := homologationCollection.Aggregate(context.TODO(),
		homologationQuery)
	if err != nil {
		return nil, err
	}
	var dashboardTotals []*responses.DashboardTotal = []*responses.DashboardTotal{}
	for cursor.Next(context.TODO()) {
		var dashboardTotal *responses.DashboardTotal
		err := cursor.Decode(&dashboardTotal)
		if err != nil {
			return nil, err
		}
		dashboardTotals = append(dashboardTotals, dashboardTotal)
	}
	return dashboardTotals, nil
}
func (r *homologationRepository) GetByCountry(countryId primitive.ObjectID) (*responses.Homologation, error) {

	var homologation *responses.Homologation
	err := homologationCollection.FindOne(context.TODO(),
		queries.GetHomologationsByCountry(countryId)).Decode(&homologation)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	return homologation, nil
}
func (r *homologationRepository) GetByCompany(companyId primitive.ObjectID) (*responses.Homologation, error) {

	var homologation *responses.Homologation
	err := homologationCollection.FindOne(context.TODO(),
		queries.GetHomologationsByCompany(companyId)).Decode(&homologation)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	return homologation, nil
}
func (r *homologationRepository) GetByDevice(deviceId primitive.ObjectID) (*responses.Homologation, error) {

	var homologation *responses.Homologation
	err := homologationCollection.FindOne(context.TODO(),
		queries.GetHomologationsByDevice(deviceId)).Decode(&homologation)

	if err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, nil
		default:
			return nil, err
		}
	}
	return homologation, nil
}
func (r *homologationRepository) Update(id primitive.ObjectID, homologation *models.Homologation) error {

	filter, update := queries.UpdateHomologation(homologation, id)

	_, err := homologationCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}

	return nil
}
func (r *homologationRepository) Delete(id primitive.ObjectID) error {

	_, err := homologationCollection.DeleteOne(context.TODO(), queries.DeleteHomologation(id))

	if err != nil {
		return err
	}
	return nil
}
func (r *homologationRepository) DeleteHierarchy(deviceId primitive.ObjectID,
	countryId primitive.ObjectID, companyId primitive.ObjectID, isInternal bool) error {

	_, err := homologationCollection.DeleteMany(context.TODO(),
		queries.DeleteHomologationHierarchy(deviceId, countryId, companyId, isInternal))

	if err != nil {
		return err
	}
	return nil
}
