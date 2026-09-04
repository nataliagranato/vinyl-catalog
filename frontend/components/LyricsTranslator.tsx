"use client";

import { useState } from "react";
import { Globe, X } from "lucide-react";

interface TranslationState {
  original: string;
  translated: string | null;
  lang: string;
  loading: boolean;
  error: string | null;
}

export function useTranslation(original: string) {
  const [state, setState] = useState<TranslationState>({
    original,
    translated: null,
    lang: "en",
    loading: false,
    error: null,
  });

  const translate = async (targetLang: string) => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const res = await fetch("/api/translate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text: original, targetLang }),
      });
      if (!res.ok) throw new Error("Translation failed");
      const data = await res.json();
      setState((prev) => ({ ...prev, translated: data.translated, lang: targetLang, loading: false }));
    } catch (e) {
      setState((prev) => ({ ...prev, error: "Translation failed", loading: false }));
    }
  };

  const reset = () => {
    setState((prev) => ({ ...prev, translated: null, error: null }));
  };

  const setLang = (lang: string) => {
    setState((prev) => ({ ...prev, lang }));
  };

  return { state, translate, reset, setLang };
}

export function LyricsTranslateButton({
  state,
  onTranslate,
  onReset,
  onLangChange,
}: {
  state: TranslationState;
  onTranslate: (lang: string) => void;
  onReset: () => void;
  onLangChange: (lang: string) => void;
}) {
  if (state.translated) {
    return (
      <button
        onClick={onReset}
        className="p-1.5 text-muted hover:text-foreground transition-colors"
        title="Close translation"
      >
        <X size={13} />
      </button>
    );
  }

  return (
    <button
      onClick={() => onTranslate(state.lang)}
      disabled={state.loading}
      className="p-1.5 text-muted hover:text-accent transition-colors disabled:opacity-50"
      title="Translate lyrics"
    >
      <Globe size={13} />
    </button>
  );
}

export function LyricsSideBySide({ original, state }: { original: string; state: TranslationState }) {
  return (
    <div className="space-y-3">
      <div>
        <p className="text-xs text-muted mb-1">Original</p>
        <p className="text-sm text-foreground whitespace-pre-wrap">{original}</p>
      </div>
      {state.translated && (
        <div>
          <p className="text-xs text-muted mb-1">Translation ({state.lang})</p>
          <p className="text-sm text-foreground whitespace-pre-wrap">{state.translated}</p>
        </div>
      )}
      {state.error && <p className="text-xs text-red-400">{state.error}</p>}
    </div>
  );
}
