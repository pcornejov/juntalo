// Normaliza para búsqueda insensible a mayúsculas y tildes ("María" === "maria").
export function normalizeForSearch(s: string): string {
  return s.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '')
}
