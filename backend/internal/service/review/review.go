package review

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/permission"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/cursor"
)

type Service struct {
	reviewRepo   domain.ReviewRepository
	bookingRepo  domain.BookingRepository
	locationRepo domain.LocationRepository
}

func NewService(reviewRepo domain.ReviewRepository, bookingRepo domain.BookingRepository, locationRepo domain.LocationRepository) *Service {
	return &Service{
		reviewRepo:   reviewRepo,
		bookingRepo:  bookingRepo,
		locationRepo: locationRepo,
	}
}

func CheckBookingEligibility(booking domain.Booking, participant domain.BookingParticipant, now time.Time) error {
	if booking.Status != types.BookingStatusCompleted || participant.Status != types.BookingStatusCompleted {
		return ErrBookingNotCompleted
	}

	// 15 days after completing the booking
	if now.Sub(booking.ToDate) > 30*24*time.Hour {
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

func (s *Service) GetReviewStats(ctx context.Context, merchantId uuid.UUID, locationId *int, employeeId *int) (domain.ReviewStats, error) {
	if locationId == nil {
		defaultLoc, err := s.locationRepo.GetDefaultLocation(ctx, merchantId)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return domain.ReviewStats{}, fmt.Errorf("GetDefaultLocation: %w", err)
		}
		if err == nil {
			locationId = &defaultLoc.Id
		}
	}

	stats, err := s.reviewRepo.GetReviewStats(ctx, merchantId, locationId, employeeId)
	if err != nil {
		return domain.ReviewStats{}, err
	}

	stats.ServiceCounts, err = s.reviewRepo.GetReviewCountsByService(ctx, merchantId, locationId, employeeId)
	if err != nil {
		return domain.ReviewStats{}, err
	}

	return stats, nil
}

func (s *Service) GetReviewStatsByMerchant(ctx context.Context) (domain.ReviewStats, error) {
	actor := actor.MustGetFromContext(ctx)

	return s.GetReviewStats(ctx, actor.MerchantId, nil, permission.For(actor).ViewFilter(permission.Reviews))
}

type reviewCursor struct {
	Offset int `json:"offset"`
}

type GetReviewsInput struct {
	MerchantId uuid.UUID
	LocationId *int
	ReviewQueryParams
}

type GetReviewsResult struct {
	Reviews     []domain.ReviewDetails
	NextCursor  *string
	HasNextPage bool
}

func (s *Service) GetReviews(ctx context.Context, input GetReviewsInput) (GetReviewsResult, error) {
	decodedCursor, err := cursor.Decode[reviewCursor](input.Cursor)
	if err != nil {
		return GetReviewsResult{}, fmt.Errorf("decode review cursor: %w", err)
	}

	locationId := input.LocationId
	if locationId == nil {
		defaultLoc, err := s.locationRepo.GetDefaultLocation(ctx, input.MerchantId)
		if err != nil {
			return GetReviewsResult{}, fmt.Errorf("GetDefaultLocation: %w", err)
		}
		locationId = &defaultLoc.Id
	}

	filter := domain.ReviewFilter{
		MerchantId:  input.MerchantId,
		LocationId:  locationId,
		Rating:      input.Rating,
		EmployeeIds: input.EmployeeIds,
		ServiceIds:  input.ServiceIds,
		HasReply:    input.HasReply,
		Statuses:    []string{"published", "reported"},
		Sort:        input.Sort,
		Limit:       input.PageSize + 1,
		Offset:      decodedCursor.Offset,
	}

	reviews, err := s.reviewRepo.GetReviews(ctx, filter)
	if err != nil {
		return GetReviewsResult{}, err
	}

	hasNextPage := len(reviews) > input.PageSize

	var nextCursor *string
	if hasNextPage {
		encoded, err := cursor.Encode(reviewCursor{
			Offset: decodedCursor.Offset + input.PageSize,
		})
		if err != nil {
			return GetReviewsResult{}, fmt.Errorf("encode cursor: %w", err)
		}
		nextCursor = &encoded
		reviews = reviews[:input.PageSize]
	}

	return GetReviewsResult{
		Reviews:     reviews,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}

func (s *Service) GetReviewsByMerchant(ctx context.Context, params ReviewQueryParams) (GetReviewsResult, error) {
	actor := actor.MustGetFromContext(ctx)

	if employeeId := permission.For(actor).ViewFilter(permission.Reviews); employeeId != nil {
		for _, id := range params.EmployeeIds {
			if id != *employeeId {
				return GetReviewsResult{}, domain.ErrEmployeeAccessDenied
			}
		}

		params.EmployeeIds = []int{*employeeId}
	}

	return s.GetReviews(ctx, GetReviewsInput{
		MerchantId:        actor.MerchantId,
		ReviewQueryParams: params,
	})
}

func (s *Service) checkReviewEditable(ctx context.Context, actor actor.EmployeeContext, reviewId int) error {
	review, err := s.reviewRepo.GetReviewById(ctx, reviewId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrReviewNotFound
		}

		return err
	}

	if review.MerchantId != actor.MerchantId {
		return ErrReviewNotFound
	}

	can := permission.For(actor)

	var canEdit bool

	// Reviews whose employee was deleted are only editable by those who can edit everyone's
	if review.EmployeeId != nil {
		canEdit = can.Edit(permission.Reviews, *review.EmployeeId)
	} else {
		canEdit = can.EditAll(permission.Reviews)
	}

	if !canEdit {
		return ErrReviewNotFound
	}

	return nil
}

func (s *Service) ReplyReview(ctx context.Context, reviewId int, comment string) error {
	actor := actor.MustGetFromContext(ctx)

	if err := s.checkReviewEditable(ctx, actor, reviewId); err != nil {
		return err
	}

	return s.reviewRepo.UpsertReply(ctx, actor.MerchantId, reviewId, actor.EmployeeId, comment)
}

func (s *Service) DeleteReviewReply(ctx context.Context, reviewId int) error {
	actor := actor.MustGetFromContext(ctx)

	if err := s.checkReviewEditable(ctx, actor, reviewId); err != nil {
		return err
	}

	return s.reviewRepo.DeleteReply(ctx, actor.MerchantId, reviewId)
}

func (s *Service) ReportReview(ctx context.Context, reviewId int, reason string) error {
	actor := actor.MustGetFromContext(ctx)

	if err := s.checkReviewEditable(ctx, actor, reviewId); err != nil {
		return err
	}

	return s.reviewRepo.ReportReview(ctx, reviewId, actor.MerchantId, reason)
}
