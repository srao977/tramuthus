# Bar Sequence Lab viewer

Portable Next.js PWA for raw bar-sequence waves. Independent of Fin_FeedSat_1_Viewer and DSE_TransSat_1_viewer.

The ordered bar sequence is the wave. X = `generator_sequence_no` (BAR-INDEX REPLAY). Price and Volume are separate. Primary chart is **klinecharts** with four independently scaled presentation slots. No maths / Ehlers / Hop.

Server-only env (never `NEXT_PUBLIC_*`): `BAR_SEQ_LAB_MONGO_URI`, `BAR_SEQ_LAB_MONGO_DB`, `BAR_SEQ_LAB_MONGO_COLLECTION`. Defaults: `mongodb://127.0.0.1:27017`, `bar_sequence_db`, `bar_sequence`.

```
npm install
npm test
npm run build
npm run start -- -H 0.0.0.0 -p 3000
```

Open http://127.0.0.1:3000

- V0.1 PWA + Mongo: [docs/BAR_SEQUENCE_LAB_PWA_VIEWER_IMPLEMENTATION_V0_1_091026.md](docs/BAR_SEQUENCE_LAB_PWA_VIEWER_IMPLEMENTATION_V0_1_091026.md)
- Multi-wave klinecharts replay: [docs/BAR_SEQUENCE_LAB_MULTI_WAVE_KLINECHARTS_REPLAY_091026.md](docs/BAR_SEQUENCE_LAB_MULTI_WAVE_KLINECHARTS_REPLAY_091026.md)

## Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
