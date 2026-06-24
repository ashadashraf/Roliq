from __future__ import annotations

import argparse
import hashlib
import json
import textwrap
import uuid
from pathlib import Path
from typing import Any

from reportlab.lib.colors import HexColor
from reportlab.lib.pagesizes import LETTER
from reportlab.pdfgen import canvas


ROOT = Path(__file__).resolve().parents[2]
FIXTURE_ROOT = ROOT / "fixtures" / "resumes"
CREATED_AT = "2026-01-01T00:00:00Z"
NAMESPACE = uuid.UUID("83d60c1d-e690-4fac-9fe2-2bf6ad4ba675")


SCENARIOS: tuple[dict[str, Any], ...] = (
    {
        "id": "senior_backend_engineer",
        "name": "Avery Sen",
        "headline": "Senior Backend Engineer",
        "email": "avery.sen@example.com",
        "location": "Bengaluru, India",
        "summary": "Backend engineer with nine years of experience building reliable multi-tenant platforms and event-driven services.",
        "skills": ["Go", "PostgreSQL", "Kubernetes", "Kafka", "AWS", "OpenTelemetry"],
        "experience": [
            {
                "company": "Northstar Systems",
                "title": "Senior Backend Engineer",
                "start": "2021-04",
                "end": None,
                "range": "Apr 2021 - Present",
                "highlights": [
                    "Led a Go platform serving 18 million monthly API requests with a 99.95 percent availability target.",
                    "Reduced event-processing latency by 42 percent through partition redesign and idempotent consumers.",
                ],
            },
            {
                "company": "Harbor Stack",
                "title": "Backend Engineer",
                "start": "2017-06",
                "end": "2021-03",
                "range": "Jun 2017 - Mar 2021",
                "highlights": [
                    "Designed PostgreSQL tenancy controls and audit trails for regulated customer workflows.",
                ],
            },
        ],
        "education": [{"institution": "Riverdale Institute of Technology", "degree": "BTech", "field": "Computer Science", "end": "2017"}],
        "projects": [],
        "certifications": [{"name": "Cloud Architecture Professional", "issuer": "Synthetic Cloud Foundation", "issued": "2023-09"}],
        "languages": [("English", "Professional"), ("Hindi", "Native")],
        "links": [("Portfolio", "https://example.com/avery-sen")],
        "layout_tags": ["dense", "reverse-chronological", "metrics", "current-role"],
        "evaluation_emphasis": ["multiple roles", "current employment", "quantified highlights", "certification"],
    },
    {
        "id": "frontend_engineer",
        "name": "Maya Torres",
        "headline": "Frontend Engineer",
        "email": "maya.torres@example.com",
        "location": "Madrid, Spain",
        "summary": "Frontend engineer focused on accessible design systems, performance, and maintainable product interfaces.",
        "skills": ["TypeScript", "React", "Next.js", "CSS", "Accessibility", "Playwright"],
        "experience": [
            {
                "company": "Lumen Commerce",
                "title": "Frontend Engineer",
                "start": "2022-02",
                "end": None,
                "range": "Feb 2022 - Present",
                "highlights": [
                    "Built an accessible component library adopted by six product teams.",
                    "Improved largest contentful paint from 3.4 seconds to 1.8 seconds on the checkout flow.",
                ],
            },
            {
                "company": "Brightside Labs",
                "title": "Junior Web Developer",
                "start": "2020-07",
                "end": "2022-01",
                "range": "Jul 2020 - Jan 2022",
                "highlights": ["Created tested React interfaces from product and research specifications."],
            },
        ],
        "education": [{"institution": "Central Design University", "degree": "BSc", "field": "Interactive Media", "end": "2020"}],
        "projects": [{"name": "Contrast Compass", "description": "Open-source color contrast review utility.", "technologies": ["TypeScript", "React"]}],
        "certifications": [],
        "languages": [("Spanish", "Native"), ("English", "Professional")],
        "links": [("GitHub", "https://example.com/maya-torres")],
        "layout_tags": ["portfolio-link", "performance-metrics", "accessibility"],
        "evaluation_emphasis": ["technology punctuation", "performance numbers", "project extraction"],
    },
    {
        "id": "data_scientist",
        "name": "Jordan Kim",
        "headline": "Data Scientist",
        "email": "jordan.kim@example.com",
        "location": "Toronto, Canada",
        "summary": "Data scientist translating ambiguous business questions into monitored statistical and machine-learning products.",
        "skills": ["Python", "SQL", "PyTorch", "scikit-learn", "Experiment Design", "Data Visualization"],
        "experience": [
            {
                "company": "Atlas Health Analytics",
                "title": "Data Scientist",
                "start": "2021-08",
                "end": None,
                "range": "Aug 2021 - Present",
                "highlights": [
                    "Developed demand forecasts that reduced weekly inventory error by 17 percent.",
                    "Introduced model monitoring for drift, calibration, and subgroup performance.",
                ],
            },
            {
                "company": "Cedar Insights",
                "title": "Data Analyst",
                "start": "2019-05",
                "end": "2021-07",
                "range": "May 2019 - Jul 2021",
                "highlights": ["Designed controlled experiments and executive dashboards for retention programs."],
            },
        ],
        "education": [{"institution": "North Lake University", "degree": "MSc", "field": "Statistics", "end": "2019"}],
        "projects": [{"name": "Forecast Reliability Kit", "description": "Synthetic benchmark for interval calibration.", "technologies": ["Python", "PyTorch"]}],
        "certifications": [],
        "languages": [("English", "Native"), ("Korean", "Conversational")],
        "links": [("Research notes", "https://example.com/jordan-kim")],
        "layout_tags": ["scientific-terms", "mixed-case-skills", "metrics"],
        "evaluation_emphasis": ["hyphenated skills", "model terminology", "education degree"],
    },
    {
        "id": "fresh_graduate",
        "name": "Nia Patel",
        "headline": "Graduate Software Engineer",
        "email": "nia.patel@example.com",
        "location": "Pune, India",
        "summary": "Computer science graduate with internship and project experience in web services, testing, and data structures.",
        "skills": ["Java", "Python", "Git", "REST APIs", "Unit Testing"],
        "experience": [
            {
                "company": "Maple Byte Studio",
                "title": "Software Engineering Intern",
                "start": "2025-01",
                "end": "2025-06",
                "range": "Jan 2025 - Jun 2025",
                "highlights": ["Added API validation and unit tests for a campus scheduling product."],
            }
        ],
        "education": [{"institution": "Western Valley College", "degree": "BEng", "field": "Computer Science", "end": "2025", "grade": "8.6 GPA"}],
        "projects": [{"name": "Study Circle", "description": "Peer study-group scheduling application.", "technologies": ["Java", "REST APIs"]}],
        "certifications": [],
        "languages": [("English", "Professional"), ("Marathi", "Native")],
        "links": [("Projects", "https://example.com/nia-patel")],
        "layout_tags": ["education-first", "single-internship", "project-heavy"],
        "evaluation_emphasis": ["limited experience", "grade", "internship versus full-time role"],
    },
    {
        "id": "product_manager",
        "name": "Elias Morgan",
        "headline": "Senior Product Manager",
        "email": "elias.morgan@example.com",
        "location": "Dublin, Ireland",
        "summary": "Product manager leading discovery, delivery, and measurable adoption for B2B financial software.",
        "skills": ["Product Strategy", "Customer Research", "Roadmapping", "SQL", "Experimentation", "Stakeholder Management"],
        "experience": [
            {
                "company": "Orbit Payments",
                "title": "Senior Product Manager",
                "start": "2021-11",
                "end": None,
                "range": "Nov 2021 - Present",
                "highlights": [
                    "Launched reconciliation workflows that reached 64 percent customer adoption in two quarters.",
                    "Led quarterly planning across engineering, design, operations, and compliance.",
                ],
            },
            {
                "company": "Fieldnote Software",
                "title": "Product Manager",
                "start": "2018-03",
                "end": "2021-10",
                "range": "Mar 2018 - Oct 2021",
                "highlights": ["Established customer interview and experiment-review practices for three product areas."],
            },
        ],
        "education": [{"institution": "Eastborough University", "degree": "BA", "field": "Economics", "end": "2017"}],
        "projects": [],
        "certifications": [{"name": "Product Discovery Practitioner", "issuer": "Synthetic Product Guild", "issued": "2020-05"}],
        "languages": [("English", "Native")],
        "links": [("Case studies", "https://example.com/elias-morgan")],
        "layout_tags": ["business-language", "cross-functional", "adoption-metrics"],
        "evaluation_emphasis": ["nontechnical skills", "adoption metric", "certification"],
    },
    {
        "id": "career_switcher",
        "name": "Rowan Brooks",
        "headline": "Junior Data Analyst",
        "email": "rowan.brooks@example.com",
        "location": "Manchester, United Kingdom",
        "summary": "Former mathematics teacher moving into data analysis with applied SQL, Python, and dashboard project experience.",
        "skills": ["SQL", "Python", "Excel", "Power BI", "Data Cleaning", "Presentation"],
        "experience": [
            {
                "company": "Riverside Community School",
                "title": "Mathematics Teacher",
                "start": "2018-09",
                "end": "2024-08",
                "range": "Sep 2018 - Aug 2024",
                "highlights": [
                    "Analyzed assessment trends for 240 students and presented intervention recommendations.",
                    "Coordinated curriculum planning across a team of seven teachers.",
                ],
            }
        ],
        "education": [{"institution": "Moorland University", "degree": "BSc", "field": "Mathematics", "end": "2018"}],
        "projects": [{"name": "Transit Delay Explorer", "description": "Cleaned public transport data and built an interactive delay dashboard.", "technologies": ["Python", "Power BI"]}],
        "certifications": [{"name": "Applied Data Analysis Certificate", "issuer": "Northbridge Learning", "issued": "2025-02"}],
        "languages": [("English", "Native")],
        "links": [("Data portfolio", "https://example.com/rowan-brooks")],
        "layout_tags": ["career-transition", "transferable-skills", "project-and-certificate"],
        "evaluation_emphasis": ["do not invent analyst employment", "transferable experience", "new certification"],
    },
)


