import { publicMongoError } from "@/lib/mongo";
import { listRuns } from "@/lib/replay-provider";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET() {
  try {
    return Response.json({ runs: await listRuns() }, { headers: { "Cache-Control": "no-store" } });
  } catch (error) {
    return Response.json({ error: publicMongoError(error) }, { status: 503 });
  }
}