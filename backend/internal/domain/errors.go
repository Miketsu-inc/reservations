package domain

import "github.com/miketsu-inc/reservations/backend/pkg/apperr"

var ErrEmailNotUnique = &apperr.Error{Code: "email_not_unique", Message: "this email is already used"}
var ErrPhoneNumberNotUnique = &apperr.Error{Code: "phone_number_not_unique", Message: "this phone number is already used"}

var ErrBookingNotFound = &apperr.Error{Code: "booking_not_found", Message: "booking not found"}

var ErrEmployeesRequired = &apperr.Error{Code: "employees_required", Message: "at least one team member must be assigned"}
var ErrDuplicateEmployee = &apperr.Error{Code: "duplicate_employee", Message: "a team member can only be assigned once"}
var ErrEmployeeNotActive = &apperr.Error{Code: "employee_not_active", Message: "team member must be active and belong to this merchant"}
