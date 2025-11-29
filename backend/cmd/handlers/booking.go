package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bojanz/currency"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/lang"
	"github.com/miketsu-inc/reservations/backend/cmd/types/booking"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/email"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Booking struct {
	Postgresdb database.PostgreSQL
}

func (a *Booking) CreateByCustomer(w http.ResponseWriter, r *http.Request) {
	type BookingData struct {
		MerchantName string `json:"merchant_name" validate:"required"`
		ServiceId    int    `json:"service_id" validate:"required"`
		LocationId   int    `json:"location_id" validate:"required"`
		TimeStamp    string `json:"timeStamp" validate:"required"`
		CustomerNote string `json:"customer_note"`
	}
	var bookData BookingData

	if err := validate.ParseStruct(r, &bookData); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.MustGetUserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByUrlName(r.Context(), bookData.MerchantName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching merchant by this name: %s", err.Error()))
		return
	}

	merchantTz, err := a.Postgresdb.GetMerchantTimezoneById(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	service, err := a.Postgresdb.GetServiceWithPhasesById(r.Context(), bookData.ServiceId, merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching service by this id: %s", err.Error()))
		return
	}

	bookingSettings, err := a.Postgresdb.GetBookingSettingsByMerchantAndService(r.Context(), merchantId, service.Id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting booking settings for merchant %s", err.Error()))
		return
	}

	bookedLocation, err := a.Postgresdb.GetLocationById(r.Context(), bookData.LocationId, merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching location by this id: %s", err.Error()))
		return
	}

	duration := time.Duration(service.TotalDuration) * time.Minute

	timeStamp, err := time.Parse(time.RFC3339, bookData.TimeStamp)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("timestamp could not be converted to time: %s", err.Error()))
		return
	}

	now := time.Now().In(merchantTz)

	if timeStamp.Before(now.Add(time.Duration(bookingSettings.BookingWindowMin) * time.Minute)) {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("appointment must be booked at least %d minutes in advance", bookingSettings.BookingWindowMin))
		return
	}

	if timeStamp.After(now.AddDate(0, bookingSettings.BookingWindowMax, 0)) {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("appointment cannot be booked more than %d months in advance", bookingSettings.BookingWindowMax))
		return
	}

	toDate := timeStamp.Add(duration)

	customerId, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating customer id: %s", err.Error()))
		return
	}

	// prevent null booking price and cost to avoid a lot of headaches
	var price currencyx.Price
	var cost currencyx.Price
	if service.Price == nil || service.Cost == nil {
		curr, err := a.Postgresdb.GetMerchantCurrency(r.Context(), merchantId)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while getting merchant's currency: %s", err.Error()))
			return
		}

		zeroAmount, err := currency.NewAmount("0", curr)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while creating new amount: %s", err.Error()))
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

	bookingId, err := a.Postgresdb.NewBookingByCustomer(r.Context(), database.NewCustomerBooking{
		Status:         booking.Booked,
		BookingType:    booking.Appointment,
		MerchantId:     merchantId,
		ServiceId:      bookData.ServiceId,
		LocationId:     bookData.LocationId,
		FromDate:       timeStamp,
		ToDate:         toDate,
		CustomerNote:   &bookData.CustomerNote,
		PricePerPerson: price,
		CostPerPerson:  cost,
		UserId:         userID,
		CustomerId:     customerId,
		Phases:         service.Phases,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not make new booking: %s", err.Error()))
		return
	}

	userInfo, err := a.Postgresdb.GetUserById(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not get email for the user: %s", err.Error()))
		return
	}

	toDateMerchantTz := toDate.In(merchantTz)
	fromDateMerchantTz := timeStamp.In(merchantTz)

	emailData := email.BookingConfirmationData{
		Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:        fromDateMerchantTz.Format("Monday, January 2"),
		Location:    bookedLocation.FormattedLocation,
		ServiceName: service.Name,
		TimeZone:    merchantTz.String(),
		ModifyLink:  "http://reservations.local:3000/m/" + bookData.MerchantName + "/cancel/" + strconv.Itoa(bookingId),
	}

	lang := lang.LangFromContext(r.Context())

	err = email.BookingConfirmation(r.Context(), lang, userInfo.Email, emailData)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not send confirmation email for the booking: %s", err.Error()))
		return
	}

	hoursUntilBooking := time.Until(fromDateMerchantTz).Hours()

	if hoursUntilBooking >= 24 {

		reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)
		email_id, err := email.BookingReminder(r.Context(), lang, userInfo.Email, emailData, reminderDate)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not schedule reminder email: %s", err.Error()))
			return
		}

		if email_id != "" { //check because return "" when email sending is off
			err = a.Postgresdb.UpdateEmailIdForBooking(r.Context(), bookingId, email_id)
			if err != nil {
				httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to update email ID: %s", err.Error()))
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *Booking) GetCalendarEvents(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	bookings, err := a.Postgresdb.GetCalendarEventsByMerchant(r.Context(), employee.MerchantId, start, end)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	httputil.Success(w, http.StatusOK, bookings)
}

func (a *Booking) CancelBookingByMerchant(w http.ResponseWriter, r *http.Request) {
	type cancelBookingData struct {
		CancellationReason string `json:"cancellation_reason"`
	}

	urlId := chi.URLParam(r, "id")

	bookingId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting booking id to int: %s", err.Error()))
		return
	}

	var cancelData cancelBookingData

	if err := validate.ParseStruct(r, &cancelData); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	merchantTz, err := a.Postgresdb.GetMerchantTimezoneById(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	emailData, err := a.Postgresdb.GetBookingDataForEmail(r.Context(), bookingId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving data for email sending: %s", err.Error()))
		return
	}

	if emailData.FromDate.Before(time.Now().UTC()) {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("you cannot cancel past bookings"))
		return
	}

	toDateMerchantTz := emailData.ToDate.In(merchantTz)
	fromDateMerchantTz := emailData.FromDate.In(merchantTz)

	err = a.Postgresdb.CancelBookingByMerchant(r.Context(), employee.MerchantId, bookingId, cancelData.CancellationReason)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling booking by merchant: %s", err.Error()))
		return
	}

	lang := lang.LangFromContext(r.Context())

	err = email.BookingCancellation(r.Context(), lang, emailData.CustomerEmail, email.BookingCancellationData{
		Time:           fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:           fromDateMerchantTz.Format("Monday, January 2"),
		Location:       emailData.FormattedLocation,
		ServiceName:    emailData.ServiceName,
		TimeZone:       merchantTz.String(),
		Reason:         cancelData.CancellationReason,
		NewBookingLink: "http://reservations.local:3000/m/" + emailData.MerchantName,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while sending cancellation email: %s", err.Error()))
		return
	}

	if emailData.EmailId != uuid.Nil {
		err = email.Cancel(emailData.EmailId.String())
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling the scheduled email for the booking: %s", err.Error()))
			return
		}
	}
}

