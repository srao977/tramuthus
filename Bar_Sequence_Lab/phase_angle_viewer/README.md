# Bar Sequence Lab Phase Angle Viewer

Independent Next.js PWA for read-only inspection of persisted phase-angle evidence. It reads `bar_sequence_phase_angle_series` through server routes and retrieves individual raw observations from `bar_sequence` by lineage. It does not calculate phase, write MongoDB, interpolate null values, or implement replay or strategy semantics.

Server-only environment variables:

- `BAR_SEQ_LAB_MONGO_URI` (default `mongodb://127.0.0.1:27017`)
- `BAR_SEQ_LAB_MONGO_DB` (default `bar_sequence_db`)
- `BAR_SEQ_LAB_PHASE_COLLECTION` (default `bar_sequence_phase_angle_series`)
- `BAR_SEQ_LAB_RAW_COLLECTION` (default `bar_sequence`)

```powershell
npm install
npm test
npm run lint
npx tsc --noEmit
npm run build
npm run dev -- -H 127.0.0.1 -p 3001
```

Open http://127.0.0.1:3001.
