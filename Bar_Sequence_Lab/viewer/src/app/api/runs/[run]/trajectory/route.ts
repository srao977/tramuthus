import { labMongoProvider } from "@/lib/lab-mongo-provider";
import { mongoPublicError } from "@/lib/mongo";
import type { TrajectoryType } from "@/lib/types";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

const SYMBOL_RE = /^[A-Z0-9.-]{1,15}$/;

export async function GET(
  request: Request,
  context: { params: Promise<{ run: string }> },
) {
  const { run } = await context.params;
  const collectionRunId = decodeURIComponent(run ?? "").trim();
  const url = new URL(request.url);
  const symbol = (url.searchParams.get("symbol") ?? "").trim().toUpperCase();
  const typeParam = (url.searchParams.get("type") ?? "price").trim().toLowerCase();

  if (!collectionRunId) {
    return Response.json({ error: "collection_run_id is required" }, { status: 400 });
  }
  if (!SYMBOL_RE.test(symbol)) {
    return Response.json({ error: "Invalid symbol" }, { status: 400 });
  }
  if (typeParam !== "price" && typeParam !== "volume") {
    return Response.json({ error: "type must be price or volume" }, { status: 400 });
  }
  const trajectory = typeParam as TrajectoryType;

  try {
    const series = await labMongoProvider.getTrajectory(collectionRunId, symbol, trajectory);
    return Response.json(series, { headers: { "Cache-Control": "no-store" } });
  } catch (error) {
    return Response.json({ error: mongoPublicError(error) }, { status: 503 });
  }
}
