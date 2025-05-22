package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
"github.com/miketsu-inc/reservations/backend/pkg/email"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Appointment struct {
	Postgresdb database.PostgreSQL
}

func (a *Appointment) Create(w http.ResponseWriter, r *http.Request) {
	type NewAppointment struct {
		MerchantName string `json:"merchant_name" validate:"required"`
		ServiceId    int    `json:"service_id" validate:"required"`
		LocationId   int    `json:"location_id" validate:"required"`
		TimeStamp    string `json:"timeStamp" validate:"required"`
				UserNote     string `json:"user_note"`
	}
	var newApp NewAppointment

	if err := validate.ParseStruct(r, &newApp); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByUrlName(r.Context(), newApp.MerchantName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching merchant by this name: %s", err.Error()))
		return
	}

	err = a.Postgresdb.IsUserBlacklisted(r.Context(), merchantId, userID)
	if err != nil {
		httputil.Error(w, http.StatusForbidden, err)
		return
	}

	service, err := a.Postgresdb.GetServiceById(r.Context(), newApp.ServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching service by this id: %s", err.Error()))
		return
	}

	if service.MerchantId != merchantId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("this serivce id does not belong to this merchant"))
		return
	}

	location, err := a.Postgresdb.GetLocationById(r.Context(), newApp.LocationId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching location by this id: %s", err.Error()))
		return
	}

	if location.MerchantId != merchantId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("this location id does not belong to this merchant"))
		return
	}

	duration, err := time.ParseDuration(strconv.Itoa(service.Duration) + "m")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("duration could not be parsed: %s", err.Error()))
		return
	}

	timeStamp, err := time.Parse(time.RFC3339, newApp.TimeStamp)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("timestamp could not be converted to time: %s", err.Error()))
		return
	}

	const postgresTimeFormat = "2006-01-02T15:04:05-0700"

	toDate := timeStamp.Add(duration)
	from_date := timeStamp.Format(postgresTimeFormat)
	to_date := toDate.Format(postgresTimeFormat)

	appointmentId, err := a.Postgresdb.NewAppointment(r.Context(), database.Appointment{
		Id:           0,
		UserId:       userID,
		MerchantId:   merchantId,
		ServiceId:    newApp.ServiceId,
		LocationId:   newApp.LocationId,
		FromDate:     from_date,
		ToDate:       to_date,
		UserNote:     newApp.UserNote,
		MerchantNote: "",
		PriceThen:    service.Price,
		CostThen:     service.Cost,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not make new apppointment: %s", err.Error()))
		return
	}

	userInfo, err := a.Postgresdb.GetUserById(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not get email for the user: %s", err.Error()))
		return
	}

	timezone, err := a.Postgresdb.GetMerchantTimezoneById(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	merchantTz, err := time.LoadLocation(timezone)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while parsing merchant's timezone: %s", err.Error()))
		return
	}

	toDateMerchantTz := toDate.In(merchantTz)
	fromDateMerchantTz := timeStamp.In(merchantTz)

	emailData := email.AppointmentConfirmationData{
		Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:        fromDateMerchantTz.Format("MONDAY, JANUARY 2"),
		Location:    location.PostalCode + " " + location.City + " " + location.Address,
		ServiceName: service.Name,
		TimeZone:    merchantTz.String(),
		ModifyLink:  "http://localhost:5173/settings/profile",
	}

	err = email.AppointmentConfirmation(r.Context(), userInfo.Email, emailData)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not send confirmation email for the appointment: %s", err.Error()))
		return
	}

	hoursUntilAppointment := time.Until(timeStamp).Hours()

	if hoursUntilAppointment >= 24 {

		reminderDate := timeStamp.Add(-24 * time.Hour)
		email_id, err := email.AppointmentReminder(r.Context(), userInfo.Email, emailData, reminderDate)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not schedule reminder email: %s", err.Error()))
			return
		}

		if email_id != "" {
			err = a.Postgresdb.UpdateEmailIdForAppointment(r.Context(), appointmentId, email_id)
			if err != nil {
				httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to update email ID: %s", err.Error()))
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *Appointment) GetAppointments(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	apps, err := a.Postgresdb.GetAppointmentsByMerchant(r.Context(), merchantId, start, end)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	httputil.Success(w, http.StatusOK, apps)
}

func (a *Appointment) CancelAppointmentByMerchant(w http.ResponseWriter, r *http.Request) {
	type cancelAppointmentData struct {
		CancellationReason string `json:"cancellation_reason"`
	}

	urlId := chi.URLParam(r, "id")

	if urlId == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid appointment id provided"))
		return
	}

	appId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting appointment id to int: %s", err.Error()))
		return
	}

	var cancelData cancelAppointmentData

	if err := validate.ParseStruct(r, &cancelData); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = a.Postgresdb.CancelAppointmentByMerchant(r.Context(), merchantId, appId, cancelData.CancellationReason)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling appointment by merchant: %s", err.Error()))
		return
	}

	timezone, err := a.Postgresdb.GetMerchantTimezoneById(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	merchantTz, err := time.LoadLocation(timezone)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while parsing merchant's timezone: %s", err.Error()))
		return
	}

	emailData, err := a.Postgresdb.GetAppointmentDataForEmail(r.Context(), appId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving data for email sending: %s", err.Error()))
		return
	}

	toDateMerchantTz := emailData.ToDate.In(merchantTz)
	fromDateMerchantTz := emailData.FromDate.In(merchantTz)

	err = email.AppointmentCancellation(r.Context(), emailData.UserEmail, email.AppointmentCancellationData{
		Time:               fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:               fromDateMerchantTz.Format("MONDAY, JANUARY 2"),
		Location:           emailData.ShortLocation,
		ServiceName:        emailData.ServiceName,
		TimeZone:           merchantTz.String(),
		Reason:             cancelData.CancellationReason,
		NewAppointmentLink: "http://localhost:5173/settings/profile",
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while sending cancellation email: %s", err.Error()))
		return
	}

	if emailData.EmailId != uuid.Nil {
		err = email.Cancel(emailData.EmailId.String())
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling the scheduled email for the appointment: %s", err.Error()))
			return
		}
	}
}

func (a *Appointment) UpdateAppointmentData(w http.ResponseWriter, r *http.Request) {
	// validate:"required" on MerchantNote would fail
	// if an empty string arrives as a note
	type appointmentData struct {
		Id           int    `json:"id" validate:"required"`
		MerchantNote string `json:"merchant_note"`
		FromDate     string `json:"from_date" validate:"required"`
		ToDate       string `json:"to_date" validate:"required"`
	}

	var appData appointmentData

	if err := validate.ParseStruct(r, &appData); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching for the merchant by this owner id: %s", err.Error()))
		return
	}

	fromDate, err := time.Parse(time.RFC3339, appData.FromDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("from_date could not be converted to time: %s", err.Error()))
		return
	}

	toDate, err := time.Parse(time.RFC3339, appData.ToDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("to_date could not be converted to time: %s", err.Error()))
		return
	}

	const postgresTimeFormat = "2006-01-02T15:04:05-0700"

	formattedFromDate := fromDate.Format(postgresTimeFormat)
	formattedToDate := toDate.Format(postgresTimeFormat)

	if err := a.Postgresdb.UpdateAppointmentData(r.Context(), merchantId, appData.Id, appData.MerchantNote, formattedFromDate, formattedToDate); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}
}
