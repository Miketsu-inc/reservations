package servicecategories

import catalogServ "github.com/miketsu-inc/reservations/backend/internal/service/catalog"

func mapToNewCategoryInput(in newReq) catalogServ.NewCategoryInput {
	return catalogServ.NewCategoryInput{
		Name: in.Name,
	}
}

func mapToUpdateCategoryInput(in updateReq) catalogServ.UpdateCategoryInput {
	return catalogServ.UpdateCategoryInput{
		Name: in.Name,
	}
}

func mapToReorderCategoriesInput(in reorderCategoriesReq) catalogServ.ReorderCategoriesInput {
	return catalogServ.ReorderCategoriesInput{
		Categories: in.Categories,
	}
}
