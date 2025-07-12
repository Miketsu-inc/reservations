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
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/lang"
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
		CustomerNote string `json:"customer_note"`
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

	service, err := a.Postgresdb.GetServiceWithPhasesById(r.Context(), newApp.ServiceId, merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching service by this id: %s", err.Error()))
		return
	}

	if service.MerchantId != merchantId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("this service id does not belong to this merchant"))
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

	duration := time.Duration(service.TotalDuration) * time.Minute

	timeStamp, err := time.Parse(time.RFC3339, newApp.TimeStamp)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("timestamp could not be converted to time: %s", err.Error()))
		return
	}

	toDate := timeStamp.Add(duration)

	customerId, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating customer id: %s", err.Error()))
		return
	}

	appointmentId, err := a.Postgresdb.NewAppointment(r.Context(), database.Appointment{
		Id:           0,
		MerchantId:   merchantId,
		ServiceId:    newApp.ServiceId,
		LocationId:   newApp.LocationId,
		GroupId:      0,
		FromDate:     timeStamp,
		ToDate:       toDate,
		CustomerNote: newApp.CustomerNote,
		MerchantNote: "",
		PriceThen:    service.Price,
		CostThen:     service.Cost,
	}, service.Phases, userID, customerId)
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
		Date:        fromDateMerchantTz.Format("Monday, January 2"),
		Location:    location.PostalCode + " " + location.City + " " + location.Address,
		ServiceName: service.Name,
		TimeZone:    merchantTz.String(),
		ModifyLink:  "http://localhost:5173/m/" + newApp.MerchantName + "/cancel/" + strconv.Itoa(appointmentId),
	}

	lang := lang.LangFromContext(r.Context())

	err = email.AppointmentConfirmation(r.Context(), lang, userInfo.Email, emailData)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not send confirmation email for the appointment: %s", err.Error()))
		return
	}

	hoursUntilAppointment := time.Until(fromDateMerchantTz).Hours()

	if hoursUntilAppointment >= 24 {

		reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)
		email_id, err := email.AppointmentReminder(r.Context(), lang, userInfo.Email, emailData, reminderDate)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not schedule reminder email: %s", err.Error()))
			return
		}

		if email_id != "" { //check because return "" when email sending is off
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

	if emailData.FromDate.Before(time.Now().UTC()) {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("you cannot cancel past appointments"))
		return
	}

	toDateMerchantTz := emailData.ToDate.In(merchantTz)
	fromDateMerchantTz := emailData.FromDate.In(merchantTz)

	err = a.Postgresdb.CancelAppointmentByMerchant(r.Context(), merchantId, appId, cancelData.CancellationReason)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling appointment by merchant: %s", err.Error()))
		return
	}

	lang := lang.LangFromContext(r.Context())

	err = email.AppointmentCancellation(r.Context(), lang, emailData.CustomerEmail, email.AppointmentCancellationData{
		Time:               fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:               fromDateMerchantTz.Format("Monday, January 2"),
		Location:           emailData.ShortLocation,
		ServiceName:        emailData.ServiceName,
		TimeZone:           merchantTz.String(),
		Reason:             cancelData.CancellationReason,
		NewAppointmentLink: "http://localhost:5173/m/" + emailData.MerchantName,
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
		MerchantNote string `json:"merchant_note"`
		FromDate     string `json:"from_date" validate:"required"`
		ToDate       string `json:"to_date" validate:"required"`
	}

	var appData appointmentData

	if err := validate.ParseStruct(r, &appData); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlId := chi.URLParam(r, "id")

	appId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting appointment id to int: %s", err.Error()))
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

	oldEmailData, err := a.Postgresdb.GetAppointmentDataForEmail(r.Context(), appId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving data for email sending: %s", err.Error()))
		return
	}

	fromDateOffset := fromDate.Sub(oldEmailData.FromDate)
	toDateOffset := fromDate.Sub(oldEmailData.FromDate)

	if fromDateOffset != toDateOffset {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid from and to date supplied"))
		return
	}

	if err := a.Postgresdb.UpdateAppointmentData(r.Context(), merchantId, appId, appData.MerchantNote, fromDateOffset); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
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
	fromDateMerchantTz := fromDate.In(merchantTz)
	oldToDateMerchantTz := oldEmailData.ToDate.In(merchantTz)
	oldFromDateMerchantTz := oldEmailData.FromDate.In(merchantTz)

	lang := lang.LangFromContext(r.Context())

	err = email.AppointmentModification(r.Context(), lang, oldEmailData.CustomerEmail, email.AppointmentModificationData{
		Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:        fromDate.Format("Monday, January 2"),
		Location:    oldEmailData.ShortLocation,
		ServiceName: oldEmailData.ServiceName,
		TimeZone:    merchantId.String(),
		ModifyLink:  fmt.Sprintf("http://localhost:5173/m/%s/cancel/%d", oldEmailData.MerchantName, appId),
		OldTime:     oldFromDateMerchantTz.Format("15:04") + " - " + oldToDateMerchantTz.Format("15:04"),
		OldDate:     oldEmailData.FromDate.Format("Monday, January 2"),
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	hoursUntilAppointment := time.Until(fromDateMerchantTz).Hours()

	if oldEmailData.EmailId != uuid.Nil {
		// Always cancel the old email — content might be outdated
		err := email.Cancel(oldEmailData.EmailId.String())
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not cancel old reminder email: %s", err.Error()))
			return
		}
	}

	// Only schedule new one if new time is valid
	if hoursUntilAppointment >= 24 {
		reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)

		email_id, err := email.AppointmentReminder(r.Context(), lang, oldEmailData.CustomerEmail, email.AppointmentConfirmationData{
			Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
			Date:        fromDateMerchantTz.Format("Monday, January 2"),
			Location:    oldEmailData.ShortLocation,
			ServiceName: oldEmailData.ServiceName,
			TimeZone:    merchantTz.String(),
			ModifyLink:  fmt.Sprintf("http://localhost:5173/m/%s/cancel/%d", oldEmailData.MerchantName, appId),
		}, reminderDate)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not schedule reminder email: %s", err.Error()))
			return
		}

		if email_id != "" { //check because return "" when email sending is off
			err = a.Postgresdb.UpdateEmailIdForAppointment(r.Context(), appId, email_id)
			if err != nil {
				httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to update email ID: %s", err.Error()))
				return
			}
		}
	}

}

