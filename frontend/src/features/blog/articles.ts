export interface BlogArticle {
  slug: string
  title: string
  excerpt: string
  body: string[]
}

// Contenido educativo estático (inspirado en el blog de Ceneka): sin
// backend, sin CMS — son consejos prácticos para quien va a crear su primera
// campaña, pensados para indexar en buscadores y dar contexto antes de
// compartir el link.
export const blogArticles: BlogArticle[] = [
  {
    slug: 'como-escribir-una-buena-campana',
    title: 'Cómo escribir una campaña que la gente entienda en 10 segundos',
    excerpt: 'Un título claro y una foto real convierten más que un texto largo y perfecto.',
    body: [
      'La mayoría de las personas decide si va a leer tu campaña completa (o aportar) en los primeros segundos, casi siempre desde el celular y desde un link que alguien más les compartió por WhatsApp. Eso significa que el título y la foto de portada trabajan más que cualquier párrafo del cuerpo.',
      'Un buen título dice qué es y para quién, sin adornos: "Techo nuevo para la sede vecinal" comunica más que "Ayúdanos a lograr un sueño". Evita signos de exclamación excesivos y mayúsculas sostenidas — no suman urgencia, restan seriedad.',
      'La foto importa tanto como el título. Una imagen real del lugar, la persona o el equipo detrás de la campaña genera más confianza que un banner genérico. Si no tienes una foto profesional, una foto tomada con el celular pero real vale más que un stock photo bonito.',
      'En la descripción, responde tres preguntas en orden: qué necesitas exactamente, para qué (con el mayor detalle posible) y qué pasa con el dinero si se junta más de lo esperado (o menos). Esa transparencia es lo que hace que alguien decida aportar sin tener que preguntarte por privado.',
    ],
  },
  {
    slug: 'compartir-tu-campana-por-whatsapp',
    title: 'Compartir tu campaña por WhatsApp sin que se vea como spam',
    excerpt: 'El primer mensaje que mandas define si tus contactos abren el link o lo ignoran.',
    body: [
      'El link de tu campaña ya se ve bien al compartirse: trae imagen, título y descripción, así que no necesitas explicar todo el contexto en el mensaje — el link mismo hace ese trabajo. Enfócate en decir por qué se lo mandas a esa persona en particular.',
      'Evita el mensaje genérico copiado y pegado a 40 contactos a la vez sin nombre ni contexto. Un mensaje corto y personal ("Hola María, estamos juntando esto para arreglar el techo de la sede, cualquier aporte ayuda 🙏") convierte mucho más que un texto largo enviado a un grupo grande.',
      'Los primeros aportes son los más difíciles de conseguir y los más importantes: una campaña con cero aportes genera dudas ("¿esto es real?"), mientras que una con algunos aportes ya genera confianza. Empieza compartiendo con tu círculo más cercano antes de pedirle a la gente que comparta por ti.',
      'Actualiza a quienes ya aportaron cuando la campaña avance — un mensaje de agradecimiento con el porcentaje logrado motiva a que ellos mismos la vuelvan a compartir.',
    ],
  },
  {
    slug: 'que-pasa-si-no-alcanzo-la-meta',
    title: '¿Qué pasa si no alcanzas la meta de tu campaña?',
    excerpt: 'A diferencia de otras plataformas, en Juntalo te quedas con todo lo recaudado.',
    body: [
      'Una duda común antes de crear una campaña es qué pasa si no se junta el monto completo. En Juntalo no existe el modelo de "todo o nada": cada aporte se confirma al momento y queda disponible para el organizador, sin importar si se alcanza la meta o no.',
      'Esto significa que la meta es solo una referencia visual para quien aporta (una forma de mostrar avance), no una condición para recibir el dinero. Si juntas el 40% de tu meta, ese 40% es tuyo — no se devuelve a los aportantes ni queda retenido.',
      'Por eso conviene poner una meta realista, no la cifra ideal: una meta muy alta que nunca se alcanza puede transmitir la sensación equivocada de que la campaña "fracasó", aunque en la práctica sí haya cumplido su objetivo real.',
    ],
  },
]

export function getBlogArticle(slug: string) {
  return blogArticles.find((a) => a.slug === slug)
}
