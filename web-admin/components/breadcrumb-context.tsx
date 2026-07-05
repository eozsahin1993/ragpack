"use client";

import { createContext, useContext, useState, useCallback } from "react";

type Labels = Record<string, string>;
type BreadcrumbContextValue = {
  labels: Labels;
  setLabel: (segment: string, label: string) => void;
};

const BreadcrumbContext = createContext<BreadcrumbContextValue>({
  labels: {},
  setLabel: () => {},
});

export function BreadcrumbProvider({ children }: { children: React.ReactNode }) {
  const [labels, setLabels] = useState<Labels>({});
  const setLabel = useCallback((segment: string, label: string) => {
    setLabels(prev => ({ ...prev, [segment]: label }));
  }, []);
  return (
    <BreadcrumbContext.Provider value={{ labels, setLabel }}>
      {children}
    </BreadcrumbContext.Provider>
  );
}

export function useBreadcrumbLabel() {
  return useContext(BreadcrumbContext).setLabel;
}

export function useBreadcrumbLabels() {
  return useContext(BreadcrumbContext).labels;
}
