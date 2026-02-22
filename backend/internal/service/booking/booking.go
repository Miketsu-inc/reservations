package booking

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bojanz/currency"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/lang"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/service/email"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/teambition/rrule-go"
)

type Service struct {
	bookingRepo  domain.BookingRepository
	catalogRepo  domain.CatalogRepository
	merchantRepo domain.MerchantRepository
	userRepo     domain.UserRepository
	customerRepo domain.CustomerRepository
	mailer       email.Service
}

func NewService(booking domain.BookingRepository, catalog domain.CatalogRepository, merchant domain.MerchantRepository,
	user domain.UserRepository, customer domain.CustomerRepository, mailer email.Service) *Service {
	return &Service{
		bookingRepo:  booking,
		catalogRepo:  catalog,
		merchantRepo: merchant,
		userRepo:     user,
		customerRepo: customer,
		mailer:       mailer,
	}
}

type CreateByCustomerInput struct {
	MerchantName string
	ServiceId    int
	LocationId   int
	TimeStamp    string
	CustomerNote string
	// only present on group bookings
	BookingId *int
}

func (s *Service) CreateByCustomer(ctx context.Context, input CreateByCustomerInput) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, input.MerchantName)
	if err != nil {
		return fmt.Errorf("error while searching merchant by this name: %w", err)
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezoneById(ctx, merchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's timezone: %w", err)
	}

	service, err := s.catalogRepo.GetServiceWithPhasesById(ctx, input.ServiceId, merchantId)
	if err != nil {
		return fmt.Errorf("error while searching service by this id: %w", err)
	}

	bookingSettings, err := s.merchantRepo.GetBookingSettingsByMerchantAndService(ctx, merchantId, service.Id)
	if err != nil {
		return fmt.Errorf("error while getting booking settings for merchant %w", err)
	}

	bookedLocation, err := s.merchantRepo.GetLocationById(ctx, input.LocationId, merchantId)
	if err != nil {
		return fmt.Errorf("error while searching location by this id: %w", err)
	}

	duration := time.Duration(service.TotalDuration) * time.Minute

	timeStamp, err := time.Parse(time.RFC3339, input.TimeStamp)
	if err != nil {
		return fmt.Errorf("timestamp could not be converted to time: %w", err)
	}

	now := time.Now().In(merchantTz)

	if timeStamp.Before(now.Add(time.Duration(bookingSettings.BookingWindowMin) * time.Minute)) {
		return fmt.Errorf("appointment must be booked at least %d minutes in advance", bookingSettings.BookingWindowMin)
	}

	if timeStamp.After(now.AddDate(0, bookingSettings.BookingWindowMax, 0)) {
		return fmt.Errorf("appointment cannot be booked more than %d months in advance", bookingSettings.BookingWindowMax)
	}

	toDate := timeStamp.Add(duration)

	customerId, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("unexpected error during creating customer id: %w", err)
	}

	// prevent null booking price and cost to avoid a lot of headaches
	var price currencyx.Price
	var cost currencyx.Price
	if service.Price == nil || service.Cost == nil {
		curr, err := s.merchantRepo.GetMerchantCurrency(ctx, merchantId)
		if err != nil {
			return fmt.Errorf("error while getting merchant's currency: %w", err)
		}

		zeroAmount, err := currency.NewAmount("0", curr)
		if err != nil {
			return fmt.Errorf("error while creating new amount: %w", err)
		}

		if service.Price != nil {
			price = *service.Price
		} else {
			price = currencyx.Price{Amount: zeroAmount}
		}

		if service.Cost != nil {
			cost = *service.Cost
		} else {
			cost = currencyx.Price{Amount: zeroAmount}
		}
	} else {
		price = *service.Price
		cost = *service.Cost
	}

	bookingId, err := s.bookingRepo.NewBookingByCustomer(ctx, domain.NewCustomerBooking{
		Status:         types.BookingStatusBooked,
		BookingType:    types.BookingTypeAppointment,
		BookingId:      input.BookingId,
		MerchantId:     merchantId,
		ServiceId:      input.ServiceId,
		LocationId:     input.LocationId,
		FromDate:       timeStamp,
		ToDate:         toDate,
		CustomerNote:   &input.CustomerNote,
		PricePerPerson: price,
		CostPerPerson:  cost,
		UserId:         userId,
		CustomerId:     customerId,
		Phases:         service.Phases,
	})
	if err != nil {
		return fmt.Errorf("could not make new booking: %w", err)
	}

	userInfo, err := s.userRepo.GetUserById(ctx, userId)
	if err != nil {
		return fmt.Errorf("could not get email for the user: %w", err)
	}

	toDateMerchantTz := toDate.In(merchantTz)
	fromDateMerchantTz := timeStamp.In(merchantTz)

	emailData := email.BookingConfirmationData{
		Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:        fromDateMerchantTz.Format("Monday, January 2"),
		Location:    bookedLocation.FormattedLocation,
		ServiceName: service.Name,
		TimeZone:    merchantTz.String(),
		ModifyLink:  "http://reservations.local:3000/m/" + input.MerchantName + "/cancel/" + strconv.Itoa(bookingId),
	}

	lang := lang.LangFromContext(ctx)

	err = s.mailer.BookingConfirmation(ctx, lang, userInfo.Email, emailData)
	if err != nil {
		return fmt.Errorf("could not send confirmation email for the booking: %w", err)
	}

	hoursUntilBooking := time.Until(fromDateMerchantTz).Hours()

	if hoursUntilBooking >= 24 {

		reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)
		email_id, err := s.mailer.BookingReminder(ctx, lang, userInfo.Email, emailData, reminderDate)
		if err != nil {
			return fmt.Errorf("could not schedule reminder email: %w", err)
		}

		if email_id != "" { //check because return "" when email sending is off
			err = s.bookingRepo.UpdateEmailIdForBooking(ctx, bookingId, email_id, customerId)
			if err != nil {
				return fmt.Errorf("failed to update email ID: %w", err)
			}
		}
	}

	return nil
}

