import "server-only";

import path from "node:path";
import { credentials, loadPackageDefinition, type ClientReadableStream, type ServiceClientConstructor } from "@grpc/grpc-js";
import { loadSync } from "@grpc/proto-loader";
import { normalizeReservoirEvent, type CapitalReservoirEvent } from "./capital-reservoir";

type EvidenceResponse = { evidence?: { capitalReservoirEvent?: unknown } };
type RuntimeEvidenceClient = InstanceType<ServiceClientConstructor> & {
  SubscribeRuntimeEvidence(
    request: { runtimeId: string; afterPublicationSequence?: string },
  ): ClientReadableStream<EvidenceResponse>;
};

function protoPath(): string {
  return process.env.DSE_JEH_PROTO_PATH?.trim() || path.resolve(process.cwd(), "../DSE_JEH/api/proto/dse_jeh/v1/DSE_JEH_TransSat_1.proto");
}

export function subscribeToReservoirEvents(
  afterPublicationSequence: number | undefined,
  onEvent: (event: CapitalReservoirEvent) => void,
  onError: (error: Error) => void,
): () => void {
  const definition = loadSync(protoPath(), {
    keepCase: false,
    longs: String,
    enums: String,
    defaults: false,
    oneofs: true,
  });
  const root = loadPackageDefinition(definition) as Record<string, unknown>;
  const dseJeh = root.dse_jeh as Record<string, unknown>;
  const version = dseJeh.v1 as Record<string, unknown>;
  const Constructor = version.RuntimeEvidenceService as ServiceClientConstructor;
  const address = process.env.DSE_JEH_SERVER_ADDRESS?.trim() || "127.0.0.1:50052";
  const client = new Constructor(address, credentials.createInsecure()) as RuntimeEvidenceClient;
  const stream = client.SubscribeRuntimeEvidence({
    runtimeId: process.env.DSE_JEH_RUNTIME_ID?.trim() || "",
    ...(afterPublicationSequence ? { afterPublicationSequence: String(afterPublicationSequence) } : {}),
  });
  stream.on("data", (response: EvidenceResponse) => {
    if (!response.evidence?.capitalReservoirEvent) return;
    try {
      onEvent(normalizeReservoirEvent(response.evidence.capitalReservoirEvent));
    } catch (error) {
      onError(error instanceof Error ? error : new Error("Invalid live Capital Reservoir event"));
    }
  });
  stream.on("error", onError);
  return () => {
    stream.cancel();
    client.close();
  };
}