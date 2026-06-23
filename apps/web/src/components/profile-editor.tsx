"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CareerProfile, Education, Experience } from "@roliq/api-client";
import { Plus, Save, Trash2 } from "lucide-react";
import { useRoliqClient } from "@/lib/api";

const emptyExperience = (): Experience => ({
  company: "",
  title: "",
  location: "",
  startDate: "",
  isCurrent: false,
  description: "",
});
const emptyEducation = (): Education => ({
  institution: "",
  degree: "",
  fieldOfStudy: "",
});

export function ProfileEditor({
  onSaved,
  compact = false,
}: {
  onSaved?: () => void;
  compact?: boolean;
}) {
  const api = useRoliqClient();
  const query = useQuery({ queryKey: ["profile"], queryFn: api.getProfile });
  if (query.isLoading || !query.data)
    return <p className="muted">Loading your career profile…</p>;
  return (
    <ProfileEditorForm
      key={query.data.updatedAt ?? "new"}
      initial={query.data}
      onSaved={onSaved}
      compact={compact}
    />
  );
}
function ProfileEditorForm({
  initial,
  onSaved,
  compact,
}: {
  initial: CareerProfile;
  onSaved?: () => void;
  compact: boolean;
}) {
  const api = useRoliqClient();
  const queryClient = useQueryClient();
  const [profile, setProfile] = useState<CareerProfile>(initial);
  const [skillInput, setSkillInput] = useState("");
  const mutation = useMutation({
    mutationFn: api.saveProfile,
    onSuccess: async (saved) => {
      setProfile(saved);
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["dashboard"] }),
        queryClient.invalidateQueries({ queryKey: ["profile"] }),
      ]);
      onSaved?.();
    },
  });
  function patch<K extends keyof CareerProfile>(
    key: K,
    value: CareerProfile[K],
  ) {
    setProfile((current) => ({ ...current, [key]: value }));
  }
  function updateExperience(index: number, value: Experience) {
    patch(
      "experiences",
      profile.experiences.map((item, i) => (i === index ? value : item)),
    );
  }
  function updateEducation(index: number, value: Education) {
    patch(
      "education",
      profile.education.map((item, i) => (i === index ? value : item)),
    );
  }
  function addSkill() {
    const value = skillInput.trim();
    if (
      value &&
      !profile.skills.some((item) => item.toLowerCase() === value.toLowerCase())
    )
      patch("skills", [...profile.skills, value]);
    setSkillInput("");
  }
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        mutation.mutate(profile);
      }}
      style={{ display: "grid", gap: 22 }}
    >
      <section className="card" style={{ padding: 26 }}>
        <h2 style={{ fontFamily: "Georgia", fontSize: 25, marginTop: 0 }}>
          Professional overview
        </h2>
        <div
          className="responsive-grid"
          style={{ display: "grid", gridTemplateColumns: "2fr 1fr", gap: 16 }}
        >
          <label className="label">
            Professional headline
            <input
              className="input"
              value={profile.headline}
              maxLength={160}
              placeholder="Senior product engineer"
              onChange={(e) => patch("headline", e.target.value)}
            />
          </label>
          <label className="label">
            Years of experience
            <input
              className="input"
              type="number"
              min="0"
              max="80"
              step="0.5"
              value={profile.yearsExperience ?? ""}
              onChange={(e) =>
                patch(
                  "yearsExperience",
                  e.target.value === "" ? undefined : Number(e.target.value),
                )
              }
            />
          </label>
        </div>
        <label className="label" style={{ marginTop: 16 }}>
          Career summary
          <textarea
            className="textarea"
            value={profile.summary}
            maxLength={4000}
            placeholder="Describe the work you do best and the outcomes you create."
            onChange={(e) => patch("summary", e.target.value)}
          />
        </label>
        <div
          className="responsive-grid"
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(3,1fr)",
            gap: 16,
            marginTop: 16,
          }}
        >
          <label className="label">
            City
            <input
              className="input"
              value={profile.city}
              maxLength={120}
              onChange={(e) => patch("city", e.target.value)}
            />
          </label>
          <label className="label">
            Country code
            <input
              className="input"
              value={profile.countryCode}
              maxLength={2}
              placeholder="IN"
              onChange={(e) =>
                patch("countryCode", e.target.value.toUpperCase())
              }
            />
          </label>
          <label className="label">
            Time zone
            <input
              className="input"
              value={profile.timeZone}
              placeholder="Asia/Kolkata"
              onChange={(e) => patch("timeZone", e.target.value)}
            />
          </label>
        </div>
      </section>
      <section className="card" style={{ padding: 26 }}>
        <h2 style={{ fontFamily: "Georgia", fontSize: 25, marginTop: 0 }}>
          Skills
        </h2>
        <div style={{ display: "flex", gap: 10 }}>
          <input
            className="input"
            value={skillInput}
            maxLength={80}
            placeholder="Add a skill"
            onChange={(e) => setSkillInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                addSkill();
              }
            }}
          />
          <button type="button" className="button secondary" onClick={addSkill}>
            Add
          </button>
        </div>
        <div
          style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 16 }}
        >
          {profile.skills.map((skill) => (
            <button
              type="button"
              key={skill}
              className="pill"
              title="Remove skill"
              onClick={() =>
                patch(
                  "skills",
                  profile.skills.filter((item) => item !== skill),
                )
              }
            >
              {skill} ×
            </button>
          ))}
        </div>
      </section>
      <section className="card" style={{ padding: 26 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <h2 style={{ fontFamily: "Georgia", fontSize: 25, margin: 0 }}>
            Experience
          </h2>
          <button
            type="button"
            className="button secondary"
            onClick={() =>
              patch("experiences", [...profile.experiences, emptyExperience()])
            }
          >
            <Plus size={16} /> Add role
          </button>
        </div>
        <div style={{ display: "grid", gap: 16, marginTop: 20 }}>
          {profile.experiences.length === 0 && (
            <p className="muted">
              Add your real work history. Dates and current-role status are
              validated when saved.
            </p>
          )}
          {profile.experiences.map((item, index) => (
            <div
              key={item.id ?? index}
              style={{
                border: "1px solid var(--line)",
                borderRadius: 15,
                padding: 18,
              }}
            >
              <div
                className="responsive-grid"
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: 14,
                }}
              >
                <label className="label">
                  Company
                  <input
                    required
                    className="input"
                    value={item.company}
                    onChange={(e) =>
                      updateExperience(index, {
                        ...item,
                        company: e.target.value,
                      })
                    }
                  />
                </label>
                <label className="label">
                  Title
                  <input
                    required
                    className="input"
                    value={item.title}
                    onChange={(e) =>
                      updateExperience(index, {
                        ...item,
                        title: e.target.value,
                      })
                    }
                  />
                </label>
                <label className="label">
                  Start date
                  <input
                    required
                    type="date"
                    className="input"
                    value={item.startDate}
                    onChange={(e) =>
                      updateExperience(index, {
                        ...item,
                        startDate: e.target.value,
                      })
                    }
                  />
                </label>
                <label className="label">
                  End date
                  <input
                    type="date"
                    className="input"
                    disabled={item.isCurrent}
                    value={item.endDate ?? ""}
                    onChange={(e) =>
                      updateExperience(index, {
                        ...item,
                        endDate: e.target.value || undefined,
                      })
                    }
                  />
                </label>
              </div>
              <label
                style={{
                  display: "flex",
                  gap: 8,
                  alignItems: "center",
                  margin: "14px 0",
                }}
              >
                <input
                  type="checkbox"
                  checked={item.isCurrent}
                  onChange={(e) =>
                    updateExperience(index, {
                      ...item,
                      isCurrent: e.target.checked,
                      endDate: e.target.checked ? undefined : item.endDate,
                    })
                  }
                />{" "}
                I currently work here
              </label>
              {!compact && (
                <label className="label">
                  Description
                  <textarea
                    className="textarea"
                    value={item.description ?? ""}
                    onChange={(e) =>
                      updateExperience(index, {
                        ...item,
                        description: e.target.value,
                      })
                    }
                  />
                </label>
              )}
              <button
                type="button"
                className="button ghost"
                style={{ color: "var(--danger)", padding: 0, marginTop: 12 }}
                onClick={() =>
                  patch(
                    "experiences",
                    profile.experiences.filter((_, i) => i !== index),
                  )
                }
              >
                <Trash2 size={15} /> Remove
              </button>
            </div>
          ))}
        </div>
      </section>
      <section className="card" style={{ padding: 26 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <h2 style={{ fontFamily: "Georgia", fontSize: 25, margin: 0 }}>
            Education
          </h2>
          <button
            type="button"
            className="button secondary"
            onClick={() =>
              patch("education", [...profile.education, emptyEducation()])
            }
          >
            <Plus size={16} /> Add education
          </button>
        </div>
        <div style={{ display: "grid", gap: 16, marginTop: 20 }}>
          {profile.education.map((item, index) => (
            <div
              key={item.id ?? index}
              style={{
                border: "1px solid var(--line)",
                borderRadius: 15,
                padding: 18,
              }}
            >
              <div
                className="responsive-grid"
                style={{
                  display: "grid",
                  gridTemplateColumns: "1fr 1fr",
                  gap: 14,
                }}
              >
                <label className="label">
                  Institution
                  <input
                    required
                    className="input"
                    value={item.institution}
                    onChange={(e) =>
                      updateEducation(index, {
                        ...item,
                        institution: e.target.value,
                      })
                    }
                  />
                </label>
                <label className="label">
                  Degree
                  <input
                    className="input"
                    value={item.degree ?? ""}
                    onChange={(e) =>
                      updateEducation(index, {
                        ...item,
                        degree: e.target.value,
                      })
                    }
                  />
                </label>
                <label className="label">
                  Field of study
                  <input
                    className="input"
                    value={item.fieldOfStudy ?? ""}
                    onChange={(e) =>
                      updateEducation(index, {
                        ...item,
                        fieldOfStudy: e.target.value,
                      })
                    }
                  />
                </label>
                <label className="label">
                  Start date
                  <input
                    type="date"
                    className="input"
                    value={item.startDate ?? ""}
                    onChange={(e) =>
                      updateEducation(index, {
                        ...item,
                        startDate: e.target.value || undefined,
                      })
                    }
                  />
                </label>
                <label className="label">
                  End date
                  <input
                    type="date"
                    className="input"
                    value={item.endDate ?? ""}
                    onChange={(e) =>
                      updateEducation(index, {
                        ...item,
                        endDate: e.target.value || undefined,
                      })
                    }
                  />
                </label>
              </div>
              <button
                type="button"
                className="button ghost"
                style={{ color: "var(--danger)", padding: 0, marginTop: 12 }}
                onClick={() =>
                  patch(
                    "education",
                    profile.education.filter((_, i) => i !== index),
                  )
                }
              >
                <Trash2 size={15} /> Remove
              </button>
            </div>
          ))}
        </div>
      </section>
      {mutation.error && (
        <p role="alert" className="field-error">
          {mutation.error.message}
        </p>
      )}
      <div style={{ display: "flex", justifyContent: "flex-end" }}>
        <button className="button" disabled={mutation.isPending}>
          <Save size={17} />
          {mutation.isPending ? "Saving…" : "Save career profile"}
        </button>
      </div>
    </form>
  );
}
