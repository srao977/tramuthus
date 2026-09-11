import { mongoPublicError } from "@/lib/mongo";
import { phaseMongoProvider } from "@/lib/phase-mongo-provider";
import { validSymbol } from "@/lib/query";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const search = new URL(request.url).searchParams;
  const collectionRunId = (search.get("collectionRunId") ?? "").trim();
  const symbol = (search.get("symbol") ?? "").trim().toUpperCase();
  const generatorSequenceNo = Number(search.get("generatorSequenceNo"));
  if (!collectionRunId || !validSymbol(symbol) || !Number.isInteger(generatorSequenceNo) || generatorSequenceNo <= 0) {
    return Response.json({ error: "Valid lineage parameters are required" }, { status: 400 });
  }
  try {
    const observation = await phaseMongoProvider.getRawObservation(
      collectionRunId,
      symbol,
      generatorSequenceNo,
    );
    if (!observation) return Response.json({ error: "Raw observation not found" }, { status: 404 });
    return Response.json({ observation }, { headers: { "Cache-Control": "no-store" } });
  } catch (error) {
    return Response.json({ error: mongoPublicError(error) }, { status: 503 });
  }
}