func (a *Appointment) GetPublicAppointmentData(w http.ResponseWriter, r *http.Request) {
	urlId := chi.URLParam(r, "id")

	appId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting appointment id to int: %s", err.Error()))
		return
	}

	appInfo, err := a.Postgresdb.GetPublicAppointmentInfo(r.Context(), appId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving public data for appointment: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, appInfo)
}

func (a *Appointment) CancelAppointmentByUser(w http.ResponseWriter, r *http.Request) {
	type cancelData struct {
		AppointmentId int    `json:"appointment_id" validate:"required"`
		MerchantName  string `json:"merchant_name" validate:"required"`
	}

	var data cancelData

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlId := chi.URLParam(r, "id")

	appId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting appointment id to int: %s", err.Error()))
		return
	}

	if appId != data.AppointmentId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("the appointment ids are not matching", err.Error()))
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByUrlName(r.Context(), data.MerchantName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching merchant by this name: %s", err.Error()))
		return
	}

	customerId, err := a.Postgresdb.GetCustomerIdByUserIdAndMerchantId(r.Context(), merchantId, userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while getting customer id: %s", err.Error()))
	}

	emailId, err := a.Postgresdb.CancelAppointmentByUser(r.Context(), customerId, appId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling the appointment by user: %s", err.Error()))
		return
	}

	if emailId != uuid.Nil {
		err = email.Cancel(emailId.String())
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling the scheduled email for the appointment: %s", err.Error()))
			return
		}
	}
}
