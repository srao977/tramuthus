import { writeFileSync, mkdirSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import zlib from "node:zlib";

const root = join(dirname(fileURLToPath(import.meta.url)), "..", "public", "icons");
mkdirSync(root, { recursive: true });

function crc32(buf) {
  let c = ~0;
  for (const b of buf) {
    c ^= b;
    for (let k = 0; k < 8; k += 1) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
  }
  return ~c >>> 0;
}

function chunk(type, data) {
  const t = Buffer.from(type);
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length);
  const crcBuf = Buffer.concat([t, data]);
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(crcBuf));
  return Buffer.concat([len, t, data, crc]);
}

function png(size, paint) {
  const raw = Buffer.alloc((size * 4 + 1) * size);
  for (let y = 0; y < size; y += 1) {
    const row = y * (size * 4 + 1);
    raw[row] = 0;
    for (let x = 0; x < size; x += 1) {
      const [r, g, b, a] = paint(x, y, size);
      const i = row + 1 + x * 4;
      raw[i] = r;
      raw[i + 1] = g;
      raw[i + 2] = b;
      raw[i + 3] = a;
    }
  }
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(size, 0);
  ihdr.writeUInt32BE(size, 4);
  ihdr[8] = 8;
  ihdr[9] = 6;
  const idat = zlib.deflateSync(raw);
  return Buffer.concat([
    Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]),
    chunk("IHDR", ihdr),
    chunk("IDAT", idat),
    chunk("IEND", Buffer.alloc(0)),
  ]);
}

function waveIcon(size, maskable) {
  return png(size, (x, y, s) => {
    const pad = maskable ? s * 0.12 : 0;
    const nx = (x - pad) / (s - 2 * pad);
    const ny = (y - pad) / (s - 2 * pad);
    if (nx < 0 || ny < 0 || nx > 1 || ny > 1) return [29, 78, 74, 255];
    const bg = [29, 78, 74, 255];
    const paper = [244, 246, 245, 255];
    const ink = [20, 48, 46, 255];
    const gold = [232, 185, 49, 255];
    const dx = nx - 0.5;
    const dy = ny - 0.5;
    if (dx * dx + dy * dy > 0.48 * 0.48) return maskable ? bg : [0, 0, 0, 0];
    if (dx * dx + dy * dy > 0.42 * 0.42) return gold;
    const inside = paper;
    const seq = nx;
    const wave = 0.52 - 0.18 * Math.sin(seq * Math.PI * 3.2);
    const vol = 0.78 - 0.12 * Math.abs(Math.sin(seq * Math.PI * 6));
    if (Math.abs(ny - wave) < 0.035) return ink;
    if (ny > vol && ny < 0.86 && Math.abs(Math.sin(seq * Math.PI * 12)) > 0.35) return [29, 78, 74, 180];
    return inside;
  });
}

writeFileSync(join(root, "icon-192.png"), waveIcon(192, false));
writeFileSync(join(root, "icon-512.png"), waveIcon(512, false));
writeFileSync(join(root, "icon-maskable-512.png"), waveIcon(512, true));
