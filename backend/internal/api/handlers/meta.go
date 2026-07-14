package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/pcornejov/juntalo/backend/internal/api/dto"
	"github.com/pcornejov/juntalo/backend/internal/domain/campaign"
)

// CampaignTypes serves the declarative type registry so the frontend can
// render the creation form without hardcoding types (Etapa 4 §7).
func CampaignTypes(c *fiber.Ctx) error {
	enabled := campaign.EnabledTypes()
	out := make([]dto.CampaignTypeResponse, len(enabled))
	for i, t := range enabled {
		out[i] = dto.CampaignTypeResponse{
			Key:                string(t.Key),
			Name:               t.Labels.Name,
			CTA:                t.Labels.CTA,
			Unit:               t.Labels.Unit,
			RequiresGoalAmount: t.Rules.RequiresGoalAmount,
			AllowsFreeAmount:   t.Rules.AllowsFreeAmount,
		}
	}
	return c.JSON(fiber.Map{"items": out})
}

// CampaignCategories serves the declarative category registry so the
// frontend no hardcodea labels/íconos al filtrar/crear (inspirado en Vaki).
func CampaignCategories(c *fiber.Ctx) error {
	cats := campaign.Categories()
	out := make([]dto.CampaignCategoryResponse, len(cats))
	for i, cat := range cats {
		out[i] = dto.CampaignCategoryResponse{Key: string(cat.Key), Label: cat.Label, Icon: cat.Icon}
	}
	return c.JSON(fiber.Map{"items": out})
}
