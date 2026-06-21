import { Box, Paper, Typography } from "@mui/material";
import type { SxProps, Theme } from "@mui/material/styles";
import { alpha } from "@mui/material/styles";
import type { PropsWithChildren } from "react";

type SectionCardProps = PropsWithChildren<{
  title?: string;
  description?: string;
  sx?: SxProps<Theme>;
}>;

export function SectionCard({
  title,
  description,
  children,
  sx,
}: SectionCardProps) {
  return (
    <Paper
      sx={[
        {
          background: (theme) =>
            `radial-gradient(circle at top right, ${alpha(theme.palette.info.main, 0.08)} 0%, transparent 28%), linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.055)} 0%, ${alpha(theme.palette.info.main, 0.028)} 100%)`,
          border: "1px solid",
          borderColor: (theme) => alpha(theme.palette.primary.main, 0.13),
          borderRadius: 5,
          boxShadow: (theme) =>
            `0 18px 38px ${alpha(theme.palette.primary.main, 0.08)}`,
          overflow: "hidden",
          p: { md: 3.25, xs: 2.75 },
          position: "relative",
          "&::before": {
            background: (theme) =>
              `linear-gradient(90deg, ${alpha(theme.palette.primary.main, 0.95)} 0%, ${alpha(theme.palette.info.main, 0.4)} 100%)`,
            content: '""',
            height: 3,
            left: 0,
            position: "absolute",
            right: 0,
            top: 0,
          },
        },
        ...(Array.isArray(sx) ? sx : sx ? [sx] : []),
      ]}
    >
      <Box sx={{ display: "flex", flexDirection: "column", gap: 2.5 }}>
        {title ? (
          <Box
            sx={{
              borderBottom: "1px solid",
              borderBottomColor: (theme) => alpha(theme.palette.primary.main, 0.1),
              display: "flex",
              flexDirection: "column",
              gap: 0.75,
              pb: 1.75,
            }}
          >
            <Typography
              sx={{
                color: (theme) =>
                  theme.palette.mode === "dark"
                    ? theme.palette.primary.light
                    : theme.palette.primary.dark,
                fontWeight: 850,
                letterSpacing: "-0.02em",
              }}
              variant="h5"
            >
              {title}
            </Typography>
            {description ? (
              <Typography color="text.primary" sx={{ opacity: 0.9 }} variant="body2">
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
