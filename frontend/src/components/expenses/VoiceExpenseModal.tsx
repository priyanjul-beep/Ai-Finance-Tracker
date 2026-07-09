/**
 * VoiceExpenseModal
 *
 * Full-screen modal that guides the user through recording a voice note and
 * converts it to a structured expense via the /expenses/voice-parse API.
 *
 * UI state machine:
 *   idle → requesting → recording ↔ paused → uploading → success
 *                                         ↘ error
 *
 * Props:
 *   open        – whether the modal is visible
 *   onClose     – called when user dismisses without a result
 *   onAutoFill  – called with the parsed expense when successful
 */
"use client";

import { useEffect, useRef, useCallback, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import {
  Mic,
  MicOff,
  Pause,
  Play,
  Square,
  X,
  CheckCircle2,
  AlertCircle,
  Loader2,
  Sparkles,
  Zap,
  RotateCcw,
} from "lucide-react";
import { api } from "@/services/api";
import { useVoiceRecorder } from "@/hooks/useVoiceRecorder";
import type { AIVoiceParseResponse } from "@/types";

// ─── Props ────────────────────────────────────────────────────────────────────

export interface VoiceExpenseModalProps {
  open: boolean;
  onClose: () => void;
  onAutoFill: (result: AIVoiceParseResponse) => void;
}

// ─── Local state union ────────────────────────────────────────────────────────

type ModalPhase =
  | "idle"
  | "requesting"
  | "recording"
  | "paused"
  | "uploading"
  | "success"
  | "error";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
    .toString()
    .padStart(2, "0");
  const s = (seconds % 60).toString().padStart(2, "0");
  return `${m}:${s}`;
}

function confidenceLabel(c: number): { label: string; color: string } {
  if (c >= 0.85) return { label: "High confidence", color: "text-emerald-500" };
  if (c >= 0.6)  return { label: "Medium confidence", color: "text-amber-500" };
  return           { label: "Low confidence", color: "text-rose-500" };
}

// ─── Waveform canvas ──────────────────────────────────────────────────────────

interface WaveformProps {
  data: number[];
  active: boolean;
}

