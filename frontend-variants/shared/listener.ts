export type ListenerFormValues = Record<string, any>;

/**
 * Canonicalize UI aliases into Mihomo listener JSON. This function is shared
 * by every UI edition so a field cannot be fixed in one design system and
 * broken in another.
 */
export function formToListenerConfig(values: ListenerFormValues): Record<string, any> {
  const out: Record<string, any> = {};
  const ignored = new Set([
    'reality_enabled', 'reality_dest', 'reality_private_key', 'reality_short_id', 'reality_server_names',
    'transport_layer', 'security_layer', 'tls_enabled', 'tls_certificate', 'tls_private_key',
  ]);
  for (const [key, value] of Object.entries(values)) {
    if (ignored.has(key) || value === undefined || value === null || value === '') continue;
    out[key] = value;
  }

  const realityEnabled = values.reality_enabled || values['reality-config'] != null;
  if (realityEnabled) {
    const dest = values.reality_dest ?? values['reality-config']?.dest;
    const privateKey = values.reality_private_key ?? values['reality-config']?.['private-key'];
    const shortId = values.reality_short_id ?? values['reality-config']?.['short-id'];
    const serverNames = values.reality_server_names ?? values['reality-config']?.['server-names'];
    if (dest || privateKey || shortId?.length || serverNames?.length) {
      out['reality-config'] = {
        ...(dest ? { dest } : {}),
        ...(privateKey ? { 'private-key': privateKey } : {}),
        ...(Array.isArray(shortId) && shortId.length ? { 'short-id': shortId } : {}),
        ...(Array.isArray(serverNames) && serverNames.length ? { 'server-names': serverNames } : {}),
      };
    }
  }

  if (values.tls_enabled || values.tls_certificate || values.tls_private_key) {
    if (values.tls_certificate) out.certificate = values.tls_certificate;
    if (values.tls_private_key) out['private-key'] = values.tls_private_key;
  }
  return out;
}

export function listenerConfigToForm(raw: string | Record<string, any>): ListenerFormValues {
  const cfg = typeof raw === 'string' ? JSON.parse(raw || '{}') : raw;
  const out: ListenerFormValues = { ...cfg };
  delete out['reality-config'];
  delete out.certificate;
  delete out['private-key'];
  const reality = cfg['reality-config'];
  if (reality) {
    out.reality_enabled = true;
    out.reality_dest = reality.dest ?? '';
    out.reality_private_key = reality['private-key'] ?? '';
    out.reality_short_id = Array.isArray(reality['short-id']) ? reality['short-id'] : reality['short-id'] ? [reality['short-id']] : [];
    out.reality_server_names = Array.isArray(reality['server-names']) ? reality['server-names'] : reality['server-names'] ? [reality['server-names']] : [];
  } else {
    out.reality_enabled = false;
  }
  if (cfg.certificate) out.tls_certificate = cfg.certificate;
  if (cfg['private-key']) out.tls_private_key = cfg['private-key'];
  return out;
}

export function validateListener(values: ListenerFormValues): string[] {
  const errors: string[] = [];
  if (values.reality_enabled) {
    if (!String(values.reality_dest ?? values['reality-config']?.dest ?? '').trim()) errors.push('Reality Dest is required');
    if (!String(values.reality_private_key ?? values['reality-config']?.['private-key'] ?? '').trim()) errors.push('Reality Private Key is required');
  }
  return errors;
}
