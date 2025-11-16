import { promises as fs, statSync } from "fs";
import path from "path";

let cachedSampleDir: string | null = null;

function getSampleCsvDir() {
  if (cachedSampleDir) return cachedSampleDir;

  const dir = path.resolve(process.cwd(), "sample_csv");
  try {
    if (!statSync(dir).isDirectory()) {
      throw new Error("sample_csv is not a directory");
    }
    cachedSampleDir = dir;
    return dir;
  } catch (error) {
    throw new Error(`sample_csv directory not found under ${dir}`);
  }
}

export async function listSampleCsvFiles() {
  const dir = getSampleCsvDir();
  const entries = await fs.readdir(dir, { withFileTypes: true });
  return entries
    .filter((entry) => entry.isFile() && entry.name.endsWith(".csv"))
    .map((entry) => entry.name);
}

export function resolveSampleCsvPath(filename: string) {
  const dir = getSampleCsvDir();
  const decoded = decodeURIComponent(filename);
  const normalized = path.normalize(decoded);
  const resolved = path.resolve(dir, normalized);

  const safePrefix = dir.endsWith(path.sep) ? dir : dir + path.sep;
  if (!resolved.startsWith(safePrefix)) {
    throw new Error("Invalid path");
  }

  return resolved;
}
