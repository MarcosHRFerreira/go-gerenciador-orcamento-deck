import { Box, Paper, Typography } from "@mui/material";
import type { PropsWithChildren } from "react";

type SectionCardProps = PropsWithChildren<{
  title?: string;
  description?: string;
}>;

export function SectionCard({
  title,
  description,
  children,
}: SectionCardProps) {
  return (
    <Paper sx={{ borderRadius: 4, p: { md: 3, xs: 2.5 } }}>
      <Box sx={{ display: "flex", flexDirection: "column", gap: 2.5 }}>
        {title ? (
          <Box sx={{ display: "flex", flexDirection: "column", gap: 0.5 }}>
            <Typography variant="h5">{title}</Typography>
            {description ? (
              <Typography color="text.secondary" variant="body2">
                {description}
              </Typography>
            ) : null}
          </Box>
        ) : null}
        {children}
      </Box>
    </Paper>
  );
}
