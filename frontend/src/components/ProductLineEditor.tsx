import { useEffect, useState } from "react";
import type { Slot } from "../types/Slot";

export type ProductLineMode = "auto" | "operator" | "disabled";

interface ProductLineEditorProps {
  slot: Slot;
  pending: boolean;
  onGenerate: (mode: ProductLineMode, customProductLine: string) => Promise<void>;
}

function deriveInitialMode(slot: Slot): ProductLineMode {
  if (slot.product_line_mode === "operator" || slot.product_line_mode === "disabled") {
    return slot.product_line_mode;
  }
  return "auto";
}

export function ProductLineEditor({ slot, pending, onGenerate }: ProductLineEditorProps) {
  const [mode, setMode] = useState<ProductLineMode>(deriveInitialMode(slot));
  const [customProductLine, setCustomProductLine] = useState(slot.final_product_line ?? slot.suggested_product_line ?? "");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setMode(deriveInitialMode(slot));
    setCustomProductLine(slot.final_product_line ?? slot.suggested_product_line ?? "");
    setError(null);
  }, [slot.id, slot.final_product_line, slot.product_line_mode, slot.suggested_product_line]);

  const isReadonly = slot.status === "generating" || slot.status === "generated" || slot.status === "failed";

  async function handleSubmit() {
    if (mode === "operator" && customProductLine.trim() === "") {
      setError("Custom product line is required for operator mode.");
      return;
    }

    setError(null);
    await onGenerate(mode, customProductLine.trim());
  }

  return (
    <section className="panel">
      <p className="eyebrow">Phase 3</p>
      <h2>Product Line Review</h2>
      <p>
        Review the suggested line for the selected slot, replace it with an operator-edited version, or disable dialogue
        before starting CAFAI generation.
      </p>

      {slot.suggested_product_line ? (
        <p className="success-summary">
          <strong>Suggested line:</strong> {slot.suggested_product_line}
        </p>
      ) : null}

      <fieldset className="mode-toggle">
        <legend>Dialogue mode</legend>
        <label>
          <input
            type="radio"
            name={`product-line-mode-${slot.id}`}
            value="auto"
            checked={mode === "auto"}
            onChange={() => setMode("auto")}
            disabled={isReadonly || pending}
          />
          Auto
        </label>
        <label>
          <input
            type="radio"
            name={`product-line-mode-${slot.id}`}
            value="operator"
            checked={mode === "operator"}
            onChange={() => setMode("operator")}
            disabled={isReadonly || pending}
          />
          Operator edit
        </label>
        <label>
          <input
            type="radio"
            name={`product-line-mode-${slot.id}`}
            value="disabled"
            checked={mode === "disabled"}
            onChange={() => setMode("disabled")}
            disabled={isReadonly || pending}
          />
          Disabled
        </label>
      </fieldset>

      <label className="field">
        <span>Operator line</span>
        <textarea
          value={customProductLine}
          onChange={(event) => setCustomProductLine(event.target.value)}
          rows={4}
          disabled={mode !== "operator" || isReadonly || pending}
        />
      </label>

      {slot.status === "generating" ? <p className="form-message form-message--success">CAFAI generation is running.</p> : null}
      {slot.status === "generated" ? (
        <div className="success-summary">
          <p><strong>Generation complete.</strong></p>
          {slot.generated_clip_path ? <p>Clip: {slot.generated_clip_path}</p> : null}
          {slot.generated_audio_path ? <p>Audio: {slot.generated_audio_path}</p> : null}
        </div>
      ) : null}
      {slot.status === "failed" && slot.generation_error ? <p className="form-message form-message--error">{slot.generation_error}</p> : null}
      {error ? <p className="form-message form-message--error">{error}</p> : null}

      <div className="form-actions">
        <button type="button" onClick={() => void handleSubmit()} disabled={pending || isReadonly}>
          {pending ? "Starting..." : "Start generation"}
        </button>
      </div>
    </section>
  );
}
