package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"car-aggregator/internal/repositories"
	"github.com/ekomobile/dadata/v2"
	"github.com/ekomobile/dadata/v2/api/clean"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/validators"
)

type SearchService interface {
	Search(ctx context.Context, in dtos.UserSearchQuery) (dtos.SearchResponse, error)
	SearchTrims(ctx context.Context, in dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error)
}

type searchService struct {
	repository   repositories.OfferRepository
	dadataAPI    *clean.Api
	validator    validators.UserSearchRequestValidator
	carAPIClient CarAPIClient
}

func NewSearchService(
	repository repositories.OfferRepository,
	validator validators.UserSearchRequestValidator,
) SearchService {
	api := dadata.NewCleanApi()
	carAPIClient := NewCarAPIClient(os.Getenv("CARAPI_TOKEN"), os.Getenv("CARAPI_SECRET"))

	return &searchService{
		repository:   repository,
		validator:    validator,
		dadataAPI:    api,
		carAPIClient: carAPIClient,
	}
}

func (s *searchService) Search(ctx context.Context, in dtos.UserSearchQuery) (dtos.SearchResponse, error) {
	if err := s.validator.ValidateSearch(ctx, in); err != nil {
		return dtos.SearchResponse{}, err
	}

	result, err := s.dadataAPI.Vehicle(ctx, in.Query)
	if err != nil {
		return dtos.SearchResponse{}, err
	}

	if len(result) == 0 {
		return dtos.SearchResponse{}, fmt.Errorf("unable to parse vehicle query: %s", in.Query)
	}

	vehicleData := result[0]

	vehicleInfo, err := s.repository.GetVehicleByManufacturerAndModel(ctx, vehicleData.Brand, vehicleData.Model)
	if err != nil {
		return dtos.SearchResponse{}, err
	}

	trimsResponse, err := s.carAPIClient.GetTrimsByModel(ctx, vehicleData.Model, 10)
	if err != nil {
		return dtos.SearchResponse{}, err
	}

	// ToDo: Here I can implement interaction with other services via the bus

	info := formatVehicleInfo(vehicleInfo, trimsResponse)
	return dtos.SearchResponse{Info: info}, nil
}

func (s *searchService) SearchTrims(ctx context.Context, in dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error) {
	if err := s.validator.ValidateSearch(ctx, in); err != nil {
		return dtos.SearchTrimsResponse{}, err
	}

	result, err := s.dadataAPI.Vehicle(ctx, in.Query)
	if err != nil {
		return dtos.SearchTrimsResponse{}, err
	}

	if len(result) == 0 {
		return dtos.SearchTrimsResponse{}, fmt.Errorf("unable to parse vehicle query: %s", in.Query)
	}

	vehicleData := result[0]
	trimsResponse, err := s.carAPIClient.GetTrimsByModel(ctx, vehicleData.Model, 8)
	if err != nil {
		return dtos.SearchTrimsResponse{}, err
	}

	resp := dtos.SearchTrimsResponse{
		Query: in.Query,
		Cars:  make([]dtos.SearchTrimItem, 0),
	}
	if trimsResponse == nil {
		return resp, nil
	}

	type specs struct {
		seats int
		fuel  float64
	}
	specsCache := map[string]specs{}
	imageCache := map[string]string{}

	for _, trim := range trimsResponse.Data {
		cacheKey := fmt.Sprintf("%s|%s", trim.Make, trim.Model)
		value, ok := specsCache[cacheKey]
		if !ok {
			seats, fuel, err := s.carAPIClient.GetVehicleSpecs(ctx, trim.Year, trim.Make, trim.Model)
			if err != nil {
				seats = 0
				fuel = 0
			}
			value = specs{seats: seats, fuel: fuel}
			specsCache[cacheKey] = value
		}

		if value.seats <= 0 {
			value.seats = inferSeats(trim.Description)
		}
		// Достаточно одной картинки на год/марку/модель, иначе поиск на каждый trim сильно тормозит.
		imageKey := fmt.Sprintf("%d|%s|%s", trim.Year, trim.Make, trim.Model)
		imageURL := imageCache[imageKey]
		if imageURL == "" {
			imageURL = buildInternetImageURL(trim.Year, trim.Make, trim.Model, "")
			imageCache[imageKey] = imageURL
		}
		transmission := inferTransmission(trim.Description, trim.Trim)
		resp.Cars = append(resp.Cars, dtos.SearchTrimItem{
			ID:           trim.ID,
			Year:         trim.Year,
			Make:         trim.Make,
			Model:        trim.Model,
			Trim:         trim.Trim,
			Description:  trim.Description,
			MSRP:         trim.MSRP,
			Transmission: transmission,
			Seats:        value.seats,
			Fuel:         value.fuel,
			ImageURL:     imageURL,
		})
	}
	resp.Count = len(resp.Cars)
	return resp, nil
}

func inferSeats(description string) int {
	raw := strings.ToLower(description)
	if strings.Contains(raw, "2dr") {
		return 2
	}
	if strings.Contains(raw, "van") || strings.Contains(raw, "minivan") {
		return 7
	}
	if strings.Contains(raw, "suv") {
		return 5
	}
	return 4
}

func buildInternetImageURL(year int, make, model, trim string) string {
	query := strings.TrimSpace(fmt.Sprintf("%d %s %s", year, make, model))
	if query == "" {
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(query))
	return "https://www.regcheck.org.uk/image.aspx/@" + encoded
}

func inferTransmission(description string, trimName string) string {
	raw := strings.ToLower(description + " " + trimName)
	if strings.Contains(raw, "cvt") ||
		strings.Contains(raw, "automatic") ||
		strings.Contains(raw, " auto") ||
		strings.Contains(raw, " 6a") ||
		strings.Contains(raw, " 7a") ||
		strings.Contains(raw, " 8a") ||
		strings.Contains(raw, " 9a") ||
		strings.Contains(raw, " 10a") {
		return "Automatic"
	}
	return "Manual"
}

func formatVehicleInfo(vehicleInfo interface{}, trims *dtos.TrimsResponse) string {
	result := fmt.Sprintf("Vehicle Information:\n%v\n\n", vehicleInfo)

	if trims != nil && len(trims.Data) > 0 {
		result += fmt.Sprintf("Available Trims (Total: %d):\n", trims.Collection.Total)
		for _, trim := range trims.Data {
			result += fmt.Sprintf("- %d %s %s %s (MSRP: $%d)\n %s\n",
				trim.Year, trim.Make, trim.Model, trim.Trim, trim.MSRP, trim.Description)
		}
	}
	return result
}
