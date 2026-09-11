"use client";

import { useEffect } from "react";

export default function PwaRegister() {
  useEffect(() => {
    if (!("serviceWorker" in navigator)) return;

    if (process.env.NODE_ENV !== "production") {
      Promise.all([
        navigator.serviceWorker.getRegistrations().then((registrations) =>
          Promise.all(registrations.map((registration) => registration.unregister())),
        ),
        caches.keys().then((keys) =>
          Promise.all(keys.filter((key) => key.startsWith("bar-seq-phase-viewer-")).map((key) => caches.delete(key))),
        ),
      ]).catch(() => {
        /* development cleanup is best-effort */
      });
      return;
    }

    navigator.serviceWorker.register("/sw.js").catch(() => {
      /* installability is best-effort on HTTP LAN */
    });
  }, []);
  return null;
}
