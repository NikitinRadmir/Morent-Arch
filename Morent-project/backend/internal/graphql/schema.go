package graphql

import (
	"github.com/graphql-go/graphql"
	"morent-backend/internal/di"
	"morent-backend/internal/models"
)

// строит GraphQL-схему, обёртывающую существующие сервисы.
func NewSchema(c *di.Container) (graphql.Schema, error) {
	carType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Car",
		Fields: graphql.Fields{
			"id":           &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"name":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"type":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"capacity":     &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"price":        &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"description":  &graphql.Field{Type: graphql.String},
			"imgSrc":       &graphql.Field{Type: graphql.String},
			"fuel":         &graphql.Field{Type: graphql.Float},
			"transmission": &graphql.Field{Type: graphql.String},
		},
	})

	commentType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Comment",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"carId":       &graphql.Field{Type: graphql.NewNonNull(graphql.Int)},
			"userId":      &graphql.Field{Type: graphql.Int},
			"name":        &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"post":        &graphql.Field{Type: graphql.String},
			"photo":       &graphql.Field{Type: graphql.String},
			"date":        &graphql.Field{Type: graphql.String},
			"rating":      &graphql.Field{Type: graphql.Int},
			"description": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})

	bookingType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Booking",
		Fields: graphql.Fields{
			"startDate": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"endDate":   &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})

	// ---- Query ----

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"cars": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(carType))),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return c.CarService.GetAll()
				},
			},
			"car": &graphql.Field{
				Type: carType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idVal, _ := p.Args["id"].(string)
					if idVal == "" {
						return nil, nil
					}
					// id уже парсится в CarService как int
					car, getCarErr := c.CarService.GetByIDString(idVal)
					if getCarErr != nil {
						return nil, getCarErr
					}
					if car == nil {
						return nil, nil
					}
					return car, nil
				},
			},
			"comments": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(commentType))),
				Args: graphql.FieldConfigArgument{
					"carId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					carID, _ := p.Args["carId"].(int)
					if carID <= 0 {
						return []models.Comment{}, nil
					}
					return c.CommentService.GetByCarID(carID)
				},
			},
			// myRentals пока не используется на фронте и требует отдельной
			// аутентификации через контекст (пока убрано)
			"carBookings": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(bookingType))),
				Args: graphql.FieldConfigArgument{
					"carId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					carID, _ := p.Args["carId"].(int)
					if carID <= 0 {
						return []map[string]string{}, nil
					}
					bookings, listBookingsErr := c.RentalService.ListCarBookings(uint(carID))
					if listBookingsErr != nil {
						return nil, listBookingsErr
					}
					result := make([]map[string]string, 0, len(bookings))
					for _, b := range bookings {
						result = append(result, map[string]string{
							"startDate": b.StartDate.Format("2006-01-02"),
							"endDate":   b.EndDate.Format("2006-01-02"),
						})
					}
					return result, nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
}

