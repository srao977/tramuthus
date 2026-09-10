import { labMongoProvider } from "@/lib/lab-mongo-provider";
import { mongoPublicError } from "@/lib/mongo";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const runs = await labMongoProvider.listRuns();
    return Response.json({ runs }, { headers: { "Cache-Control": "no-store" } });
  } catch (error) {
    return Response.json({ error: mongoPublicError(error) }, { status: 503 });
  }
}
