import { Box, Typography } from "@mui/material";
import { alpha } from "@mui/material/styles";
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
          background: (theme) =>
            `radial-gradient(circle at top right, ${alpha(theme.palette.info.main, 0.12)} 0%, transparent 32%), linear-gradient(135deg, ${alpha(theme.palette.primary.main, 0.09)} 0%, ${alpha(theme.palette.info.main, 0.045)} 100%)`,
          border: "1px solid",
          borderColor: (theme) => alpha(theme.palette.primary.main, 0.16),
          borderRadius: 5,
          boxShadow: (theme) =>
            `0 18px 40px ${alpha(theme.palette.primary.main, 0.1)}`,
          display: "flex",
          flexDirection: { md: "row", xs: "column" },
          gap: 2,
          justifyContent: "space-between",
          overflow: "hidden",
          p: { md: 3.5, xs: 2.75 },
          position: "relative",
          "&::before": {
            background: (theme) => {
              const accentColor =
                theme.palette.mode === "dark"
                  ? theme.palette.primary.light
                  : theme.palette.primary.dark;

              return `linear-gradient(90deg, ${accentColor} 0%, ${alpha(accentColor, 0.2)} 100%)`;
            },
            content: '""',
            height: 4,
            left: 0,
            position: "absolute",
            right: 0,
            top: 0,
          },
        }}
      >
        <Box sx={{ flex: "1 1 auto", minWidth: 0, pr: { md: 2, xs: 0 } }}>
          <Typography
            sx={{
              color: (theme) =>
                theme.palette.mode === "dark"
                  ? theme.palette.primary.light
                  : theme.palette.primary.dark,
              fontWeight: 850,
              letterSpacing: "-0.03em",
              lineHeight: 1.05,
            }}
            variant="h3"
          >
            {title}
          </Typography>
          {description ? (
            <Typography
              color="text.primary"
              sx={{
                maxWidth: 760,
                mt: 1.25,
                opacity: 0.92,
              }}
              variant="body1"
            >
              {description}
            </Typography>
          ) : null}
        </Box>
        {action ? (
          <Box
            sx={{
              backgroundColor: (theme) =>
                alpha(theme.palette.common.white, 0.5),
              border: "1px solid",
              borderColor: (theme) => alpha(theme.palette.primary.main, 0.12),
              borderRadius: 4,
              boxShadow: (theme) =>
                `0 10px 24px ${alpha(theme.palette.primary.main, 0.06)}`,
              flexShrink: 0,
              p: { md: 1.25, xs: 1 },
              width: { md: "auto", xs: "100%" },
            }}
          >
            {action}
          </Box>
        ) : null}
      </Box>
      {children}
    </Box>
  );
}
