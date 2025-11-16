import { NextResponse } from "next/server";
import { listSampleCsvFiles } from "@/lib/sampleCsv";

export async function GET() {
  try {
    const files = await listSampleCsvFiles();
    return NextResponse.json({
      files: files.map((filename) => ({
        filename,
        downloadPath: `/api/sample-csv/${encodeURIComponent(filename)}`,
      })),
    });
  } catch (error: any) {
    return NextResponse.json(
      { error: error.message ?? "Failed to load sample CSVs" },
      { status: 500 }
    );
  }
}
