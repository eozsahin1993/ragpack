import { ReactNode } from "react";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";

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
    <Card className="py-0 [&_th]:h-11 [&_th]:px-4 [&_td]:px-4 [&_td]:py-3">
      <CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/40 hover:bg-muted/40">
              {columns.map((col, i) => (
                <TableHead
                  key={i}
                  className={cn("text-xs font-medium uppercase tracking-wide text-muted-foreground", col.className)}
                >
                  {col.label}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>{children}</TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
