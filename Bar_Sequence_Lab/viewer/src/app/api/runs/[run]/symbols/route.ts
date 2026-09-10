import { labMongoProvider } from "@/lib/lab-mongo-provider";
import { mongoPublicError } from "@/lib/mongo";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET(
  request: Request,
  context: { params: Promise<{ run: string }> },
) {
  const { run } = await context.params;
  const collectionRunId = decodeURIComponent(run ?? "").trim();
  if (!collectionRunId) {
    return Response.json({ error: "collection_run_id is required" }, { status: 400 });
  }
  const group = new URL(request.url).searchParams.get("group")?.trim() || undefined;
  try {
    const symbols = await labMongoProvider.listSymbols(collectionRunId, group);
    return Response.json({ symbols }, { headers: { "Cache-Control": "no-store" } });
  } catch (error) {
    return Response.json({ error: mongoPublicError(error) }, { status: 503 });
  }
}
