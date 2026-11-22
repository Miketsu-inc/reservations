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
		ServiceCounts: stats.ServiceCounts,
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
		FirstName: emp.FirstName,
		LastName:  emp.LastName,
	}
}

func mapToReviewItemResp(item domain.ReviewDetails) reviewItemResp {
	var reply *reviewReplyResp
	if item.Reply != nil {
		reply = &reviewReplyResp{
			Comment: item.Reply.Comment,
		}
	}

	return reviewItemResp{
		Id:          item.Id,
		ServiceName: item.ServiceName,
		Rating:      item.Rating,
		Comment:     item.Comment,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		Customer: reviewCustomerResp{
			FirstName: item.Customer.FirstName,
			LastName:  item.Customer.LastName,
			AvatarUrl: mediaServ.NewURL(item.Customer.AvatarImageId, mediaServ.VariantSmall),
		},
		Employee: mapToReviewEmployeeResp(item.Employee),
		Reply:    reply,
	}
}
