"use client";

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";

export type IslandPhase = "idle" | "listening" | "routing" | "download";

export interface IslandState {
  phase: IslandPhase;
  label: string;
  detail?: string;
}

interface IslandContextValue {
  state: IslandState;
  setIsland: (state: IslandState) => void;
  pulseDownload: (platform: string) => void;
}

const defaultState: IslandState = { phase: "idle", label: "Condura" };

const IslandContext = createContext<IslandContextValue | null>(null);

export function IslandProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<IslandState>(defaultState);

  const setIsland = useCallback((next: IslandState) => setState(next), []);

  const pulseDownload = useCallback((platform: string) => {
    setState({
      phase: "download",
      label: "Preparing download",
      detail: platform,
    });
    window.setTimeout(() => setState(defaultState), 3200);
  }, []);

  const value = useMemo(
    () => ({ state, setIsland, pulseDownload }),
    [state, setIsland, pulseDownload]
  );

  return <IslandContext.Provider value={value}>{children}</IslandContext.Provider>;
}

export function useIsland() {
  const ctx = useContext(IslandContext);
  if (!ctx) throw new Error("useIsland must be used within IslandProvider");
  return ctx;
}