function Waveform({ data, active }: WaveformProps) {
  return (
    <div className="flex items-end justify-center gap-[2px] h-14">
      {data.map((v, i) => {
        const heightPct = active ? Math.max(4, (v / 255) * 100) : 8;
        return (
          <motion.div
            key={i}
            animate={{ height: `${heightPct}%` }}
            transition={{ duration: 0.05, ease: "easeOut" }}
            className={`w-1 rounded-full ${
              active ? "bg-primary" : "bg-border"
            }`}
            style={{ height: `${heightPct}%` }}
          />
        );
      })}
    </div>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────

export function VoiceExpenseModal({
  open,
  onClose,
  onAutoFill,
}: VoiceExpenseModalProps) {
  const recorder = useVoiceRecorder();
  const [modalPhase, setModalPhase] = useState<ModalPhase>("idle");
  const [result, setResult] = useState<AIVoiceParseResponse | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploadingStep, setUploadingStep] = useState(0);
  const blobRef = useRef<Blob | null>(null);

  // Map recorder phase → modal phase
  useEffect(() => {
    if (recorder.phase === "requesting") setModalPhase("requesting");
    if (recorder.phase === "recording")  setModalPhase("recording");
    if (recorder.phase === "paused")     setModalPhase("paused");
    if (recorder.phase === "idle" && modalPhase === "requesting") setModalPhase("idle");
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [recorder.phase]);

  // Show error from recorder (e.g. permission denied)
  useEffect(() => {
    if (recorder.error) {
      setUploadError(recorder.error);
      setModalPhase("error");
    }
  }, [recorder.error]);

  // Reset when modal closes
  useEffect(() => {
    if (!open) {
      recorder.cancel();
      setModalPhase("idle");
      setResult(null);
      setUploadError(null);
      blobRef.current = null;
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  // ── Step animation for uploading phase ─────────────────────────────────────
  useEffect(() => {
    if (modalPhase !== "uploading") return;
    setUploadingStep(0);
    const t1 = setTimeout(() => setUploadingStep(1), 900);
    const t2 = setTimeout(() => setUploadingStep(2), 1_800);
    return () => { clearTimeout(t1); clearTimeout(t2); };
  }, [modalPhase]);

  // ── Actions ─────────────────────────────────────────────────────────────────

  const handleStart = useCallback(async () => {
    setUploadError(null);
    setResult(null);
    await recorder.start();
  }, [recorder]);

  const handleStop = useCallback(async () => {
    const blob = await recorder.stop();
    if (!blob || blob.size < 1000) {
      setUploadError("Recording was too short. Please say your expense clearly and try again.");
      setModalPhase("error");
      return;
    }
    blobRef.current = blob;
    setModalPhase("uploading");

    try {
      const form = new FormData();
      form.append("audio", blob, "voice.webm");

      const resp = await api.post<AIVoiceParseResponse>(
        "/expenses/voice-parse",
        form,
        { headers: { "Content-Type": "multipart/form-data" }, timeout: 45_000 }
      );
      setResult(resp.data);
      setModalPhase("success");
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error
        ?? "Failed to process audio. Please try again.";
      setUploadError(msg);
      setModalPhase("error");
    }
  }, [recorder]);

  const handleConfirm = useCallback(() => {
    if (result) {
      onAutoFill(result);
      onClose();
    }
  }, [result, onAutoFill, onClose]);

  const handleRetry = useCallback(() => {
    recorder.reset();
    setResult(null);
    setUploadError(null);
    blobRef.current = null;
    setModalPhase("idle");
  }, [recorder]);

  // ─── Render helpers ─────────────────────────────────────────────────────────

  const renderBody = () => {
    switch (modalPhase) {
      // ── Idle ──────────────────────────────────────────────────────────────
      case "idle":
        return (
          <motion.div
            key="idle"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            className="flex flex-col items-center gap-6 py-4"
          >
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-primary/10 ring-4 ring-primary/20">
              <Mic className="h-9 w-9 text-primary" />
            </div>

            <div className="text-center space-y-1">
              <p className="font-semibold text-base">Speak your expense</p>
              <p className="text-sm text-muted-foreground">
                Describe what you spent — amount, merchant, and category.
              </p>
            </div>

            {/* Example hints */}
            <div className="w-full rounded-xl border border-border bg-muted/40 p-4 space-y-2">
              <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
                Try saying…
              </p>
              {[
                "Paid ₹250 for coffee at Blue Tokai using UPI",
                "Spent $45 on groceries at Whole Foods yesterday",
                "Netflix subscription ₹649 card payment today",
                "Refund of ₹1200 from Flipkart for cancelled order",
              ].map((ex) => (
                <p key={ex} className="flex gap-2 text-sm text-muted-foreground">
                  <Zap className="h-3.5 w-3.5 mt-0.5 shrink-0 text-primary/60" />
                  <span>"{ex}"</span>
                </p>
              ))}
            </div>

            <button
              onClick={handleStart}
              className="flex items-center gap-2 rounded-xl bg-primary px-8 py-3 text-sm font-semibold text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              <Mic className="h-4 w-4" />
              Start Recording
            </button>
          </motion.div>
        );

      // ── Requesting permission ──────────────────────────────────────────────
      case "requesting":
        return (
          <motion.div
            key="requesting"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="flex flex-col items-center gap-4 py-8"
          >
            <Loader2 className="h-10 w-10 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">Requesting microphone permission…</p>
          </motion.div>
        );

      // ── Recording ─────────────────────────────────────────────────────────
      case "recording":
      case "paused":
        return (
          <motion.div
            key="recording"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            className="flex flex-col items-center gap-5 py-2"
          >
            {/* Status pill */}
            <div className="flex items-center gap-2 rounded-full border border-border bg-muted/40 px-3 py-1">
              <span
                className={`h-2 w-2 rounded-full ${
                  modalPhase === "recording"
                    ? "bg-rose-500 animate-pulse"
                    : "bg-amber-400"
                }`}
              />
              <span className="text-xs font-medium text-muted-foreground">
                {modalPhase === "recording" ? "Recording" : "Paused"}
              </span>
            </div>

            {/* Duration */}
            <span className="text-4xl font-mono font-semibold tracking-tight">
              {formatDuration(recorder.duration)}
            </span>

            {/* Waveform */}
            <div className="w-full px-2">
              <Waveform data={recorder.waveform} active={modalPhase === "recording"} />
            </div>

            {/* Controls */}
            <div className="flex items-center gap-3 pt-1">
              {/* Cancel */}
              <button
                onClick={handleRetry}
                className="flex h-10 w-10 items-center justify-center rounded-full border border-border hover:bg-muted text-muted-foreground transition-colors"
                title="Cancel"
              >
                <X className="h-4 w-4" />
              </button>

              {/* Pause / Resume */}
              {modalPhase === "recording" ? (
                <button
                  onClick={recorder.pause}
                  className="flex h-12 w-12 items-center justify-center rounded-full border border-border hover:bg-muted transition-colors"
                  title="Pause"
                >
                  <Pause className="h-5 w-5" />
                </button>
              ) : (
                <button
                  onClick={recorder.resume}
                  className="flex h-12 w-12 items-center justify-center rounded-full border border-border hover:bg-muted transition-colors"
                  title="Resume"
                >
                  <Play className="h-5 w-5" />
                </button>
              )}

              {/* Stop & process */}
              <button
                onClick={handleStop}
                className="flex items-center gap-2 rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-primary/90 transition-colors"
                title="Stop and process"
              >
                <Square className="h-4 w-4 fill-current" />
                Done
              </button>
            </div>

            <p className="text-xs text-muted-foreground text-center">
              Speak clearly. The AI will parse your expense automatically.
            </p>
          </motion.div>
        );

      // ── Uploading / processing ─────────────────────────────────────────────
      case "uploading": {
        const steps = [
          "Uploading audio…",
          "Transcribing speech…",
          "Extracting expense details…",
        ];
        return (
          <motion.div
            key="uploading"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="flex flex-col items-center gap-6 py-8"
          >
            <div className="relative flex h-20 w-20 items-center justify-center">
              <motion.div
                className="absolute inset-0 rounded-full border-2 border-primary/30"
                animate={{ scale: [1, 1.15, 1] }}
                transition={{ repeat: Infinity, duration: 1.4, ease: "easeInOut" }}
              />
              <Sparkles className="h-9 w-9 text-primary" />
            </div>

            <div className="space-y-2 text-center">
              {steps.map((step, idx) => (
                <motion.p
                  key={step}
                  initial={{ opacity: 0, x: -8 }}
                  animate={{
                    opacity: uploadingStep >= idx ? 1 : 0.3,
                    x: 0,
                  }}
                  transition={{ delay: idx * 0.3 }}
                  className={`text-sm flex items-center gap-2 justify-center ${
                    uploadingStep === idx
                      ? "text-foreground font-medium"
                      : uploadingStep > idx
                      ? "text-muted-foreground line-through"
                      : "text-muted-foreground"
                  }`}
                >
                  {uploadingStep === idx && (
                    <Loader2 className="h-3.5 w-3.5 animate-spin text-primary shrink-0" />
                  )}
                  {step}
                </motion.p>
              ))}
            </div>
          </motion.div>
        );
      }

      // ── Success ───────────────────────────────────────────────────────────
      case "success":
        if (!result) return null;
        const conf = confidenceLabel(result.confidence);
        return (
          <motion.div
            key="success"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            className="space-y-4"
          >
            {/* Header */}
            <div className="flex items-center gap-3">
              <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0" />
              <div>
                <p className="font-semibold text-sm">Expense detected</p>
                <p className={`text-xs ${conf.color}`}>{conf.label}</p>
              </div>
              {result.cached && (
                <span className="ml-auto rounded-full bg-sky-100 px-2 py-0.5 text-[10px] font-medium text-sky-700 dark:bg-sky-900/40 dark:text-sky-300">
                  cached
                </span>
              )}
            </div>

            {/* Transcript */}
            <div className="rounded-lg border border-border bg-muted/30 p-3">
              <p className="text-xs text-muted-foreground mb-1 font-medium uppercase tracking-wide">
                You said
              </p>
              <p className="text-sm italic text-foreground/80">
                "{result.transcript || "—"}"
              </p>
            </div>

            {/* Parsed fields grid */}
            <div className="grid grid-cols-2 gap-2">
              {[
                { label: "Amount",   value: result.amount  ? `${result.amount}` : "—" },
                { label: "Merchant", value: result.merchant || "—" },
                { label: "Category", value: result.category || "—" },
                { label: "Date",     value: result.date    || "—" },
                { label: "Payment",  value: result.payment_method || "—" },
                { label: "Type",     value: result.expense_type  || "spend" },
              ].map(({ label, value }) => (
                <div
                  key={label}
                  className="rounded-lg border border-border bg-card p-2.5"
                >
                  <p className="text-[10px] text-muted-foreground uppercase tracking-wide mb-0.5">
                    {label}
                  </p>
                  <p className="text-sm font-medium capitalize truncate">{value}</p>
                </div>
              ))}
            </div>

            {result.notes && (
              <div className="rounded-lg border border-border bg-muted/30 p-3">
                <p className="text-xs text-muted-foreground mb-0.5 uppercase tracking-wide font-medium">Notes</p>
                <p className="text-sm">{result.notes}</p>
              </div>
            )}

            {/* Actions */}
            <div className="flex gap-2 pt-1">
              <button
                onClick={handleRetry}
                className="flex items-center gap-1.5 rounded-lg border border-border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
              >
                <RotateCcw className="h-3.5 w-3.5" />
                Re-record
              </button>
              <button
                onClick={handleConfirm}
                className="flex flex-1 items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                <CheckCircle2 className="h-4 w-4" />
                Fill Form
              </button>
            </div>
          </motion.div>
        );

      // ── Error ─────────────────────────────────────────────────────────────
      case "error":
        return (
          <motion.div
            key="error"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            className="flex flex-col items-center gap-5 py-4"
          >
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircle className="h-8 w-8 text-destructive" />
            </div>
            <div className="text-center space-y-1">
              <p className="font-semibold">Something went wrong</p>
              <p className="text-sm text-muted-foreground max-w-xs">
                {uploadError ?? "An unexpected error occurred."}
              </p>
            </div>
            <div className="flex gap-3">
              <button
                onClick={onClose}
                className="rounded-lg border border-border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
              >
                Close
              </button>
              <button
                onClick={handleRetry}
                className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                <RotateCcw className="h-4 w-4" />
                Try Again
              </button>
            </div>
          </motion.div>
        );
    }
  };

  // ─── Modal shell ───────────────────────────────────────────────────────────

  return (
    <AnimatePresence>
      {open && (
        <>
          {/* Backdrop */}
          <motion.div
            key="backdrop"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => {
              if (modalPhase === "idle" || modalPhase === "error" || modalPhase === "success") {
                onClose();
              }
            }}
            className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
          />

          {/* Panel */}
          <motion.div
            key="panel"
            initial={{ opacity: 0, scale: 0.95, y: 16 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 16 }}
            transition={{ type: "spring", stiffness: 350, damping: 30 }}
            className="fixed inset-x-4 bottom-4 z-50 mx-auto max-w-md rounded-2xl border border-border bg-card shadow-2xl overflow-hidden sm:inset-x-auto sm:left-1/2 sm:-translate-x-1/2 sm:w-full"
          >
            {/* Drag handle */}
            <div className="flex justify-center pt-3 pb-1">
              <div className="h-1 w-8 rounded-full bg-border" />
            </div>

            {/* Header */}
            <div className="flex items-center justify-between px-5 pb-3">
              <div className="flex items-center gap-2">
                <Mic className="h-4 w-4 text-primary" />
                <span className="text-sm font-semibold">Voice Input</span>
              </div>
              <button
                onClick={onClose}
                disabled={modalPhase === "recording" || modalPhase === "uploading"}
                className="flex h-7 w-7 items-center justify-center rounded-lg hover:bg-muted text-muted-foreground transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            {/* Divider */}
            <div className="border-t border-border" />

            {/* Body — animated phase transitions */}
            <div className="px-5 py-5 min-h-[200px]">
              <AnimatePresence mode="wait">
                {renderBody()}
              </AnimatePresence>
            </div>
          </motion.div>
        </>
      )}
    </AnimatePresence>
  );
}
