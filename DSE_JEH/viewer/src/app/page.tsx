import Dashboard from "@/components/Dashboard";
import { listRuns, loadRun } from "@/lib/evidence-store";

export const dynamic = "force-dynamic";

async function readRun(runId: string) {
  try {
    return { run: await loadRun(runId), error: null };
  } catch (error) {
    return { run: null, error: error instanceof Error ? error.message : "Unknown evidence error" };
  }
}

export default async function Home({
  searchParams,
}: {
  searchParams: Promise<{ run?: string }>;
}) {
  const runs = await listRuns();
  const requested = (await searchParams).run;
  const selected = runs.find((run) => run.run_id === requested) ?? runs[0];

  if (!selected) {
    return (
      <main className="empty-page">
        <section className="empty-state">
          <p className="eyebrow">DSE_JEH · PROVING VIEWER</p>
          <h1>No proving evidence found</h1>
          <p>Generate a deterministic CSV proving run, then refresh this page.</p>
          <code>DSE_JEH_EVIDENCE_ROOT or ../output</code>
        </section>
      </main>
    );
  }

  const result = await readRun(selected.run_id);
  if (!result.run) {
    return (
      <main className="empty-page">
        <section className="empty-state error-state">
          <p className="eyebrow">READ-ONLY EVIDENCE ERROR</p>
          <h1>Run could not be loaded</h1>
          <p>{result.error}</p>
        </section>
      </main>
    );
  }

  return <Dashboard runs={runs} run={result.run} />;
}