def stable_id(scenario: str, label: str) -> str:
    return str(uuid.uuid5(NAMESPACE, f"{scenario}:{label}"))


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def warning_free_fact(value: Any, source_hash: str, excerpt: str, *, normalized: str | None = None) -> dict[str, Any]:
    fact: dict[str, Any] = {
        "value": value,
        "origin": "extracted",
        "confidence": 1,
        "reviewState": "unreviewed",
        "evidence": [{"sourceSha256": source_hash, "page": 1, "excerpt": excerpt}],
        "warnings": [],
    }
    if normalized is not None:
        fact["normalizedValue"] = normalized
    return fact


def date_fact(value: str, source_hash: str, excerpt: str) -> dict[str, Any]:
    precision = {4: "year", 7: "month", 10: "day"}[len(value)]
    fact = warning_free_fact(value, source_hash, excerpt)
    fact["precision"] = precision
    return fact


def draw_wrapped(pdf: canvas.Canvas, text: str, x: float, y: float, width_chars: int = 108, leading: float = 10) -> float:
    for line in textwrap.wrap(text, width=width_chars, break_long_words=False, break_on_hyphens=False) or [""]:
        pdf.drawString(x, y, line)
        y -= leading
    return y


def render_resume(scenario: dict[str, Any], output: Path) -> list[str]:
    width, height = LETTER
    pdf = canvas.Canvas(str(output), pagesize=LETTER, invariant=1, pageCompression=1)
    pdf.setTitle(f"Synthetic resume fixture - {scenario['id']}")
    pdf.setAuthor("Roliq synthetic fixture generator")
    navy = HexColor("#14213D")
    blue = HexColor("#2563EB")
    muted = HexColor("#475569")
    x = 48
    y = height - 52
    lines: list[str] = []

    pdf.setFillColor(navy)
    pdf.setFont("Helvetica-Bold", 20)
    pdf.drawString(x, y, scenario["name"])
    lines.append(scenario["name"])
    y -= 22
    pdf.setFont("Helvetica-Bold", 11)
    pdf.setFillColor(blue)
    pdf.drawString(x, y, scenario["headline"])
    lines.append(scenario["headline"])
    y -= 17
    pdf.setFont("Helvetica", 8.5)
    pdf.setFillColor(muted)
    contact = f"{scenario['email']}  |  {scenario['location']}"
    pdf.drawString(x, y, contact)
    lines.extend([scenario["email"], scenario["location"]])
    y -= 23

    def heading(title: str) -> None:
        nonlocal y
        pdf.setFillColor(navy)
        pdf.setFont("Helvetica-Bold", 9.5)
        pdf.drawString(x, y, title.upper())
        pdf.setStrokeColor(blue)
        pdf.setLineWidth(0.8)
        pdf.line(x + 85, y + 2, width - 48, y + 2)
        y -= 14

    pdf.setFillColor(HexColor("#1E293B"))
    pdf.setFont("Helvetica", 8.5)
    heading("Profile")
    pdf.setFont("Helvetica", 8.5)
    y = draw_wrapped(pdf, scenario["summary"], x, y)
    lines.append(scenario["summary"])
    y -= 6

    heading("Skills")
    pdf.setFont("Helvetica", 8.5)
    skills_line = " | ".join(scenario["skills"])
    y = draw_wrapped(pdf, skills_line, x, y)
    lines.extend(scenario["skills"])
    y -= 5

    heading("Experience")
    for role in scenario["experience"]:
        pdf.setFillColor(navy)
        pdf.setFont("Helvetica-Bold", 9)
        role_line = f"{role['title']} - {role['company']}"
        pdf.drawString(x, y, role_line)
        lines.extend([role["title"], role["company"]])
        pdf.setFillColor(muted)
        pdf.setFont("Helvetica", 8)
        pdf.drawRightString(width - 48, y, role["range"])
        lines.append(role["range"])
        y -= 12
        pdf.setFillColor(HexColor("#1E293B"))
        pdf.setFont("Helvetica", 8.2)
        for highlight in role["highlights"]:
            pdf.drawString(x + 4, y, "-")
            y = draw_wrapped(pdf, highlight, x + 14, y, width_chars=102, leading=9)
            lines.append(highlight)
        y -= 3

    heading("Education")
    pdf.setFont("Helvetica", 8.5)
    for education in scenario["education"]:
        grade = f" | {education['grade']}" if education.get("grade") else ""
        education_line = f"{education['degree']} {education['field']} - {education['institution']} | {education['end']}{grade}"
        y = draw_wrapped(pdf, education_line, x, y)
        lines.extend([education["degree"], education["field"], education["institution"], education["end"]])
        if education.get("grade"):
            lines.append(education["grade"])
    y -= 4

    if scenario["projects"]:
        heading("Projects")
        pdf.setFont("Helvetica", 8.5)
        for project in scenario["projects"]:
            project_line = f"{project['name']} - {project['description']} Technologies: {', '.join(project['technologies'])}"
            y = draw_wrapped(pdf, project_line, x, y)
            lines.extend([project["name"], project["description"], *project["technologies"]])
        y -= 4

    if scenario["certifications"]:
        heading("Certifications")
        pdf.setFont("Helvetica", 8.5)
        for certification in scenario["certifications"]:
            certification_line = f"{certification['name']} - {certification['issuer']} | {certification['issued']}"
            y = draw_wrapped(pdf, certification_line, x, y)
            lines.extend([certification["name"], certification["issuer"], certification["issued"]])
        y -= 4

    heading("Languages and links")
    pdf.setFont("Helvetica", 8.2)
    language_line = "Languages: " + ", ".join(f"{name} ({level})" for name, level in scenario["languages"])
    link_line = " | ".join(f"{label}: {url}" for label, url in scenario["links"])
    y = draw_wrapped(pdf, language_line, x, y)
    y = draw_wrapped(pdf, link_line, x, y)
    for name, level in scenario["languages"]:
        lines.extend([name, level])
    for label, url in scenario["links"]:
        lines.extend([label, url])

    pdf.setFont("Helvetica-Oblique", 6.5)
    pdf.setFillColor(HexColor("#64748B"))
    pdf.drawString(x, 25, "Synthetic fixture created for Roliq contract and evaluation tests. No real person or employer is represented.")
    pdf.save()
    return lines


