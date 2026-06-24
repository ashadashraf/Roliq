from __future__ import annotations

from pathlib import Path

import pymupdf
from PIL import Image, ImageDraw, ImageOps


ROOT = Path(__file__).resolve().parents[2]
FIXTURES = ROOT / "fixtures" / "resumes"
OUTPUT = ROOT / "tmp" / "pdfs"


def main() -> None:
    OUTPUT.mkdir(parents=True, exist_ok=True)
    previews: list[tuple[str, Image.Image]] = []
    for pdf_path in sorted(FIXTURES.glob("*/resume.pdf")):
        document = pymupdf.open(pdf_path)
        if document.page_count != 1:
            raise RuntimeError(f"Expected one-page fixture: {pdf_path}")
        pixmap = document[0].get_pixmap(matrix=pymupdf.Matrix(1.4, 1.4), alpha=False)
        image = Image.frombytes("RGB", (pixmap.width, pixmap.height), pixmap.samples)
        preview_path = OUTPUT / f"{pdf_path.parent.name}.png"
        image.save(preview_path)
        previews.append((pdf_path.parent.name, image))
        document.close()

    thumb_width = 500
    label_height = 34
    margin = 18
    cells: list[Image.Image] = []
    for label, image in previews:
        ratio = thumb_width / image.width
        resized = image.resize((thumb_width, int(image.height * ratio)), Image.Resampling.LANCZOS)
        cell = Image.new("RGB", (thumb_width + margin * 2, resized.height + label_height + margin * 2), "#e2e8f0")
        cell.paste(resized, (margin, label_height + margin))
        ImageDraw.Draw(cell).text((margin, 9), label.replace("_", " ").title(), fill="#0f172a")
        cells.append(ImageOps.expand(cell, border=1, fill="#94a3b8"))

    columns = 2
    rows = (len(cells) + columns - 1) // columns
    cell_width = max(cell.width for cell in cells)
    cell_height = max(cell.height for cell in cells)
    sheet = Image.new("RGB", (cell_width * columns, cell_height * rows), "white")
    for index, cell in enumerate(cells):
        sheet.paste(cell, ((index % columns) * cell_width, (index // columns) * cell_height))
    sheet.save(OUTPUT / "resume-fixtures-contact-sheet.png")
    print(OUTPUT / "resume-fixtures-contact-sheet.png")


if __name__ == "__main__":
    main()
