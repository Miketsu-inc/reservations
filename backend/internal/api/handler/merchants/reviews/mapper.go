package reviews

import (
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
	reviewServ "github.com/miketsu-inc/reservations/backend/internal/service/review"
)

func mapToReviewStatsResp(stats domain.ReviewStats) reviewStatsResp {
	return reviewStatsResp{
		AverageRating: stats.AverageRating,
		TotalReviews:  stats.TotalReviews,
		Count5Star:    stats.Count5Star,
		Count4Star:    stats.Count4Star,
		Count3Star:    stats.Count3Star,
		Count2Star:    stats.Count2Star,
		Count1Star:    stats.Count1Star,
	}
}

func mapToGetReviewsResp(result reviewServ.GetReviewsResult) getReviewsResp {
	reviews := make([]reviewItemResp, len(result.Reviews))
	for i, item := range result.Reviews {
		reviews[i] = mapToReviewItemResp(item)
	}

	return getReviewsResp{
		Reviews:     reviews,
		HasNextPage: result.HasNextPage,
		NextCursor:  result.NextCursor,
	}
}

func mapToReviewEmployeeResp(emp *domain.MinimalEmployeeInfo) *reviewEmployeeResp {
	if emp == nil {
		return nil
	}

	return &reviewEmployeeResp{
		Id:        emp.Id,
		FirstName: emp.FirstName,
		LastName:  emp.LastName,
		AvatarUrl: mediaServ.NewURL(emp.AvatarImageId, mediaServ.VariantSmall),
	}
}

func mapToReviewItemResp(item domain.ReviewDetails) reviewItemResp {
	var reply *reviewReplyResp
	if item.Reply != nil {
		reply = &reviewReplyResp{
			Comment:   item.Reply.Comment,
			CreatedAt: item.Reply.CreatedAt,
			UpdatedAt: item.Reply.UpdatedAt,
			Employee:  mapToReviewEmployeeResp(item.Reply.Employee),
		}
	}

	return reviewItemResp{
		Id:           item.Id,
		BookingId:    item.BookingId,
		ServiceName:  item.ServiceName,
		ServiceId:    item.ServiceId,
		Rating:       item.Rating,
		Comment:      item.Comment,
		Status:       item.Status,
		ReportReason: item.ReportReason,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
		Customer: reviewCustomerResp{
			FirstName: item.Customer.FirstName,
			LastName:  item.Customer.LastName,
			AvatarUrl: mediaServ.NewURL(item.Customer.AvatarImageId, mediaServ.VariantSmall),
		},
		Employee: mapToReviewEmployeeResp(item.Employee),
		Reply:    reply,
	}
}
