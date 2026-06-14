import { Box, Typography } from "@mui/material";
import type { PropsWithChildren, ReactNode } from "react";

type PageHeaderProps = PropsWithChildren<{
  title: string;
  description?: string;
  action?: ReactNode;
}>;

export function PageHeader({
  title,
  description,
  action,
  children,
}: PageHeaderProps) {
  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
      <Box
        sx={{
          alignItems: { md: "center", xs: "flex-start" },
          display: "flex",
          flexDirection: { md: "row", xs: "column" },
          gap: 2,
          justifyContent: "space-between",
        }}
      >
        <Box>
          <Typography variant="h3">{title}</Typography>
          {description ? (
            <Typography color="text.secondary" sx={{ mt: 1 }} variant="body1">
              {description}
            </Typography>
          ) : null}
        </Box>
        {action}
      </Box>
      {children}
    </Box>
  );
}