func (a *Booking) UpdateBookingData(w http.ResponseWriter, r *http.Request) {
	// validate:"required" on MerchantNote would fail
	// if an empty string arrives as a note
	type bookingData struct {
		MerchantNote string `json:"merchant_note"`
		FromDate     string `json:"from_date" validate:"required"`
		ToDate       string `json:"to_date" validate:"required"`
	}

	var bookData bookingData

	if err := validate.ParseStruct(r, &bookData); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlId := chi.URLParam(r, "id")

	bookingId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting booking id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	fromDate, err := time.Parse(time.RFC3339, bookData.FromDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("from_date could not be converted to time: %s", err.Error()))
		return
	}

	toDate, err := time.Parse(time.RFC3339, bookData.ToDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("to_date could not be converted to time: %s", err.Error()))
		return
	}

	oldEmailData, err := a.Postgresdb.GetBookingDataForEmail(r.Context(), bookingId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving data for email sending: %s", err.Error()))
		return
	}

	fromDateOffset := fromDate.Sub(oldEmailData.FromDate)
	toDateOffset := fromDate.Sub(oldEmailData.FromDate)

	if fromDateOffset != toDateOffset {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid from and to date supplied"))
		return
	}

	if err := a.Postgresdb.UpdateBookingData(r.Context(), employee.MerchantId, bookingId, bookData.MerchantNote, fromDateOffset); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	merchantTz, err := a.Postgresdb.GetMerchantTimezoneById(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	toDateMerchantTz := toDate.In(merchantTz)
	fromDateMerchantTz := fromDate.In(merchantTz)
	oldToDateMerchantTz := oldEmailData.ToDate.In(merchantTz)
	oldFromDateMerchantTz := oldEmailData.FromDate.In(merchantTz)

	lang := lang.LangFromContext(r.Context())

	err = email.BookingModification(r.Context(), lang, oldEmailData.CustomerEmail, email.BookingModificationData{
		Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:        fromDate.Format("Monday, January 2"),
		Location:    oldEmailData.FormattedLocation,
		ServiceName: oldEmailData.ServiceName,
		TimeZone:    employee.MerchantId.String(),
		ModifyLink:  fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", oldEmailData.MerchantName, bookingId),
		OldTime:     oldFromDateMerchantTz.Format("15:04") + " - " + oldToDateMerchantTz.Format("15:04"),
		OldDate:     oldEmailData.FromDate.Format("Monday, January 2"),
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	hoursUntilBooking := time.Until(fromDateMerchantTz).Hours()

	if oldEmailData.EmailId != uuid.Nil {
		// Always cancel the old email — content might be outdated
		err := email.Cancel(oldEmailData.EmailId.String())
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not cancel old reminder email: %s", err.Error()))
			return
		}
	}

	// Only schedule new one if new time is valid
	if hoursUntilBooking >= 24 {
		reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)

		email_id, err := email.BookingReminder(r.Context(), lang, oldEmailData.CustomerEmail, email.BookingConfirmationData{
			Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
			Date:        fromDateMerchantTz.Format("Monday, January 2"),
			Location:    oldEmailData.FormattedLocation,
			ServiceName: oldEmailData.ServiceName,
			TimeZone:    merchantTz.String(),
			ModifyLink:  fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", oldEmailData.MerchantName, bookingId),
		}, reminderDate)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not schedule reminder email: %s", err.Error()))
			return
		}

		if email_id != "" { //check because return "" when email sending is off
			err = a.Postgresdb.UpdateEmailIdForBooking(r.Context(), bookingId, email_id)
			if err != nil {
				httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to update email ID: %s", err.Error()))
				return
			}
		}
	}

}

