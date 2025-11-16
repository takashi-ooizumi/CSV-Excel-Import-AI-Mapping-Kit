import { NextResponse } from "next/server";
import { resolveSampleCsvPath } from "@/lib/sampleCsv";
import { promises as fs } from "fs";
import path from "path";

type Params = {
  filename: string;
};

export async function GET(_: Request, { params }: { params: Params }) {
  try {
    const filePath = resolveSampleCsvPath(params.filename);
    const fileBuffer = await fs.readFile(filePath);
    const arrayBuffer = fileBuffer.buffer.slice(
      fileBuffer.byteOffset,
      fileBuffer.byteOffset + fileBuffer.byteLength
    ) as ArrayBuffer;
    const baseName = path.basename(filePath);

    return new NextResponse(arrayBuffer, {
      headers: {
        "Content-Type": "text/csv; charset=utf-8",
        "Content-Disposition": `attachment; filename="${encodeURIComponent(baseName)}"`,
      },
    });
  } catch (error: any) {
    return NextResponse.json({ error: "File not found" }, { status: 404 });
  }
}
