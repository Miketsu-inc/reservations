package review

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type Service struct {
	reviewRepo  domain.ReviewRepository
	bookingRepo domain.BookingRepository
}

func NewService(reviewRepo domain.ReviewRepository, bookingRepo domain.BookingRepository) *Service {
	return &Service{
		reviewRepo:  reviewRepo,
		bookingRepo: bookingRepo,
	}
}

func CheckBookingEligibility(booking domain.Booking, participant domain.BookingParticipant, now time.Time) error {
	if booking.Status != types.BookingStatusCompleted || participant.Status != types.BookingStatusCompleted {
		return ErrBookingNotCompleted
	}

	// 15 days after completing the booking
	if now.Sub(booking.ToDate) > 15*24*time.Hour {
		return ErrBookingExpiredForReview
	}

	return nil
}

func CheckBookingMatchesReview(booking domain.Booking, review domain.Review) error {
	// if original booking, it always matches
	if review.BookingId != nil && booking.Id == *review.BookingId {
		return nil
	}

	if booking.MerchantId != review.MerchantId ||
		booking.LocationId != review.LocationId ||
		booking.EmployeeId == nil || review.EmployeeId == nil || *booking.EmployeeId != *review.EmployeeId ||
		booking.ServiceId == nil || review.ServiceId == nil || *booking.ServiceId != *review.ServiceId {
		return ErrBookingNotFound
	}
	return nil
}

func CheckReviewModifiable(review domain.Review) error {
	switch review.Status {
	case "reported":
		return ErrReviewFrozen
	case "removed":
		return ErrReviewRemoved
	default:
		return nil
	}
}

func CheckSubsequentBookingRequirement(existingReview domain.Review, booking domain.Booking, bookingId int) error {
	if existingReview.BookingId != nil && *existingReview.BookingId == bookingId {
		return ErrBookingAlreadyReviewed
	}

	if existingReview.Status != "removed" {
		return ErrReviewAlreadyExists
	}

	anchorDate := existingReview.CreatedAt
	if existingReview.ModeratedAt != nil {
		anchorDate = *existingReview.ModeratedAt
	}

	if !booking.FromDate.After(anchorDate) {
		return ErrSubsequentBookingRequired
	}

	return nil
}

func (s *Service) GetReviewForBooking(ctx context.Context, bookingId int) (*domain.Review, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	review, err := s.reviewRepo.GetReviewForBooking(ctx, userId, bookingId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return review, nil

}

func (s *Service) CreateReview(ctx context.Context, bookingId int, rating int16, comment *string) error {
	userId := jwt.MustGetUserIDFromContext(ctx)
	now := time.Now().UTC()

	participant, err := s.bookingRepo.GetBookingParticipantByUser(ctx, bookingId, userId)
	if err != nil {
		return err
	}

	booking, err := s.bookingRepo.GetBooking(ctx, bookingId)
	if err != nil {
		return err
	}

	if err := CheckBookingEligibility(booking, participant, now); err != nil {
		return err
	}

	existingForBooking, err := s.reviewRepo.GetReviewForBooking(ctx, userId, bookingId)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	if existingForBooking != nil && existingForBooking.BookingId != nil && *existingForBooking.BookingId == bookingId {
		return ErrBookingAlreadyReviewed
	}

	if booking.EmployeeId != nil && booking.ServiceId != nil {
		existingReview, err := s.reviewRepo.GetExistingReviewCombination(ctx, userId, booking.MerchantId, booking.LocationId, *booking.EmployeeId, *booking.ServiceId)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		if existingReview != nil {
			if err := CheckSubsequentBookingRequirement(*existingReview, booking, bookingId); err != nil {
				return err
			}
		}
	}

	review := domain.Review{
		BookingId:   &bookingId,
		UserId:      userId,
		MerchantId:  booking.MerchantId,
		LocationId:  booking.LocationId,
		EmployeeId:  booking.EmployeeId,
		ServiceId:   booking.ServiceId,
		ServiceName: booking.ServiceName,
		Rating:      rating,
		Comment:     comment,
	}

	return s.reviewRepo.NewReview(ctx, review)
}

func (s *Service) UpdateReview(ctx context.Context, reviewId int, bookingId *int, rating int16, comment *string) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	review, err := s.reviewRepo.GetReviewById(ctx, reviewId)
	if err != nil {
		return err
	}

	if review.UserId != userId {
		return ErrUnauthorized
	}

	if err := CheckReviewModifiable(review); err != nil {
		return err
	}

	bookingIdToUpdate := bookingId
	if bookingIdToUpdate == nil {
		bookingIdToUpdate = review.BookingId
	}

	if bookingIdToUpdate == nil {
		return ErrBookingNotFound
	}

	now := time.Now().UTC()
	participant, err := s.bookingRepo.GetBookingParticipantByUser(ctx, *bookingIdToUpdate, userId)
	if err != nil {
		return err
	}

	booking, err := s.bookingRepo.GetBooking(ctx, *bookingIdToUpdate)
	if err != nil {
		return err
	}

	if err := CheckBookingEligibility(booking, participant, now); err != nil {
		return err
	}

	if err := CheckBookingMatchesReview(booking, review); err != nil {
		return err
	}

	err = s.reviewRepo.UpdateReview(ctx, reviewId, userId, bookingId, rating, comment)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteReview(ctx context.Context, reviewId int) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	review, err := s.reviewRepo.GetReviewById(ctx, reviewId)
	if err != nil {
		return err
	}

	if review.UserId != userId {
		return ErrUnauthorized
	}

	if err := CheckReviewModifiable(review); err != nil {
		return err
	}

	err = s.reviewRepo.DeleteReview(ctx, reviewId, userId)
	if err != nil {
		return err
	}

	return nil
}
