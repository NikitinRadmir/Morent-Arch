package grpcapi

import (
	"context"

	"morent-backend/internal/grpc/bookingpb/booking/v1"
	"morent-backend/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BookingServiceServer struct {
	bookingpb.UnimplementedBookingServiceServer
	rentals *service.RentalService
}

func NewBookingServiceServer(rentals *service.RentalService) *BookingServiceServer {
	return &BookingServiceServer{rentals: rentals}
}

func (s *BookingServiceServer) CreateBooking(ctx context.Context, in *bookingpb.CreateBookingRequest) (*bookingpb.CreateBookingResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if in.UserId == 0 || in.CarId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and car_id are required")
	}
	if in.StartDate == nil || in.EndDate == nil {
		return nil, status.Error(codes.InvalidArgument, "start_date and end_date are required")
	}

	start := in.StartDate.AsTime()
	end := in.EndDate.AsTime()

	r, err := s.rentals.CreateRental(uint(in.UserId), uint(in.CarId), start, end)
	if err != nil {
		switch err {
		case service.ErrCarNotFound:
			return nil, status.Error(codes.NotFound, err.Error())
		case service.ErrInvalidRentalPeriod:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case service.ErrCarAlreadyBooked:
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &bookingpb.CreateBookingResponse{
		Booking: &bookingpb.Booking{
			Id:         uint64(r.ID),
			UserId:     uint64(in.UserId),
			CarId:      uint64(r.Car.ID),
			StartDate:  timestamppb.New(r.StartDate),
			EndDate:    timestamppb.New(r.EndDate),
			TotalPrice: r.TotalPrice,
			CreatedAt:  timestamppb.New(r.CreatedAt),
		},
	}, nil
}

func (s *BookingServiceServer) GetBookingById(ctx context.Context, in *bookingpb.GetBookingByIdRequest) (*bookingpb.GetBookingByIdResponse, error) {
	if in == nil || in.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	r, err := s.rentals.GetRentalByID(uint(in.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if r == nil {
		return nil, status.Error(codes.NotFound, "booking not found")
	}

	return &bookingpb.GetBookingByIdResponse{
		Booking: &bookingpb.Booking{
			Id:         uint64(r.ID),
			UserId:     uint64(r.UserID),
			CarId:      uint64(r.CarID),
			StartDate:  timestamppb.New(r.StartDate),
			EndDate:    timestamppb.New(r.EndDate),
			TotalPrice: r.TotalPrice,
			CreatedAt:  timestamppb.New(r.CreatedAt),
		},
	}, nil
}

func (s *BookingServiceServer) ListCarBookings(ctx context.Context, in *bookingpb.ListCarBookingsRequest) (*bookingpb.ListCarBookingsResponse, error) {
	if in == nil || in.CarId == 0 {
		return nil, status.Error(codes.InvalidArgument, "car_id is required")
	}

	bookings, err := s.rentals.ListCarBookings(uint(in.CarId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	out := make([]*bookingpb.DateRange, 0, len(bookings))
	for _, b := range bookings {
		out = append(out, &bookingpb.DateRange{
			StartDate: timestamppb.New(b.StartDate),
			EndDate:   timestamppb.New(b.EndDate),
		})
	}

	return &bookingpb.ListCarBookingsResponse{Booked: out}, nil
}

func (s *BookingServiceServer) CancelBooking(ctx context.Context, in *bookingpb.CancelBookingRequest) (*bookingpb.CancelBookingResponse, error) {
	if in == nil || in.Id == 0 || in.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "id and user_id are required")
	}

	cancelled, err := s.rentals.CancelRental(uint(in.Id), uint(in.UserId))
	if err != nil {
		if err == service.ErrForbidden {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !cancelled {
		return nil, status.Error(codes.NotFound, "booking not found")
	}

	return &bookingpb.CancelBookingResponse{Cancelled: true}, nil
}

