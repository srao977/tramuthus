import { mongoPublicError } from "@/lib/mongo";
import { phaseMongoProvider } from "@/lib/phase-mongo-provider";
import { experimentFromSearch, validateExperiment } from "@/lib/query";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const experiment = experimentFromSearch(new URL(request.url).searchParams);
  const invalid = validateExperiment(experiment);
  if (invalid) return Response.json({ error: invalid }, { status: 400 });
  try {
    return Response.json(
      { experiment, symbols: await phaseMongoProvider.listSymbols(experiment) },
      { headers: { "Cache-Control": "no-store" } },
    );
  } catch (error) {
    return Response.json({ error: mongoPublicError(error) }, { status: 503 });
  }
}
