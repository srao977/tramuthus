import { mongoPublicError } from "@/lib/mongo";
import { phaseMongoProvider } from "@/lib/phase-mongo-provider";
import {
  experimentFromSearch,
  validPartition,
  validSymbol,
  validateExperiment,
} from "@/lib/query";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const search = new URL(request.url).searchParams;
  const experiment = experimentFromSearch(search);
  const partitionId = (search.get("partitionId") ?? "").trim().toUpperCase();
  const symbol = (search.get("symbol") ?? "").trim().toUpperCase();
  const invalid = validateExperiment(experiment);
  if (invalid) return Response.json({ error: invalid }, { status: 400 });
  if (!validPartition(partitionId) || !validSymbol(symbol)) {
    return Response.json({ error: "Valid partitionId and symbol are required" }, { status: 400 });
  }
  try {
    return Response.json(
      await phaseMongoProvider.getSeries(experiment, partitionId, symbol),
      { headers: { "Cache-Control": "no-store" } },
    );
  } catch (error) {
    return Response.json({ error: mongoPublicError(error) }, { status: 503 });
  }
}