func (a *Booking) GetPublicBookingData(w http.ResponseWriter, r *http.Request) {
	urlId := chi.URLParam(r, "id")

	bookingId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting booking id to int: %s", err.Error()))
		return
	}

	bookingInfo, err := a.Postgresdb.GetPublicBookingInfo(r.Context(), bookingId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving public data for booking: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, bookingInfo)
}

func (a *Booking) CancelBookingByCustomer(w http.ResponseWriter, r *http.Request) {
	type cancelData struct {
		BookingId    int    `json:"booking_id" validate:"required"`
		MerchantName string `json:"merchant_name" validate:"required"`
	}

	var data cancelData

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlId := chi.URLParam(r, "id")

	bookingId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting booking id to int: %s", err.Error()))
		return
	}

	userId := jwt.MustGetUserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByUrlName(r.Context(), data.MerchantName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching merchant by this name: %s", err.Error()))
		return
	}

	customerId, err := a.Postgresdb.GetCustomerIdByUserIdAndMerchantId(r.Context(), merchantId, userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while getting customer id: %s", err.Error()))
	}

	//TODO: write seperate query for getting only fromDate and cancel deadline
	emailData, err := a.Postgresdb.GetBookingDataForEmail(r.Context(), bookingId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving data for email sending: %s", err.Error()))
		return
	}

	latestCancelTime := emailData.FromDate.Add(-time.Duration(emailData.CancelDeadline) * time.Minute)

	if time.Now().After(latestCancelTime) {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("it's too late to cancel this appointments"))
		return
	}

	emailId, err := a.Postgresdb.CancelBookingByCustomer(r.Context(), customerId, bookingId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling the booking by user: %s", err.Error()))
		return
	}

	if emailId != uuid.Nil {
		err = email.Cancel(emailId.String())
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling the scheduled email for the booking: %s", err.Error()))
			return
		}
	}
}