def build_expected(scenario: dict[str, Any], source_hash: str) -> dict[str, Any]:
    scenario_id = scenario["id"]
    candidate = {
        "fullName": warning_free_fact(scenario["name"], source_hash, scenario["name"]),
        "headline": warning_free_fact(scenario["headline"], source_hash, scenario["headline"]),
        "email": warning_free_fact(scenario["email"], source_hash, scenario["email"]),
        "location": warning_free_fact(scenario["location"], source_hash, scenario["location"]),
    }
    skills = [
        {"id": stable_id(scenario_id, f"skill:{index}"), "name": warning_free_fact(skill, source_hash, skill)}
        for index, skill in enumerate(scenario["skills"], start=1)
    ]
    experience = []
    for index, role in enumerate(scenario["experience"], start=1):
        item: dict[str, Any] = {
            "id": stable_id(scenario_id, f"experience:{index}"),
            "company": warning_free_fact(role["company"], source_hash, role["company"]),
            "title": warning_free_fact(role["title"], source_hash, role["title"]),
            "startDate": date_fact(role["start"], source_hash, role["range"]),
            "isCurrent": warning_free_fact(role["end"] is None, source_hash, role["range"]),
            "highlights": [warning_free_fact(value, source_hash, value) for value in role["highlights"]],
        }
        if role["end"]:
            item["endDate"] = date_fact(role["end"], source_hash, role["range"])
        experience.append(item)
    education = []
    for index, value in enumerate(scenario["education"], start=1):
        item = {
            "id": stable_id(scenario_id, f"education:{index}"),
            "institution": warning_free_fact(value["institution"], source_hash, value["institution"]),
            "degree": warning_free_fact(value["degree"], source_hash, value["degree"]),
            "fieldOfStudy": warning_free_fact(value["field"], source_hash, value["field"]),
            "endDate": date_fact(value["end"], source_hash, value["end"]),
        }
        if value.get("grade"):
            item["grade"] = warning_free_fact(value["grade"], source_hash, value["grade"])
        education.append(item)
    projects = []
    for index, value in enumerate(scenario["projects"], start=1):
        projects.append(
            {
                "id": stable_id(scenario_id, f"project:{index}"),
                "name": warning_free_fact(value["name"], source_hash, value["name"]),
                "description": warning_free_fact(value["description"], source_hash, value["description"]),
                "highlights": [],
                "technologies": [warning_free_fact(item, source_hash, item) for item in value["technologies"]],
            }
        )
    certifications = []
    for index, value in enumerate(scenario["certifications"], start=1):
        certifications.append(
            {
                "id": stable_id(scenario_id, f"certification:{index}"),
                "name": warning_free_fact(value["name"], source_hash, value["name"]),
                "issuer": warning_free_fact(value["issuer"], source_hash, value["issuer"]),
                "issuedDate": date_fact(value["issued"], source_hash, value["issued"]),
            }
        )
    languages = [
        {
            "id": stable_id(scenario_id, f"language:{index}"),
            "name": warning_free_fact(name, source_hash, name),
            "proficiency": warning_free_fact(level, source_hash, level),
        }
        for index, (name, level) in enumerate(scenario["languages"], start=1)
    ]
    links = [
        {
            "id": stable_id(scenario_id, f"link:{index}"),
            "label": warning_free_fact(label, source_hash, label),
            "url": warning_free_fact(url, source_hash, url),
        }
        for index, (label, url) in enumerate(scenario["links"], start=1)
    ]
    return {
        "schemaVersion": "resume-document.v1",
        "documentId": stable_id(scenario_id, "document"),
        "createdAt": CREATED_AT,
        "source": {
            "resumeId": stable_id(scenario_id, "resume"),
            "sourceVersionId": stable_id(scenario_id, "source-version"),
            "fileObjectId": stable_id(scenario_id, "file-object"),
            "contentSha256": source_hash,
            "mediaType": "application/pdf",
            "pageCount": 1,
        },
        "candidate": candidate,
        "summary": warning_free_fact(scenario["summary"], source_hash, scenario["summary"]),
        "skills": skills,
        "experience": experience,
        "education": education,
        "projects": projects,
        "certifications": certifications,
        "languages": languages,
        "links": links,
        "review": {"state": "needs_review", "revision": 1, "warnings": []},
    }


