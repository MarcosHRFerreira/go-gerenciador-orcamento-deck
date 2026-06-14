import "@fontsource/inter/400.css";
import "@fontsource/inter/500.css";
import "@fontsource/inter/600.css";
import "@fontsource/inter/700.css";

import { QueryClientProvider } from "@tanstack/react-query";
import type { PropsWithChildren } from "react";
import { AuthProvider } from "../../features/auth/context/AuthContext";
import { queryClient } from "../../lib/query/queryClient";
import { ThemeModeProvider } from "../theme/ThemeModeContext";

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeModeProvider>
        <AuthProvider>{children}</AuthProvider>
      </ThemeModeProvider>
    </QueryClientProvider>
  );
}
