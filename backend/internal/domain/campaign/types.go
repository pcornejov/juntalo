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

// Registry lists every known type. "collection", "sale", "event" y "raffle"
// están habilitados; course/presale quedan como slots reservados (Etapa 1
// §3) hasta que haya demanda real de habilitarlos.
var Registry = map[TypeKey]TypeDefinition{
	TypeCollection: {
		Key:     TypeCollection,
		Enabled: true,
		Labels:  TypeLabels{Name: "Colecta", CTA: "Aportar", Unit: "aportantes"},
		Rules:   TypeRules{RequiresGoalAmount: false, AllowsFreeAmount: true},
	},
	TypeSale: {
		Key:     TypeSale,
		Enabled: true,
		Labels:  TypeLabels{Name: "Venta", CTA: "Comprar", Unit: "compradores"},
		Rules:   TypeRules{RequiresGoalAmount: false, AllowsFreeAmount: false},
	},
	TypeEvent: {
		Key:     TypeEvent,
		Enabled: true,
		Labels:  TypeLabels{Name: "Evento", CTA: "Inscribirse", Unit: "inscritos"},
		Rules:   TypeRules{RequiresGoalAmount: false, AllowsFreeAmount: false},
	},
	TypeCourse:  {Key: TypeCourse, Enabled: false, Labels: TypeLabels{Name: "Curso", CTA: "Inscribirse", Unit: "inscritos"}},
	TypePresale: {Key: TypePresale, Enabled: false, Labels: TypeLabels{Name: "Preventa", CTA: "Reservar", Unit: "reservas"}},
	TypeRaffle: {
		Key:     TypeRaffle,
		Enabled: true,
		Labels:  TypeLabels{Name: "Rifa", CTA: "Participar", Unit: "participantes"},
		Rules:   TypeRules{RequiresGoalAmount: false, AllowsFreeAmount: false},
	},
}

// typeOrder es el orden estable en que se muestran los tipos habilitados —
// iterar Registry directo no sirve porque el orden de un map en Go es
// aleatorio en cada ejecución, y con un solo tipo habilitado nunca importó,
// pero con varios el frontend (que usa el primero como default) mostraría
// un tipo distinto en cada carga de página.
var typeOrder = []TypeKey{TypeCollection, TypeSale, TypeEvent, TypeRaffle, TypeCourse, TypePresale}

// EnabledTypes returns only the types the MVP allows creating, in a stable order.
func EnabledTypes() []TypeDefinition {
	out := make([]TypeDefinition, 0, len(typeOrder))
	for _, key := range typeOrder {
		if t := Registry[key]; t.Enabled {
			out = append(out, t)
		}
	}
	return out
}

func IsEnabled(key TypeKey) bool {
	t, ok := Registry[key]
	return ok && t.Enabled
}