type CancelByCustomerInput struct {
	BookingId    int
	MerchantName string
}

func (s *Service) CancelByCustomer(ctx context.Context, bookingId int, input CancelByCustomerInput) error {
	if bookingId != input.BookingId {
		return fmt.Errorf("invalid booking id")
	}

	userId := jwt.MustGetUserIDFromContext(ctx)

	merchantId, err := s.merchantRepo.GetMerchantIdByUrlName(ctx, input.MerchantName)
	if err != nil {
		return fmt.Errorf("error while searching merchant by this name: %s", err.Error())
	}

	customerId, err := s.customerRepo.GetCustomerIdByUserIdAndMerchantId(ctx, merchantId, userId)
	if err != nil {
		return fmt.Errorf("error while getting customer id: %s", err.Error())
	}

	//TODO: write seperate query for getting only fromDate and cancel deadline
	emailData, err := s.bookingRepo.GetBookingDataForEmail(ctx, bookingId)
	if err != nil {
		return fmt.Errorf("error while retrieving data for email sending: %s", err.Error())
	}

	latestCancelTime := emailData.FromDate.Add(-time.Duration(emailData.CancelDeadline) * time.Minute)

	if time.Now().After(latestCancelTime) {
		return fmt.Errorf("it's too late to cancel this appointments")
	}

	emailId, err := s.bookingRepo.CancelBookingByCustomer(ctx, customerId, bookingId)
	if err != nil {
		return fmt.Errorf("error while cancelling the booking by user: %s", err.Error())
	}

	if emailId != uuid.Nil {
		err = s.mailer.Cancel(emailId.String())
		if err != nil {
			return fmt.Errorf("error while cancelling the scheduled email for the booking: %s", err.Error())
		}
	}

	return nil
}

