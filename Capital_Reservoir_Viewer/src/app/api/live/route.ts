import { subscribeToReservoirEvents } from "@/lib/grpc-live";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

export async function GET(request: Request) {
  const encoder = new TextEncoder();
  let stop = () => {};
  const stream = new ReadableStream({
    start(controller) {
      const send = (event: string, value: unknown) => controller.enqueue(encoder.encode(`event: ${event}\ndata: ${JSON.stringify(value)}\n\n`));
      stop = subscribeToReservoirEvents(undefined, (event) => send("reservoir", event), (error) => {
        send("fault", { message: error.message });
      });
      send("ready", { transport: "grpc-sse" });
      request.signal.addEventListener("abort", () => {
        stop();
        controller.close();
      }, { once: true });
    },
    cancel() { stop(); },
  });
  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache, no-transform",
      Connection: "keep-alive",
    },
  });
}