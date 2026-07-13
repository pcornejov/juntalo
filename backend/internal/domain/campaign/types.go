package campaign

type TypeKey string

const (
	TypeCollection TypeKey = "collection"
	TypeSale       TypeKey = "sale"
	TypeEvent      TypeKey = "event"
	TypeCourse     TypeKey = "course"
	TypePresale    TypeKey = "presale"
	TypeRaffle     TypeKey = "raffle"
)

type TypeLabels struct {
	Name string `json:"name"`
	CTA  string `json:"cta"`
	Unit string `json:"unit"`
}

type TypeRules struct {
	RequiresGoalAmount bool `json:"requires_goal_amount"`
	AllowsFreeAmount   bool `json:"allows_free_amount"`
}

// TypeDefinition is the declarative registration of a campaign type
// (Etapa 1 riesgo 1, Etapa 2 §2.4): tipos como código, no motor genérico.
type TypeDefinition struct {
	Key     TypeKey    `json:"key"`
	Enabled bool       `json:"enabled"`
	Labels  TypeLabels `json:"labels"`
	Rules   TypeRules  `json:"rules"`
}

// Registry lists every known type. Only "collection" is enabled in the MVP;
// the rest are reserved slots (Etapa 1 §3: rifas deshabilitadas).
var Registry = map[TypeKey]TypeDefinition{
	TypeCollection: {
		Key:     TypeCollection,
		Enabled: true,
		Labels:  TypeLabels{Name: "Colecta", CTA: "Aportar", Unit: "aportantes"},
		Rules:   TypeRules{RequiresGoalAmount: false, AllowsFreeAmount: true},
	},
	TypeSale:    {Key: TypeSale, Enabled: false, Labels: TypeLabels{Name: "Venta", CTA: "Comprar", Unit: "compradores"}},
	TypeEvent:   {Key: TypeEvent, Enabled: false, Labels: TypeLabels{Name: "Evento", CTA: "Inscribirse", Unit: "inscritos"}},
	TypeCourse:  {Key: TypeCourse, Enabled: false, Labels: TypeLabels{Name: "Curso", CTA: "Inscribirse", Unit: "inscritos"}},
	TypePresale: {Key: TypePresale, Enabled: false, Labels: TypeLabels{Name: "Preventa", CTA: "Reservar", Unit: "reservas"}},
	TypeRaffle:  {Key: TypeRaffle, Enabled: false, Labels: TypeLabels{Name: "Rifa", CTA: "Participar", Unit: "participantes"}},
}

// EnabledTypes returns only the types the MVP allows creating.
func EnabledTypes() []TypeDefinition {
	var out []TypeDefinition
	for _, t := range Registry {
		if t.Enabled {
			out = append(out, t)
		}
	}
	return out
}

func IsEnabled(key TypeKey) bool {
	t, ok := Registry[key]
	return ok && t.Enabled
}
