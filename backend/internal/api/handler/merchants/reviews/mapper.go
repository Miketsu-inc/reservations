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
	reviews := make([]reviewItemResp, 0, len(result.Reviews))
	for _, item := range result.Reviews {
		reviews = append(reviews, mapToReviewItemResp(item))
	}

	return getReviewsResp{
		Reviews:     reviews,
		HasNextPage: result.HasNextPage,
		NextCursor:  result.NextCursor,
	}
}

func mapToReviewItemResp(item domain.ReviewDetails) reviewItemResp {
	var userAvatar *string
	if item.Customer.AvatarImageId != nil {
		userAvatar = mediaServ.NewURL(item.Customer.AvatarImageId, mediaServ.VariantSmall)
	}

	var emp *reviewEmployeeResp
	if item.Employee != nil {
		var empAvatar *string
		if item.Employee.AvatarImageId != nil {
			empAvatar = mediaServ.NewURL(item.Employee.AvatarImageId, mediaServ.VariantSmall)
		}
		emp = &reviewEmployeeResp{
			Id:        item.Employee.Id,
			FirstName: item.Employee.FirstName,
			LastName:  item.Employee.LastName,
			AvatarUrl: empAvatar,
		}
	}

	var reply *reviewReplyResp
	if item.Reply != nil {
		var replyEmp *reviewEmployeeResp
		if item.Reply.Employee != nil {
			var replyEmpAvatar *string
			if item.Reply.Employee.AvatarImageId != nil {
				replyEmpAvatar = mediaServ.NewURL(item.Reply.Employee.AvatarImageId, mediaServ.VariantSmall)
			}
			replyEmp = &reviewEmployeeResp{
				Id:        item.Reply.Employee.Id,
				FirstName: item.Reply.Employee.FirstName,
				LastName:  item.Reply.Employee.LastName,
				AvatarUrl: replyEmpAvatar,
			}
		}
		reply = &reviewReplyResp{
			Comment:   item.Reply.Comment,
			CreatedAt: item.Reply.CreatedAt,
			UpdatedAt: item.Reply.UpdatedAt,
			Employee:  replyEmp,
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
			AvatarUrl: userAvatar,
		},
		Employee: emp,
		Reply:    reply,
	}
}