func (s *Service) GetByCustomer(ctx context.Context, bookingId int) (domain.PublicBooking, error) {
	publicBooking, err := s.bookingRepo.GetPublicBooking(ctx, bookingId)
	if err != nil {
		return domain.PublicBooking{}, err
	}

	return publicBooking, nil
}

type CreateByMerchantInput struct {
	Customers    []CustomerInput
	ServiceId    int
	TimeStamp    string
	MerchantNote *string
	IsRecurring  bool
	Rrule        *RecurringRuleInput
}

type CustomerInput struct {
	CustomerId  *uuid.UUID
	FirstName   *string
	LastName    *string
	Email       *string
	PhoneNumber *string
}

type RecurringRuleInput struct {
	Frequency string
	Interval  int
	Weekdays  []string
	Until     string
}

func (s *Service) CreateByMerchant(ctx context.Context, input CreateByMerchantInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	service, err := s.catalogRepo.GetServiceWithPhasesById(ctx, input.ServiceId, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while searching service by this id: %s", err.Error())
	}

	if service.BookingType == types.BookingTypeAppointment {
		if len(input.Customers) > 1 {
			return fmt.Errorf("appointments cannot have more than 1 customer")
		}
	}

	if service.BookingType != types.BookingTypeAppointment {
		if len(input.Customers) > service.MaxParticipants {
			return fmt.Errorf("customer count (%d) exceeds class limit of %d", len(input.Customers), service.MaxParticipants)
		}
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezoneById(ctx, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	bookedLocation, err := s.merchantRepo.GetLocationById(ctx, employee.LocationId, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while searching location by this id: %s", err.Error())
	}

	// TODO: this should be a separate function
	// prevent null booking price and cost to avoid a lot of headaches
	var price currencyx.Price
	var cost currencyx.Price
	if service.Price == nil || service.Cost == nil {
		curr, err := s.merchantRepo.GetMerchantCurrency(ctx, employee.MerchantId)
		if err != nil {
			return fmt.Errorf("error while getting merchant's currency: %s", err.Error())
		}

		zeroAmount, err := currency.NewAmount("0", curr)
		if err != nil {
			return fmt.Errorf("error while creating new amount: %s", err.Error())
		}

		if service.Price != nil {
			price = *service.Price
		} else {
			price = currencyx.Price{Amount: zeroAmount}
		}

		if service.Cost != nil {
			cost = *service.Cost
		} else {
			cost = currencyx.Price{Amount: zeroAmount}
		}
	} else {
		price = *service.Price
		cost = *service.Cost
	}

	timeStamp, err := time.Parse(time.RFC3339, input.TimeStamp)
	if err != nil {
		return fmt.Errorf("timestamp could not be converted to time: %s", err.Error())
	}

	duration := time.Duration(service.TotalDuration) * time.Minute

	fromDate := timeStamp.Truncate(time.Second)
	toDate := timeStamp.Add(duration)

	var participantIds []*uuid.UUID
	var emailsToSend []*uuid.UUID
	isWalkIn := false

	//maybe check if the customers without an id have first and last name
	if len(input.Customers) == 0 {
		isWalkIn = true
	} else {
		for _, customer := range input.Customers {
			if customer.CustomerId != nil {
				participantIds = append(participantIds, customer.CustomerId)
				emailsToSend = append(emailsToSend, customer.CustomerId)
			} else {
				if customer.FirstName != nil && customer.LastName != nil {
					newId, err := uuid.NewV7()
					if err != nil {
						return fmt.Errorf("unexpected error during creating customer id: %s", err.Error())
					}

					customerId := &newId

					if err := s.customerRepo.NewCustomer(ctx, employee.MerchantId, domain.Customer{
						Id:          *customerId,
						FirstName:   customer.FirstName,
						LastName:    customer.LastName,
						Email:       customer.Email,
						PhoneNumber: customer.PhoneNumber,
						Birthday:    nil,
						Note:        nil,
					}); err != nil {
						return fmt.Errorf("unexpected error inserting customer for merchant: %s", err.Error())
					}
					participantIds = append(participantIds, customerId)
					if customer.Email != nil {
						emailsToSend = append(emailsToSend, customerId)
					}
				}
			}
		}
	}

	var bookingId int

	if input.IsRecurring && input.Rrule != nil {
		var freq rrule.Frequency

		switch strings.ToUpper(input.Rrule.Frequency) {
		case "DAILY":
			freq = rrule.DAILY
		case "WEEKLY":
			freq = rrule.WEEKLY
		case "MONTHLY":
			freq = rrule.MONTHLY
		default:
			return fmt.Errorf("recurring rule frequency is invalid")
		}

		untilTimeStamp, err := time.Parse(time.RFC3339, input.Rrule.Until)
		if err != nil {
			return fmt.Errorf("until timestamp could not be converted to time: %s", err.Error())
		}

		var weekdays []rrule.Weekday

		for _, wkd := range input.Rrule.Weekdays {
			switch strings.ToUpper(wkd) {
			case rrule.MO.String():
				weekdays = append(weekdays, rrule.MO)
			case rrule.TU.String():
				weekdays = append(weekdays, rrule.TU)
			case rrule.WE.String():
				weekdays = append(weekdays, rrule.WE)
			case rrule.TH.String():
				weekdays = append(weekdays, rrule.TH)
			case rrule.FR.String():
				weekdays = append(weekdays, rrule.FR)
			case rrule.SA.String():
				weekdays = append(weekdays, rrule.SA)
			case rrule.SU.String():
				weekdays = append(weekdays, rrule.SU)
			default:
				return fmt.Errorf("incorrect weekday")
			}
		}

		rrule, err := rrule.NewRRule(rrule.ROption{
			Freq:      freq,
			Dtstart:   fromDate,
			Interval:  input.Rrule.Interval,
			Byweekday: weekdays,
			Until:     untilTimeStamp,
		})
		if err != nil {
			return fmt.Errorf("error while creating rrule: %s", err.Error())
		}

		// recurring bookings have to be stored in local time and converted to UTC during generation
		fromDate = timeStamp.In(merchantTz)

		series, err := s.bookingRepo.NewBookingSeries(ctx, domain.NewBookingSeries{
			BookingType:     service.BookingType,
			MerchantId:      employee.MerchantId,
			EmployeeId:      employee.Id,
			ServiceId:       service.Id,
			LocationId:      bookedLocation.Id,
			Rrule:           rrule.String(),
			Dstart:          fromDate,
			Timezone:        merchantTz,
			PricePerPerson:  price,
			CostPerPerson:   cost,
			MinParticipants: service.MinParticipants,
			MaxParticipants: service.MaxParticipants,
			Participants:    participantIds,
		})
		if err != nil {
			return fmt.Errorf("error while creating new booking series: %s", err.Error())
		}

		bookingId, err = s.generateRecurringBookings(ctx, series, service.Phases)
		if err != nil {
			return fmt.Errorf("error while generating recurring bookings: %s", err.Error())
		}
	} else {
		bookingId, err = s.bookingRepo.NewBookingByMerchant(ctx, domain.NewMerchantBooking{
			Status:          types.BookingStatusBooked,
			BookingType:     service.BookingType,
			MerchantId:      employee.MerchantId,
			ServiceId:       service.Id,
			LocationId:      bookedLocation.Id,
			FromDate:        fromDate,
			ToDate:          toDate,
			MerchantNote:    input.MerchantNote,
			PricePerPerson:  price,
			CostPerPerson:   cost,
			MinParticipants: service.MinParticipants,
			MaxParticipants: service.MaxParticipants,
			Participants:    participantIds,
			Phases:          service.Phases,
		})
		if err != nil {
			return fmt.Errorf("error while creating new booking: %s", err.Error())
		}
	}

	if !isWalkIn {
		urlName, err := s.merchantRepo.GetMerchantUrlNameById(ctx, employee.MerchantId)
		if err != nil {
			return fmt.Errorf("error while getting merchant's url name: %s", err.Error())
		}

		toDateMerchantTz := toDate.In(merchantTz)
		fromDateMerchantTz := timeStamp.In(merchantTz)

		emailData := email.BookingConfirmationData{
			Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
			Date:        fromDateMerchantTz.Format("Monday, January 2"),
			Location:    bookedLocation.FormattedLocation,
			ServiceName: service.Name,
			TimeZone:    merchantTz.String(),
			ModifyLink:  fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", urlName, bookingId),
		}

		lang := lang.LangFromContext(ctx)
		hoursUntilBooking := time.Until(fromDateMerchantTz).Hours()

		for _, customerId := range emailsToSend {
			customerEmail, err := s.customerRepo.GetCustomerEmailById(ctx, employee.MerchantId, *customerId)
			if err != nil {
				return fmt.Errorf("error while getting customer's email: %s", err.Error())
			}

			if customerEmail != nil {

				err = s.mailer.BookingConfirmation(ctx, lang, *customerEmail, emailData)
				if err != nil {
					return fmt.Errorf("could not send confirmation email for the booking: %s", err.Error())
				}

				if hoursUntilBooking >= 24 {

					reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)
					email_id, err := s.mailer.BookingReminder(ctx, lang, *customerEmail, emailData, reminderDate)
					if err != nil {
						return fmt.Errorf("could not schedule reminder email: %s", err.Error())
					}

					if email_id != "" { //check because return "" when email sending is off
						err = s.bookingRepo.UpdateEmailIdForBooking(ctx, bookingId, email_id, *customerId)
						if err != nil {
							return fmt.Errorf("failed to update email ID: %s", err.Error())
						}
					}
				}
			}
		}
	}

	return nil
}

type UpdateByMerchantInput struct {
	MerchantNote string
	FromDate     time.Time
	ToDate       time.Time
}

// TODO: updating the participants as well
func (s *Service) UpdateByMerchant(ctx context.Context, bookingId int, input UpdateByMerchantInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	oldEmailData, err := s.bookingRepo.GetBookingDataForEmail(ctx, bookingId)
	if err != nil {
		return fmt.Errorf("error while retrieving data for email sending: %s", err.Error())
	}

	fromDateOffset := input.FromDate.Sub(oldEmailData.FromDate)
	toDateOffset := input.ToDate.Sub(oldEmailData.FromDate)

	if fromDateOffset != toDateOffset {
		return fmt.Errorf("invalid from and to date supplied")
	}

	if err := s.bookingRepo.UpdateBookingData(ctx, employee.MerchantId, bookingId, input.MerchantNote, fromDateOffset); err != nil {
		return err
	}

	merchantTz, err := s.merchantRepo.GetMerchantTimezoneById(ctx, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	toDateMerchantTz := input.ToDate.In(merchantTz)
	fromDateMerchantTz := input.FromDate.In(merchantTz)
	oldToDateMerchantTz := oldEmailData.ToDate.In(merchantTz)
	oldFromDateMerchantTz := oldEmailData.FromDate.In(merchantTz)

	oldFormattedDate := oldEmailData.FromDate.Format("Monday, January 2")
	oldFormattedTime := oldFromDateMerchantTz.Format("15:04") + " - " + oldToDateMerchantTz.Format("15:04")
	formattedTime := fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04")
	formattedDate := input.FromDate.Format("Monday, January 2")
	modifyLink := fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", oldEmailData.MerchantName, bookingId)

	for _, participant := range oldEmailData.Participants {
		// email is nil if it's a walk-in or if a customer doesnt have an email
		if participant.Email != nil {
			lang := lang.LangFromContext(ctx)

			err = s.mailer.BookingModification(ctx, lang, *participant.Email, email.BookingModificationData{
				Time:        formattedTime,
				Date:        formattedDate,
				Location:    oldEmailData.FormattedLocation,
				ServiceName: oldEmailData.ServiceName,
				TimeZone:    merchantTz.String(),
				ModifyLink:  modifyLink,
				OldTime:     oldFormattedTime,
				OldDate:     oldFormattedDate,
			})
			if err != nil {
				return err
			}

			hoursUntilBooking := time.Until(fromDateMerchantTz).Hours()

			if participant.EmailId != uuid.Nil {
				// Always cancel the old email — content might be outdated
				err := s.mailer.Cancel(participant.EmailId.String())
				if err != nil {
					return fmt.Errorf("could not cancel old reminder email: %s", err.Error())
				}
			}

			// Only schedule new one if new time is valid
			if hoursUntilBooking >= 24 {
				reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)

				new_email_id, err := s.mailer.BookingReminder(ctx, lang, *participant.Email, email.BookingConfirmationData{
					Time:        formattedTime,
					Date:        formattedDate,
					Location:    oldEmailData.FormattedLocation,
					ServiceName: oldEmailData.ServiceName,
					TimeZone:    merchantTz.String(),
					ModifyLink:  modifyLink,
				}, reminderDate)
				if err != nil {
					return fmt.Errorf("could not schedule reminder email: %s", err.Error())
				}

				if new_email_id != "" { //check because return "" when email sending is off
					err = s.bookingRepo.UpdateEmailIdForBooking(ctx, bookingId, new_email_id, participant.CustomerId)
					if err != nil {
						return fmt.Errorf("failed to update email ID: %s", err.Error())
					}
				}
			}
		}
	}

	return nil
}

type CancelByMerchantInput struct {
	CancellationReason string
}

func (s *Service) CancelByMerchant(ctx context.Context, bookingId int, input CancelByMerchantInput) error {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	merchantTz, err := s.merchantRepo.GetMerchantTimezoneById(ctx, employee.MerchantId)
	if err != nil {
		return fmt.Errorf("error while getting merchant's timezone: %s", err.Error())
	}

	emailData, err := s.bookingRepo.GetBookingDataForEmail(ctx, bookingId)
	if err != nil {
		return fmt.Errorf("error while retrieving data for email sending: %s", err.Error())
	}

	if emailData.FromDate.Before(time.Now().UTC()) {
		return fmt.Errorf("you cannot cancel past bookings")
	}

	toDateMerchantTz := emailData.ToDate.In(merchantTz)
	fromDateMerchantTz := emailData.FromDate.In(merchantTz)
	formattedTime := fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04")
	formattedDate := fromDateMerchantTz.Format("Monday, January 2")
	newBookingLink := "http://reservations.local:3000/m/" + emailData.MerchantName

	err = s.bookingRepo.CancelBookingByMerchant(ctx, employee.MerchantId, bookingId, input.CancellationReason)
	if err != nil {
		return fmt.Errorf("error while cancelling booking by merchant: %s", err.Error())
	}

	for _, participant := range emailData.Participants {

		// email is nil if it's a walk-in
		if participant.Email != nil {
			lang := lang.LangFromContext(ctx)

			err = s.mailer.BookingCancellation(ctx, lang, *participant.Email, email.BookingCancellationData{
				Time:           formattedTime,
				Date:           formattedDate,
				Location:       emailData.FormattedLocation,
				ServiceName:    emailData.ServiceName,
				TimeZone:       merchantTz.String(),
				Reason:         input.CancellationReason,
				NewBookingLink: newBookingLink,
			})
			if err != nil {
				return fmt.Errorf("error while sending cancellation email: %s", err.Error())
			}

			if participant.EmailId != uuid.Nil {
				err = s.mailer.Cancel(participant.EmailId.String())
				if err != nil {
					return fmt.Errorf("error while cancelling the scheduled email for the booking: %s", err.Error())
				}
			}
		}
	}

	return nil
}

func (s *Service) GetCalendarEvents(ctx context.Context, start string, end string) (domain.CalendarEvents, error) {
	employee := jwt.MustGetEmployeeFromContext(ctx)

	bookings, err := s.bookingRepo.GetCalendarEventsByMerchant(ctx, employee.MerchantId, start, end)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	return bookings, nil
}