def notes(scenario: dict[str, Any]) -> str:
    emphasis = "\n".join(f"- {item}" for item in scenario["evaluation_emphasis"])
    tags = ", ".join(scenario["layout_tags"])
    return f"""# Evaluation notes: {scenario['id']}

This is an entirely synthetic resume fixture. The person, employers, institutions, achievements, and links are fictional and exist only for Roliq testing.

## Scenario

- Candidate archetype: {scenario['headline']}
- Layout tags: {tags}
- Source format: deterministic one-page PDF

## Evaluation emphasis

{emphasis}

## Review expectations

- Every extracted fact must have page-1 evidence matching visible source text.
- No employment, skill, education, certification, or date may be inferred beyond the PDF.
- Contact values use `example.com` and must never be treated as deliverable addresses.
- The expected document remains `needs_review`; fixture truth does not bypass product review policy.
"""


def generate() -> None:
    FIXTURE_ROOT.mkdir(parents=True, exist_ok=True)
    manifest = {"schemaVersion": 1, "syntheticOnly": True, "fixtures": []}
    for scenario in SCENARIOS:
        directory = FIXTURE_ROOT / scenario["id"]
        directory.mkdir(parents=True, exist_ok=True)
        pdf_path = directory / "resume.pdf"
        source_lines = render_resume(scenario, pdf_path)
        source_hash = sha256(pdf_path)
        expected = build_expected(scenario, source_hash)
        metadata = {
            "fixtureId": scenario["id"],
            "synthetic": True,
            "sourceFile": "resume.pdf",
            "sourceSha256": source_hash,
            "pageCount": 1,
            "layoutTags": scenario["layout_tags"],
            "evaluationEmphasis": scenario["evaluation_emphasis"],
            "metricWeights": {
                "experience": 1,
                "skills": 1,
                "education": 1,
                "certifications": 1,
                "dates": 1,
                "hallucination": 2,
                "evidenceCoverage": 2,
            },
        }
        ocr = {
            "fixtureId": scenario["id"],
            "sourceSha256": source_hash,
            "pages": [{"pageNumber": 1, "text": "\n".join(source_lines), "confidence": 1}],
        }
        (directory / "expected_resume_document_v1.json").write_text(
            json.dumps(expected, indent=2, sort_keys=True) + "\n", encoding="utf-8"
        )
        (directory / "evaluation_metadata.json").write_text(
            json.dumps(metadata, indent=2, sort_keys=True) + "\n", encoding="utf-8"
        )
        (directory / "fake_ocr_output.json").write_text(
            json.dumps(ocr, indent=2, sort_keys=True) + "\n", encoding="utf-8"
        )
        (directory / "evaluation_notes.md").write_text(notes(scenario), encoding="utf-8")
        manifest["fixtures"].append(
            {
                "id": scenario["id"],
                "directory": scenario["id"],
                "sourceSha256": source_hash,
                "expectedDocument": "expected_resume_document_v1.json",
                "fakeOcrOutput": "fake_ocr_output.json",
            }
        )
    (FIXTURE_ROOT / "manifest.json").write_text(
        json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate deterministic synthetic resume fixtures")
    parser.parse_args()
    generate()
