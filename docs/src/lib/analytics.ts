declare const gtag: (...args: unknown[]) => void;

export function track(action: string, label?: string) {
  if (typeof gtag !== 'undefined') {
    gtag('event', action, { event_label: label });
  }
}
