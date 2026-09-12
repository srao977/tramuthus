import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "DSE_JEH Proving Viewer",
    short_name: "DSE_JEH",
    description: "Observational viewer for deterministic DSE_JEH evidence",
    start_url: "/",
    display: "standalone",
    background_color: "#101719",
    theme_color: "#101719",
    icons: [{ src: "/favicon.ico", sizes: "any", type: "image/x-icon" }],
  };
}