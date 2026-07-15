import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Check, Trash2, Users, Megaphone, Wallet, TrendingUp } from 'lucide-react'
import {
  useAdminMetrics,
  useAdminUsers,
  useAdminCampaigns,
  useAdminPayments,
  useDeleteAdminCampaign,
  useUpdateOrgCommissionRate,
  usePendingPayouts,
  useCreatePayout,
} from '../hooks/useAdmin'
import { formatCLP } from '../../../shared/lib/clp'
import { typeLabels } from '../../campaigns/typeMeta'
import { Badge, Button, Card, Input } from '../../../shared/ui'
import type { AdminCampaign, AdminPayment, AdminUser, PendingPayout } from '../api'

const campaignStatusTone: Record<string, 'neutral' | 'success' | 'warning' | 'danger'> = {
  draft: 'neutral',
  active: 'success',
  paused: 'warning',
  finished: 'neutral',
  suspended: 'danger',
}

const paymentStatusTone: Record<string, 'neutral' | 'success' | 'warning' | 'danger'> = {
  pending: 'warning',
  confirmed: 'success',
  failed: 'danger',
  refunded: 'neutral',
  partially_refunded: 'neutral',
}

function MetricCard({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Users
  label: string
  value: string
}) {
  return (
    <Card className="flex items-center gap-3">
      <div className="flex h-9 w-9 flex-none items-center justify-center rounded-full bg-accent-tint text-brand-hover">
        <Icon className="h-4 w-4" strokeWidth={1.75} />
      </div>
      <div>
        <p className="text-[11px] font-semibold uppercase tracking-wide text-text-secondary">{label}</p>
        <p className="font-display text-lg font-bold tabular-nums">{value}</p>
      </div>
    </Card>
  )
}

// CommissionCell: input de porcentaje editable inline — el backend guarda
// la comisión como fracción (0.05), acá se muestra/edita como porcentaje
// (5) para que el admin no tenga que hacer la conversión mentalmente.
function CommissionCell({ user }: { user: AdminUser }) {
  const currentPercent = user.organization_commission_rate * 100
  const [value, setValue] = useState(String(currentPercent))
  const updateRate = useUpdateOrgCommissionRate()
  const dirty = Number(value) !== currentPercent

  function handleSave() {
    const percent = Number(value)
    if (Number.isNaN(percent) || percent < 0 || percent > 50) return
    updateRate.mutate({ orgId: user.organization_id, rate: percent / 100 })
  }

  return (
    <div className="flex items-center gap-1.5">
      <Input
        type="number"
        min={0}
        max={50}
        step={0.1}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        className="w-16 py-1 text-right tabular-nums"
      />
      <span className="text-text-secondary">%</span>
      {dirty && (
        <button
          type="button"
          onClick={handleSave}
          disabled={updateRate.isPending}
          className="text-text-secondary hover:text-brand-hover disabled:opacity-50"
          title="Guardar comisión"
        >
          <Check className="h-4 w-4" strokeWidth={2} />
        </button>
      )}
    </div>
  )
}

