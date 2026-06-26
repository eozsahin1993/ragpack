import { ReactNode } from "react";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface Column {
  label: string;
  className?: string;
}

interface DataTableProps {
  columns: Column[];
  children: ReactNode;
}

export function DataTable({ columns, children }: DataTableProps) {
  return (
    <div className="rounded-lg border bg-white overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow className="bg-zinc-50">
            {columns.map((col, i) => (
              <TableHead key={i} className={col.className}>
                {col.label}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>{children}</TableBody>
      </Table>
    </div>
  );
}
