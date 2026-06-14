import { CssBaseline, ThemeProvider } from "@mui/material";
import type { PaletteMode } from "@mui/material";
import type { PropsWithChildren } from "react";
import { createContext, useEffect, useMemo, useState } from "react";
import { createAppTheme } from "./theme";

type ThemeModeContextValue = {
  mode: PaletteMode;
  toggleThemeMode: () => void;
};

const themeModeStorageKey = "app-theme-mode";

const ThemeModeContext = createContext<ThemeModeContextValue | undefined>(
  undefined,
);

function getInitialThemeMode(): PaletteMode {
  if (typeof window === "undefined") {
    return "light";
  }

  const storedMode = window.localStorage.getItem(themeModeStorageKey);

  if (storedMode === "light" || storedMode === "dark") {
    return storedMode;
  }

  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function ThemeModeProvider({ children }: PropsWithChildren) {
  const [mode, setMode] = useState<PaletteMode>(getInitialThemeMode);
  const theme = useMemo(() => createAppTheme(mode), [mode]);

  useEffect(() => {
    window.localStorage.setItem(themeModeStorageKey, mode);
  }, [mode]);

  const contextValue = useMemo<ThemeModeContextValue>(
    () => ({
      mode,
      toggleThemeMode: () => {
        setMode((currentMode) => (currentMode === "light" ? "dark" : "light"));
      },
    }),
    [mode],
  );

  return (
    <ThemeModeContext.Provider value={contextValue}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </ThemeModeContext.Provider>
  );
}

export { ThemeModeContext };