function UsersTable({ items }: { items: AdminUser[] }) {
  if (items.length === 0) return <p className="text-sm text-text-secondary">Sin usuarios todavía.</p>
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-border-default text-text-secondary">
            <th className="py-2 pr-4">Nombre</th>
            <th className="py-2 pr-4">Email</th>
            <th className="py-2 pr-4">Organización</th>
            <th className="py-2 pr-4">Comisión</th>
            <th className="py-2 pr-4">Campañas</th>
            <th className="py-2 pr-4">Registrado</th>
          </tr>
        </thead>
        <tbody>
          {items.map((u) => (
            <tr key={u.id} className="border-b border-border-default last:border-0">
              <td className="py-2 pr-4">{u.full_name}</td>
              <td className="py-2 pr-4 text-text-secondary">
                {u.email}
                {!u.email_verified && (
                  <span className="ml-1.5 text-xs text-text-secondary">(sin verificar)</span>
                )}
              </td>
              <td className="py-2 pr-4">{u.organization_name}</td>
              <td className="py-2 pr-4">
                <CommissionCell user={u} />
              </td>
              <td className="py-2 pr-4 tabular-nums">{u.campaign_count}</td>
              <td className="py-2 pr-4 text-text-secondary">
                {new Date(u.created_at).toLocaleDateString('es-CL')}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function CampaignsTable({ items }: { items: AdminCampaign[] }) {
  const deleteCampaign = useDeleteAdminCampaign()

  function handleDelete(c: AdminCampaign) {
    if (!window.confirm(`¿Eliminar "${c.title}"? Esta acción no se puede deshacer.`)) return
    deleteCampaign.mutate(c.id)
  }

  if (items.length === 0) return <p className="text-sm text-text-secondary">Sin campañas todavía.</p>
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-border-default text-text-secondary">
            <th className="py-2 pr-4">Campaña</th>
            <th className="py-2 pr-4">Tipo</th>
            <th className="py-2 pr-4">Organizador</th>
            <th className="py-2 pr-4">Estado</th>
            <th className="py-2 pr-4">Recaudado</th>
            <th className="py-2 pr-4">Aportantes</th>
            <th className="py-2 pr-4" />
          </tr>
        </thead>
        <tbody>
          {items.map((c) => (
            <tr key={c.id} className="border-b border-border-default last:border-0">
              <td className="py-2 pr-4">
                <Link to={`/public/${c.slug}`} target="_blank" className="hover:underline">
                  {c.title}
                </Link>
              </td>
              <td className="py-2 pr-4 text-text-secondary">{typeLabels[c.type_key] ?? c.type_key}</td>
              <td className="py-2 pr-4 text-text-secondary">{c.organizer_email}</td>
              <td className="py-2 pr-4">
                <Badge tone={campaignStatusTone[c.status] ?? 'neutral'}>{c.status}</Badge>
              </td>
              <td className="py-2 pr-4 tabular-nums">{formatCLP(c.raised_gross)}</td>
              <td className="py-2 pr-4 tabular-nums">{c.contributor_count}</td>
              <td className="py-2 pr-4">
                <button
                  type="button"
                  onClick={() => handleDelete(c)}
                  disabled={deleteCampaign.isPending}
                  className="text-text-secondary hover:text-danger disabled:opacity-50"
                  title="Eliminar campaña"
                >
                  <Trash2 className="h-4 w-4" strokeWidth={1.75} />
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function PaymentsTable({ items }: { items: AdminPayment[] }) {
  if (items.length === 0) return <p className="text-sm text-text-secondary">Sin pagos todavía.</p>
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-border-default text-text-secondary">
            <th className="py-2 pr-4">Campaña</th>
            <th className="py-2 pr-4">Aportante</th>
            <th className="py-2 pr-4">Proveedor</th>
            <th className="py-2 pr-4">Estado</th>
            <th className="py-2 pr-4">Bruto</th>
            <th className="py-2 pr-4">Comisión</th>
            <th className="py-2 pr-4">Neto</th>
            <th className="py-2 pr-4">Fecha</th>
          </tr>
        </thead>
        <tbody>
          {items.map((p) => (
            <tr key={p.id} className="border-b border-border-default last:border-0">
              <td className="py-2 pr-4">
                <Link to={`/public/${p.campaign_slug}`} target="_blank" className="hover:underline">
                  {p.campaign_title}
                </Link>
              </td>
              <td className="py-2 pr-4 text-text-secondary">{p.contributor_name}</td>
              <td className="py-2 pr-4 text-text-secondary">{p.provider}</td>
              <td className="py-2 pr-4">
                <Badge tone={paymentStatusTone[p.status] ?? 'neutral'}>{p.status}</Badge>
              </td>
              <td className="py-2 pr-4 tabular-nums">{formatCLP(p.amount_gross)}</td>
              <td className="py-2 pr-4 tabular-nums text-text-secondary">
                {formatCLP(p.commission_amount)}
              </td>
              <td className="py-2 pr-4 tabular-nums">{formatCLP(p.amount_net)}</td>
              <td className="py-2 pr-4 text-text-secondary">
                {new Date(p.created_at).toLocaleDateString('es-CL')}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

// PayoutForm: registra una transferencia manual ya hecha por el operador —
// no dispara ningún movimiento de dinero real, solo deja constancia para
// que ListPendingPayouts no la vuelva a contar como pendiente.
function PayoutForm({ payout, onDone }: { payout: PendingPayout; onDone: () => void }) {
  const [amount, setAmount] = useState(String(payout.pending_amount))
  const [note, setNote] = useState('')
  const createPayout = useCreatePayout()

  function submit() {
    const parsed = Number(amount)
    if (!parsed || parsed <= 0 || parsed > payout.pending_amount) return
    createPayout.mutate(
      { orgId: payout.organization_id, amount: parsed, note: note || undefined },
      { onSuccess: onDone },
    )
  }

  return (
    <div className="flex flex-wrap items-center gap-2 py-2">
      <Input
        type="number"
        min={1}
        max={payout.pending_amount}
        value={amount}
        onChange={(e) => setAmount(e.target.value)}
        className="w-32"
        aria-label="Monto transferido"
      />
      <Input
        type="text"
        placeholder="Nota (opcional)"
        value={note}
        onChange={(e) => setNote(e.target.value)}
        className="w-48"
      />
      <Button onClick={submit} disabled={createPayout.isPending}>
        {createPayout.isPending ? 'Guardando…' : 'Confirmar transferencia'}
      </Button>
      <Button variant="secondary" onClick={onDone} disabled={createPayout.isPending}>
        Cancelar
      </Button>
      {createPayout.isError && (
        <span className="text-xs text-danger">No se pudo registrar la transferencia.</span>
      )}
    </div>
  )
}

const accountTypeLabels: Record<string, string> = {
  corriente: 'Cuenta corriente',
  vista: 'Cuenta vista',
  ahorro: 'Cuenta de ahorro',
  rut: 'Cuenta RUT',
}

function PendingPayoutsTable({ items }: { items: PendingPayout[] }) {
  const [paying, setPaying] = useState<string | null>(null)

  if (items.length === 0) {
    return <p className="text-sm text-text-secondary">No hay liquidaciones pendientes.</p>
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-border-default text-text-secondary">
            <th className="py-2 pr-4">Organización</th>
            <th className="py-2 pr-4">Datos bancarios</th>
            <th className="py-2 pr-4">Pendiente</th>
            <th className="py-2 pr-4" />
          </tr>
        </thead>
        <tbody>
          {items.map((p) => (
            <tr key={p.organization_id} className="border-b border-border-default last:border-0">
              {paying === p.organization_id ? (
                <td colSpan={4} className="px-2">
                  <PayoutForm payout={p} onDone={() => setPaying(null)} />
                </td>
              ) : (
                <>
                  <td className="py-2 pr-4">{p.organization_name}</td>
                  <td className="py-2 pr-4 text-text-secondary">
                    {p.payout_bank ? (
                      <>
                        {p.payout_bank} — {accountTypeLabels[p.payout_account_type] ?? p.payout_account_type}
                        <br />
                        {p.payout_account_number} · {p.payout_holder_name} · {p.rut}
                      </>
                    ) : (
                      <span className="text-danger">Sin datos bancarios cargados</span>
                    )}
                  </td>
                  <td className="py-2 pr-4 tabular-nums">{formatCLP(p.pending_amount)}</td>
                  <td className="py-2 pr-4 text-right">
                    {p.payout_bank && (
                      <button
                        className="text-xs font-medium text-brand-hover hover:underline"
                        onClick={() => setPaying(p.organization_id)}
                      >
                        Marcar como pagado
                      </button>
                    )}
                  </td>
                </>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

const tabs = [
  { id: 'users', label: 'Usuarios' },
  { id: 'campaigns', label: 'Campañas' },
  { id: 'payments', label: 'Pagos' },
  { id: 'payouts', label: 'Liquidaciones' },
] as const

type TabID = (typeof tabs)[number]['id']

// Backoffice del operador de la plataforma: solo accesible para las cuentas
// en ADMIN_EMAILS (el gate real vive en el backend — ver
// middleware.RequireAdminUser — esta página solo asume que ya pasó
// RequireAdmin en el router). Un solo tab-switcher en vez de 3 páginas
// separadas: a esta escala no vale la pena la complejidad de sub-rutas.
export function BackofficePage() {
  const [tab, setTab] = useState<TabID>('users')
  const { data: metrics, isLoading: loadingMetrics } = useAdminMetrics()
  const users = useAdminUsers()
  const campaigns = useAdminCampaigns()
  const payments = useAdminPayments()
  const payouts = usePendingPayouts()

  const active =
    tab === 'users' ? users : tab === 'campaigns' ? campaigns : tab === 'payments' ? payments : payouts

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-2xl font-bold tracking-tight">Backoffice</h1>
        <p className="text-sm text-text-secondary">Vista global de la plataforma — todas las organizaciones.</p>
      </div>

      {!loadingMetrics && metrics && (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          <MetricCard icon={Users} label="Usuarios" value={String(metrics.total_users)} />
          <MetricCard
            icon={Megaphone}
            label="Campañas activas"
            value={`${metrics.active_campaigns} / ${metrics.total_campaigns}`}
          />
          <MetricCard icon={TrendingUp} label="Recaudado bruto" value={formatCLP(metrics.raised_gross)} />
          <MetricCard icon={Wallet} label="Comisión generada" value={formatCLP(metrics.total_commission)} />
        </div>
      )}

      <div className="flex gap-2 border-b border-border-default">
        {tabs.map((t) => (
          <button
            key={t.id}
            type="button"
            onClick={() => setTab(t.id)}
            className={`border-b-2 px-3 py-2 text-sm font-medium transition-colors ${
              tab === t.id
                ? 'border-brand-hover text-brand-hover'
                : 'border-transparent text-text-secondary hover:text-text-primary'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <Card>
        {active.isLoading ? (
          <p className="text-sm text-text-secondary">Cargando…</p>
        ) : (
          <>
            {tab === 'users' && <UsersTable items={users.items} />}
            {tab === 'campaigns' && <CampaignsTable items={campaigns.items} />}
            {tab === 'payments' && <PaymentsTable items={payments.items} />}
            {tab === 'payouts' && <PendingPayoutsTable items={payouts.items} />}
          </>
        )}
        {active.hasNextPage && (
          <div className="mt-4 text-center">
            <Button
              variant="secondary"
              onClick={() => active.fetchNextPage()}
              disabled={active.isFetchingNextPage}
            >
              {active.isFetchingNextPage ? 'Cargando…' : 'Cargar más'}
            </Button>
          </div>
        )}
      </Card>
    </div>
  )
}
