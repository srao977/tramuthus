import { publicMongoError } from "@/lib/mongo";
import { readRun } from "@/lib/replay-provider";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET(_request: Request, context: { params: Promise<{ run: string }> }) {
  const { run } = await context.params;
  if (!run || run.length > 200) return Response.json({ error: "Invalid pipeline run ID." }, { status: 400 });
  try {
    return Response.json({ events: await readRun(run) }, { headers: { "Cache-Control": "no-store" } });
  } catch (error) {
    return Response.json({ error: publicMongoError(error) }, { status: 503 });
  }
}