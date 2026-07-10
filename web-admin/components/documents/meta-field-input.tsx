import { MetadataField } from "@/lib/api";

interface Props {
  field: MetadataField;
  value: string;
  onChange: (value: string) => void;
}

export function MetaFieldInput({ field, value, onChange }: Props) {
  const base = "w-full rounded border border-zinc-200 px-2 py-1 text-xs font-mono focus:outline-none focus:ring-1 focus:ring-zinc-300";

  switch (field.type) {
    case "bool":
      return (
        <div className="flex items-center h-[26px]">
          <input
            type="checkbox"
            checked={value === "true"}
            onChange={e => onChange(e.target.checked ? "true" : "false")}
            className="w-4 h-4 accent-primary"
          />
        </div>
      );
    case "date":
      return (
        <input
          type="date"
          value={value}
          onChange={e => onChange(e.target.value)}
          className={base}
        />
      );
    case "num":
      return (
        <input
          type="number"
          step="any"
          value={value}
          onChange={e => onChange(e.target.value)}
          placeholder="e.g. 4.5"
          className={base}
        />
      );
    case "arr":
      return (
        <input
          type="text"
          value={value}
          onChange={e => onChange(e.target.value)}
          placeholder="comma-separated: go, rust"
          className={base}
        />
      );
    default:
      return (
        <input
          type="text"
          value={value}
          onChange={e => onChange(e.target.value)}
          placeholder="value"
          className={base}
        />
      );
  }
}
