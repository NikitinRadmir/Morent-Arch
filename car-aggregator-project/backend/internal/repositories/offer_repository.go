package repositories

import (
	"context"
	"fmt"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/models"
	"gorm.io/gorm"
)

type OfferRepository interface {
	GetVehicleByManufacturerAndModel(ctx context.Context, manufacturer string, model string) ([]dtos.VehicleInfo, error)
	PurchaseOffer(ctx context.Context, req models.PurchaseRequest) error
	ReleaseOffer(ctx context.Context, req models.ReleaseRequest) error
	GetAvailableOffers(ctx context.Context, manufacturer, model string) ([]models.Offer, error)
	GetPurchasedOffers(ctx context.Context, serviceName string) ([]models.Offer, error)
	GetTakenCarIDs(ctx context.Context, carIDs []int) (map[int]bool, error)
}

type gormOfferRepository struct {
	db *gorm.DB
}

func NewOfferRepository(db *gorm.DB) OfferRepository {
	return &gormOfferRepository{db: db}
}

// Legacy support. В новой схеме хранится только offers, поэтому возвращаем пустой список.
func (r *gormOfferRepository) GetVehicleByManufacturerAndModel(ctx context.Context, manufacturer string, model string) ([]dtos.VehicleInfo, error) {
	return []dtos.VehicleInfo{}, nil
}

func (r *gormOfferRepository) PurchaseOffer(ctx context.Context, req models.PurchaseRequest) error {
	offer := models.Offer{
		CustomerService: req.CustomerService,
		CarID:           req.CarID,
		Make:            req.Make,
		Model:           req.Model,
		Trim:            req.Trim,
		Price:           req.Price,
	}
	if err := r.db.WithContext(ctx).Create(&offer).Error; err != nil {
		return fmt.Errorf("create offer: %w", err)
	}
	return nil
}

func (r *gormOfferRepository) ReleaseOffer(ctx context.Context, req models.ReleaseRequest) error {
	res := r.db.WithContext(ctx).
		Where("customer_service = ? AND car_id = ?", req.CustomerService, req.CarID).
		Delete(&models.Offer{})
	if res.Error != nil {
		return fmt.Errorf("release offer: %w", res.Error)
	}
	return nil
}

func (r *gormOfferRepository) GetAvailableOffers(ctx context.Context, manufacturer, model string) ([]models.Offer, error) {
	// "available" в новой модели не используется — всё, что в offers, уже занято.
	return []models.Offer{}, nil
}

func (r *gormOfferRepository) GetPurchasedOffers(ctx context.Context, serviceName string) ([]models.Offer, error) {
	var offers []models.Offer
	q := r.db.WithContext(ctx).Model(&models.Offer{}).Order("created_at desc")
	if serviceName != "" {
		q = q.Where("customer_service = ?", serviceName)
	}
	if err := q.Find(&offers).Error; err != nil {
		return nil, fmt.Errorf("get purchased offers: %w", err)
	}
	return offers, nil
}

func (r *gormOfferRepository) GetTakenCarIDs(ctx context.Context, carIDs []int) (map[int]bool, error) {
	out := map[int]bool{}
	if len(carIDs) == 0 {
		return out, nil
	}

	type row struct {
		CarID int
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&models.Offer{}).
		Select("car_id").
		Where("car_id IN ?", carIDs).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("get taken car ids: %w", err)
	}
	for _, r := range rows {
		out[r.CarID] = true
	}
	return out, nil
}
