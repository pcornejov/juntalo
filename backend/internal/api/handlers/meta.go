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
