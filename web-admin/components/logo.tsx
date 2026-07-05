export function LogoIcon({ size = 48 }: { size?: number }) {
  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img src="/logo-icon.svg" alt="Ragpack" width={size} height={size} style={{ display: "block" }} />
  );
}

export function Logo({ height = 38 }: { height?: number }) {
  const width = Math.round(height * (363 / 83));
  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img src="/logo-with-text.svg" alt="Ragpack" width={width} height={height} style={{ display: "block" }} />
  );
}
