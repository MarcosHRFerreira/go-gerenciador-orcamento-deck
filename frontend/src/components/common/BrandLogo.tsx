import { Box } from "@mui/material";
import type { SxProps, Theme } from "@mui/material/styles";

type BrandLogoProps = {
  alt?: string;
  imageSx?: SxProps<Theme>;
  wrapperSx?: SxProps<Theme>;
};

export default function BrandLogo({
  alt = "Deck Representacao Comercial",
  imageSx,
  wrapperSx,
}: BrandLogoProps) {
  return (
    <Box
      sx={[
        {
          alignItems: "center",
          backgroundColor: "#FFFFFF",
          borderRadius: 3,
          display: "inline-flex",
          justifyContent: "center",
          overflow: "hidden",
          p: 1.25,
        },
        ...(Array.isArray(wrapperSx) ? wrapperSx : [wrapperSx]),
      ]}
    >
      <Box
        alt={alt}
        component="img"
        src="/deck.jpg"
        sx={[
          {
            display: "block",
            height: "auto",
            maxWidth: "100%",
            width: 220,
          },
          ...(Array.isArray(imageSx) ? imageSx : [imageSx]),
        ]}
      />
    </Box>
  );
}
