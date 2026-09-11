import { mongoPublicError } from "@/lib/mongo";
import { phaseMongoProvider } from "@/lib/phase-mongo-provider";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET() {
  try {
    return Response.json(
      { experiments: await phaseMongoProvider.listExperiments() },
      { headers: { "Cache-Control": "no-store" } },
    );
  } catch (error) {
    return Response.json({ error: mongoPublicError(error) }, { status: 503 });
  }
}
