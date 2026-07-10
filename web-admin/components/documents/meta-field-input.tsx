import { MetadataField } from "@/lib/api";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

interface Props {
  field: MetadataField;
  value: string;
  onChange: (value: string) => void;
}

const BOOL_UNSET = "__unset__";

export function MetaFieldInput({ field, value, onChange }: Props) {
  switch (field.type) {
    case "bool": {
      const selectValue = value === "true" || value === "false" ? value : BOOL_UNSET;
      return (
        <Select value={selectValue} onValueChange={v => onChange(!v || v === BOOL_UNSET ? "" : v)}>
          <SelectTrigger className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={BOOL_UNSET}>Not set</SelectItem>
            <SelectItem value="true">True</SelectItem>
            <SelectItem value="false">False</SelectItem>
          </SelectContent>
        </Select>
      );
    }
    case "date":
      return (
        <Input
          type="date"
          value={value}
          onChange={e => onChange(e.target.value)}
        />
      );
    case "num":
      return (
        <Input
          type="number"
          step="any"
          value={value}
          onChange={e => onChange(e.target.value)}
          placeholder="e.g. 4.5"
        />
      );
    case "arr":
      return (
        <Input
          type="text"
          value={value}
          onChange={e => onChange(e.target.value)}
          placeholder="comma-separated: go, rust"
        />
      );
    default:
      return (
        <Input
          type="text"
          value={value}
          onChange={e => onChange(e.target.value)}
          placeholder="value"
        />
      );
  }
}
