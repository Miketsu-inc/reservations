package review

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

func parseIdList(q url.Values, key string) ([]int, error) {
	raw := q.Get(key)
	if raw == "" {
		return nil, nil
	}

	var ids []int
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil || id < 1 {
			return nil, validate.NewError("invalid id in the array")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

type ReviewQueryParams struct {
	Rating      *int16
	EmployeeIds []int
	ServiceIds  []int
	HasReply    *bool
	Sort        string
	Cursor      string
	PageSize    int
}

func ParseReviewQueryParams(query url.Values) (ReviewQueryParams, error) {
	var rating *int16
	if val := query.Get("stars"); val != "" {
		n, err := strconv.Atoi(val)
		if err != nil || n < 1 || n > 5 {
			return ReviewQueryParams{}, validate.NewError("stars must be between 1 and 5")
		}
		r := int16(n)
		rating = &r
	}

	employeeIds, err := parseIdList(query, "employee_ids")
	if err != nil {
		return ReviewQueryParams{}, err
	}

	serviceIds, err := parseIdList(query, "service_ids")
	if err != nil {
		return ReviewQueryParams{}, err
	}

	var hasReply *bool
	if val := query.Get("has_reply"); val != "" {
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return ReviewQueryParams{}, validate.NewError("invalid has_reply boolean")
		}
		hasReply = &parsed
	}

	limit := 10
	if val := query.Get("limit"); val != "" {
		n, err := strconv.Atoi(val)
		if err != nil || n < 1 || n > 20 {
			return ReviewQueryParams{}, validate.NewError("limit must be between 1 and 20")
		}
		limit = n
	}

	return ReviewQueryParams{
		Rating:      rating,
		EmployeeIds: employeeIds,
		ServiceIds:  serviceIds,
		HasReply:    hasReply,
		Sort:        query.Get("sort"),
		Cursor:      query.Get("cursor"),
		PageSize:    limit,
	}, nil
}
