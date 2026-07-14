package campaign

// Category is puramente descriptiva — a diferencia de TypeKey, no rige
// ninguna regla de negocio, solo cómo se agrupa/filtra una campaña en la
// sección pública "Explorar campañas" (inspirado en Vaki: Emergencia,
// Médico, Evento, etc.).
type Category string

const (
	CategoryEmergencia Category = "emergencia"
	CategorySalud      Category = "salud"
	CategoryEducacion  Category = "educacion"
	CategoryEvento     Category = "evento"
	CategoryComunidad  Category = "comunidad"
	CategoryMascotas   Category = "mascotas"
	CategoryOtro       Category = "otro"
)

type CategoryDefinition struct {
	Key   Category `json:"key"`
	Label string   `json:"label"`
	// Icon nombra un ícono de lucide-react (Etapa "frontend usa el mismo
	// set en toda la app") — el backend no depende de lucide-react, solo
	// declara qué nombre usar para no duplicar la decisión en cada cliente.
	Icon string `json:"icon"`
}

var CategoryRegistry = map[Category]CategoryDefinition{
	CategoryEmergencia: {Key: CategoryEmergencia, Label: "Emergencia", Icon: "TriangleAlert"},
	CategorySalud:      {Key: CategorySalud, Label: "Salud", Icon: "HeartPulse"},
	CategoryEducacion:  {Key: CategoryEducacion, Label: "Educación", Icon: "GraduationCap"},
	CategoryEvento:     {Key: CategoryEvento, Label: "Evento", Icon: "PartyPopper"},
	CategoryComunidad:  {Key: CategoryComunidad, Label: "Comunidad", Icon: "Users"},
	CategoryMascotas:   {Key: CategoryMascotas, Label: "Mascotas", Icon: "PawPrint"},
	CategoryOtro:       {Key: CategoryOtro, Label: "Otro", Icon: "Sparkles"},
}

// Categories returns every category in a stable, curated display order
// (map iteration order is random in Go) — "Otro" siempre al final.
func Categories() []CategoryDefinition {
	order := []Category{
		CategoryEmergencia, CategorySalud, CategoryEducacion,
		CategoryEvento, CategoryComunidad, CategoryMascotas, CategoryOtro,
	}
	out := make([]CategoryDefinition, 0, len(order))
	for _, k := range order {
		out = append(out, CategoryRegistry[k])
	}
	return out
}

func IsValidCategory(c Category) bool {
	_, ok := CategoryRegistry[c]
	return ok
}
