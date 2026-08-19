export type ListenerFormValues = Record<string, any>;

const OWNED = new Set([
  'reality-config', 'certificate', 'private-key', 'ws-path', 'ws-headers',
  'grpc-service-name', 'xhttp-config', 'users', 'flow', 'decryption',
  'alpn', 'server-names', 'short-id', 'dest', 'tls', 'transport',
]);

function nonEmpty(v: any): boolean {
  return v !== undefined && v !== null && !(typeof v === 'string' && v.trim() === '');
}

function arrayOfStrings(v: any): string[] | undefined {
  if (!nonEmpty(v)) return undefined;
  if (Array.isArray(v)) return v.map(String).filter(Boolean);
  return String(v).split(/[\n,]/).map(s => s.trim()).filter(Boolean);
}

/** Convert a stored Mihomo listener JSON object into stable UI aliases. */
export function listenerConfigToForm(raw: string | Record<string, any> | undefined | null): ListenerFormValues {
  let cfg: Record<string, any> = {};
  if (typeof raw === 'string') {
    try { cfg = JSON.parse(raw); } catch { return {}; }
  } else if (raw && typeof raw === 'object') cfg = raw;

  const out: ListenerFormValues = { ...cfg };
  const reality = cfg['reality-config'];
  if (reality && typeof reality === 'object') {
    out.reality_enabled = true;
    out.reality_dest = reality.dest ?? '';
    out.reality_private_key = reality['private-key'] ?? '';
    out.reality_short_id = arrayOfStrings(reality['short-id']) ?? [];
    out.reality_server_names = arrayOfStrings(reality['server-names']) ?? [];
  } else {
    out.reality_enabled = false;
    out.reality_dest = '';
    out.reality_private_key = '';
    out.reality_short_id = [];
    out.reality_server_names = [];
  }

  out.tls_certificate = cfg.certificate ?? '';
  out.tls_private_key = cfg['private-key'] ?? '';
  out.transport_layer = cfg['xhttp-config'] ? 'xhttp' : cfg['grpc-service-name'] ? 'grpc' : cfg['ws-path'] ? 'ws' : 'raw';
  out.ws_path = cfg['ws-path'] ?? '';
  out.grpc_service_name = cfg['grpc-service-name'] ?? '';
  const xhttp = cfg['xhttp-config'];
  if (xhttp && typeof xhttp === 'object') {
    out.xhttp_path = xhttp.path ?? '';
    out.xhttp_host = xhttp.host ?? '';
    out.xhttp_mode = xhttp.mode ?? '';
  }
  return out;
}

/**
 * Serialize UI aliases into Mihomo listener JSON. This deliberately removes
 * UI-only aliases so they can never leak into the generated listener config.
 */
export function listenerFormToConfig(values: ListenerFormValues, previous?: Record<string, any> | null): Record<string, any> {
  const cfg: Record<string, any> = {};
  if (previous && typeof previous === 'object') {
    for (const [key, value] of Object.entries(previous)) {
      if (!key.startsWith('reality_') && !key.startsWith('tls_') && !key.startsWith('xhttp_') &&
          !key.startsWith('ws_') && !key.startsWith('grpc_') && !OWNED.has(key)) {
        cfg[key] = value;
      }
    }
  }

  for (const [key, value] of Object.entries(values)) {
    if (key.startsWith('reality_') || key.startsWith('tls_') || key.startsWith('xhttp_') ||
        key.startsWith('ws_') || key.startsWith('grpc_') || key === 'transport_layer' ||
        key === 'reality_enabled') continue;
    if (key === 'reality-config' || key === 'xhttp-config') continue;
    if (nonEmpty(value)) cfg[key] = value;
  }

  if (values.reality_enabled || nonEmpty(values.reality_dest) || nonEmpty(values.reality_private_key)) {
    if (!nonEmpty(values.reality_dest)) throw new Error('Reality Dest is required');
    if (!nonEmpty(values.reality_private_key)) throw new Error('Reality Private Key is required');
    const reality: Record<string, any> = {
      dest: String(values.reality_dest).trim(),
      'private-key': String(values.reality_private_key).trim(),
    };
    const shortId = arrayOfStrings(values.reality_short_id);
    const names = arrayOfStrings(values.reality_server_names);
    if (shortId?.length) reality['short-id'] = shortId;
    if (names?.length) reality['server-names'] = names;
    cfg['reality-config'] = reality;
  }

  if (nonEmpty(values.tls_certificate)) cfg.certificate = values.tls_certificate;
  if (nonEmpty(values.tls_private_key)) cfg['private-key'] = values.tls_private_key;

  const transport = values.transport_layer;
  if (transport === 'ws' && nonEmpty(values.ws_path)) cfg['ws-path'] = values.ws_path;
  if (transport === 'grpc' && nonEmpty(values.grpc_service_name)) cfg['grpc-service-name'] = values.grpc_service_name;
  if (transport === 'xhttp') {
    const xhttp: Record<string, any> = {};
    if (nonEmpty(values.xhttp_path)) xhttp.path = values.xhttp_path;
    if (nonEmpty(values.xhttp_host)) xhttp.host = values.xhttp_host;
    if (nonEmpty(values.xhttp_mode)) xhttp.mode = values.xhttp_mode;
    if (Object.keys(xhttp).length) cfg['xhttp-config'] = xhttp;
  }

  return cfg;
}

export function parseListenerFormConfig(values: ListenerFormValues): Record<string, any> {
  const parsed = typeof values.config === 'string' ? JSON.parse(values.config || '{}') : (values.config || {});
  return listenerFormToConfig({ ...parsed, ...values }, parsed);
}
