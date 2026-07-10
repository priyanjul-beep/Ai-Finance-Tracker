/**
 * ReceiptScanModal
 *
 * Full AI receipt / payment-screenshot scanner modal for the Add Expense page.
 *
 * Flow:
 *   idle  →  selected  →  processing  →  success
 *                 ↕                   ↘  error
 *
 * Props:
 *   open        – controls visibility
 *   onClose     – called on dismiss / after filling
 *   onAutoFill  – receives the parsed AIReceiptScanResponse to populate the form
 */
"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import {
  X,
  Upload,
  ImagePlus,
  Loader2,
  CheckCircle2,
  AlertCircle,
  RotateCcw,
  Sparkles,
  Receipt,
  Zap,
  ShoppingBag,
  Utensils,
  Car,
  HeartPulse,
  Bolt,
  Smartphone,
  Fuel,
  CreditCard,
  BookOpen,
  Tag,
  HelpCircle,
  BadgeIndianRupee,
} from "lucide-react";
import { api } from "@/services/api";
import type { AIReceiptScanResponse } from "@/types";

// ─── Props ────────────────────────────────────────────────────────────────────

export interface ReceiptScanModalProps {
  open: boolean;
  onClose: () => void;
  onAutoFill: (result: AIReceiptScanResponse) => void;
}

// ─── Phase ────────────────────────────────────────────────────────────────────

type Phase = "idle" | "selected" | "processing" | "success" | "error";

// ─── Helpers ──────────────────────────────────────────────────────────────────

const ACCEPTED_TYPES = ["image/jpeg", "image/jpg", "image/png", "image/webp"];
const MAX_BYTES = 10 * 1024 * 1024; // 10 MB

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function confidenceBadge(c: number) {
  if (c >= 0.85) return { label: "High confidence", cls: "bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300" };
  if (c >= 0.6)  return { label: "Medium confidence", cls: "bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300" };
  return           { label: "Low confidence",    cls: "bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300" };
}

const CATEGORY_ICON: Record<string, React.ReactNode> = {
  food:          <Utensils className="h-3.5 w-3.5" />,
  travel:        <Car className="h-3.5 w-3.5" />,
  shopping:      <ShoppingBag className="h-3.5 w-3.5" />,
  entertainment: <Sparkles className="h-3.5 w-3.5" />,
  health:        <HeartPulse className="h-3.5 w-3.5" />,
  bills:         <Bolt className="h-3.5 w-3.5" />,
  recharge:      <Smartphone className="h-3.5 w-3.5" />,
  fuel:          <Fuel className="h-3.5 w-3.5" />,
  subscription:  <CreditCard className="h-3.5 w-3.5" />,
  education:     <BookOpen className="h-3.5 w-3.5" />,
  utilities:     <Bolt className="h-3.5 w-3.5" />,
  rent:          <Tag className="h-3.5 w-3.5" />,
  others:        <Tag className="h-3.5 w-3.5" />,
  unknown:       <HelpCircle className="h-3.5 w-3.5" />,
};

const SUPPORTED_EXAMPLES = [
  "Google Pay / PhonePe / Paytm screenshot",
  "Amazon / Flipkart order receipt",
  "Restaurant bill or food receipt",
  "Fuel / petrol station receipt",
  "Grocery or supermarket bill",
  "Electricity / utility bill",
  "Medical or pharmacy invoice",
];

// ─── Processing steps ─────────────────────────────────────────────────────────

const STEPS = [
  "Uploading image…",
  "Reading receipt data…",
  "Analysing with Gemini AI…",
  "Extracting expense details…",
];

// ─── Component ────────────────────────────────────────────────────────────────

export function ReceiptScanModal({ open, onClose, onAutoFill }: ReceiptScanModalProps) {
  const [phase, setPhase]         = useState<Phase>("idle");
  const [file, setFile]           = useState<File | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [step, setStep]           = useState(0);
  const [result, setResult]       = useState<AIReceiptScanResponse | null>(null);
  const [error, setError]         = useState<string | null>(null);

  const inputRef = useRef<HTMLInputElement>(null);
  const prevPreviewRef = useRef<string | null>(null);

  // Revoke object URLs to avoid memory leaks
  useEffect(() => {
    if (prevPreviewRef.current) URL.revokeObjectURL(prevPreviewRef.current);
    prevPreviewRef.current = previewUrl;
  }, [previewUrl]);

  // Reset on close
  useEffect(() => {
    if (!open) {
      setPhase("idle");
      setFile(null);
      setPreviewUrl(null);
      setResult(null);
      setError(null);
      setStep(0);
    }
  }, [open]);

  // ── File handling ──────────────────────────────────────────────────────────

  const acceptFile = useCallback((f: File) => {
    if (!ACCEPTED_TYPES.includes(f.type)) {
      setError("Only JPG, PNG, and WEBP images are supported.");
      setPhase("error");
      return;
    }
    if (f.size > MAX_BYTES) {
      setError("Image is too large. Maximum size is 10 MB.");
      setPhase("error");
      return;
    }
    setFile(f);
    setPreviewUrl(URL.createObjectURL(f));
    setPhase("selected");
    setError(null);
  }, []);

  const handleInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (f) acceptFile(f);
    // reset so same file can be re-selected
    e.target.value = "";
  }, [acceptFile]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const f = e.dataTransfer.files?.[0];
    if (f) acceptFile(f);
  }, [acceptFile]);

  // ── Processing ─────────────────────────────────────────────────────────────

  const handleProcess = useCallback(async () => {
    if (!file) return;
    setPhase("processing");
    setStep(0);

    // Animate steps while request is in-flight
    const timers = STEPS.slice(1).map((_, i) =>
      setTimeout(() => setStep(i + 1), (i + 1) * 1_200)
    );

    try {
      const form = new FormData();
      form.append("image", file);
      const resp = await api.post<AIReceiptScanResponse>(
        "/expenses/scan",
        form,
        { headers: { "Content-Type": "multipart/form-data" }, timeout: 60_000 }
      );
      timers.forEach(clearTimeout);
      setResult(resp.data);
      setPhase("success");
    } catch (err: unknown) {
      timers.forEach(clearTimeout);
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error
        ?? "Failed to process image. Please try again.";
      setError(msg);
      setPhase("error");
    }
  }, [file]);

  const handleConfirm = useCallback(() => {
    if (result) {
      onAutoFill(result);
      onClose();
    }
  }, [result, onAutoFill, onClose]);

  const handleReset = useCallback(() => {
    setPhase("idle");
    setFile(null);
    setPreviewUrl(null);
    setResult(null);
    setError(null);
    setStep(0);
  }, []);

  // ── Render body ────────────────────────────────────────────────────────────

  const renderBody = () => {
    switch (phase) {

      // ── Idle — drop zone ─────────────────────────────────────────────────
      case "idle":
        return (
          <motion.div key="idle" initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -8 }}>
            {/* Drop zone */}
            <div
              onDragOver={(e) => { e.preventDefault(); setIsDragging(true); }}
              onDragLeave={() => setIsDragging(false)}
              onDrop={handleDrop}
              onClick={() => inputRef.current?.click()}
              className={`relative flex flex-col items-center justify-center gap-3 cursor-pointer rounded-xl border-2 border-dashed transition-all py-10 ${
                isDragging
                  ? "border-primary bg-primary/5 scale-[1.01]"
                  : "border-border hover:border-primary/50 hover:bg-muted/40"
              }`}
            >
              <div className={`flex h-14 w-14 items-center justify-center rounded-2xl transition-colors ${isDragging ? "bg-primary/15" : "bg-muted"}`}>
                <ImagePlus className={`h-7 w-7 ${isDragging ? "text-primary" : "text-muted-foreground"}`} />
              </div>
              <div className="text-center">
                <p className="font-semibold text-sm">
                  {isDragging ? "Drop image here" : "Drop your receipt or screenshot here"}
                </p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  or <span className="text-primary font-medium">click to upload</span>
                </p>
              </div>
              <p className="text-[11px] text-muted-foreground">
                JPG · PNG · WEBP · max 10 MB
              </p>
            </div>

            {/* Supported examples */}
            <div className="mt-4 rounded-xl border border-border bg-muted/30 p-4">
              <p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground mb-2">
                Works with
              </p>
              <div className="grid grid-cols-2 gap-y-1.5">
                {SUPPORTED_EXAMPLES.map((ex) => (
                  <p key={ex} className="flex items-start gap-1.5 text-xs text-muted-foreground">
                    <Zap className="h-3 w-3 mt-0.5 shrink-0 text-primary/60" />
                    {ex}
                  </p>
                ))}
              </div>
            </div>

            <input
              ref={inputRef}
              type="file"
              accept="image/jpeg,image/jpg,image/png,image/webp"
              className="hidden"
              onChange={handleInputChange}
            />
          </motion.div>
        );

      // ── Selected — preview ────────────────────────────────────────────────
      case "selected":
        return (
          <motion.div key="selected" initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -8 }} className="space-y-4">
            {/* Image preview */}
            <div className="relative overflow-hidden rounded-xl border border-border bg-muted/20">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={previewUrl!}
                alt="Receipt preview"
                className="w-full max-h-52 object-contain"
              />
              <button
                onClick={handleReset}
                className="absolute top-2 right-2 flex h-7 w-7 items-center justify-center rounded-full bg-black/50 text-white hover:bg-black/70 transition-colors"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </div>

            {/* File info */}
            <div className="flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2">
              <Receipt className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="text-sm truncate flex-1 text-foreground">{file?.name}</span>
              <span className="text-xs text-muted-foreground shrink-0">{formatBytes(file?.size ?? 0)}</span>
            </div>

            {/* Actions */}
            <div className="flex gap-2">
              <button
                onClick={handleReset}
                className="flex-1 rounded-xl border border-border px-4 py-2.5 text-sm font-medium hover:bg-muted transition-colors"
              >
                Choose different
              </button>
              <button
                onClick={handleProcess}
                className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                <Sparkles className="h-4 w-4" />
                Scan Receipt
              </button>
            </div>
          </motion.div>
        );

      // ── Processing ────────────────────────────────────────────────────────
      case "processing":
        return (
          <motion.div key="processing" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="flex flex-col items-center gap-6 py-6">
            {/* Thumbnail */}
            {previewUrl && (
              <div className="h-16 w-16 rounded-xl overflow-hidden border border-border shadow-sm">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={previewUrl} alt="" className="h-full w-full object-cover" />
              </div>
            )}

            {/* Pulsing icon */}
            <div className="relative flex h-16 w-16 items-center justify-center">
              <motion.div
                className="absolute inset-0 rounded-full border-2 border-primary/30"
                animate={{ scale: [1, 1.2, 1] }}
                transition={{ repeat: Infinity, duration: 1.6, ease: "easeInOut" }}
              />
              <Sparkles className="h-8 w-8 text-primary" />
            </div>

            {/* Steps */}
            <div className="w-full space-y-2">
              {STEPS.map((s, i) => (
                <motion.div
                  key={s}
                  initial={{ opacity: 0, x: -8 }}
                  animate={{ opacity: step >= i ? 1 : 0.3, x: 0 }}
                  transition={{ delay: i * 0.15 }}
                  className={`flex items-center gap-2 text-sm ${
                    step === i ? "text-foreground font-medium" :
                    step > i  ? "text-muted-foreground line-through" :
                    "text-muted-foreground"
                  }`}
                >
                  {step === i
                    ? <Loader2 className="h-3.5 w-3.5 shrink-0 animate-spin text-primary" />
                    : step > i
                    ? <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-emerald-500" />
                    : <div className="h-3.5 w-3.5 shrink-0 rounded-full border border-muted-foreground/30" />
                  }
                  {s}
                </motion.div>
              ))}
            </div>
          </motion.div>
        );

      // ── Success — verification card ────────────────────────────────────────
      case "success":
        if (!result) return null;
        const badge = confidenceBadge(result.confidence);
        const currencySymbol = result.currency === "INR" ? "₹" : result.currency === "USD" ? "$" : result.currency;
        return (
          <motion.div key="success" initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} className="space-y-4">
            {/* Header */}
            <div className="flex items-start gap-3">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/40">
                <CheckCircle2 className="h-5 w-5 text-emerald-600 dark:text-emerald-400" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="font-semibold text-sm">Receipt Successfully Processed</p>
                <p className={`text-xs mt-0.5 font-medium px-1.5 py-0.5 rounded-full inline-flex items-center gap-1 ${badge.cls}`}>
                  {badge.label} · {Math.round(result.confidence * 100)}%
                </p>
              </div>
              {result.cached && (
                <span className="shrink-0 rounded-full bg-sky-100 px-2 py-0.5 text-[10px] font-medium text-sky-700 dark:bg-sky-900/40 dark:text-sky-300">
                  cached
                </span>
              )}
            </div>

            {/* Amount + Merchant — hero row */}
            <div className="rounded-xl border border-border bg-muted/30 p-4 flex items-center gap-4">
              <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-primary/10">
                <BadgeIndianRupee className="h-6 w-6 text-primary" />
              </div>
              <div className="min-w-0">
                <p className="text-2xl font-bold tabular-nums">
                  {currencySymbol}{result.amount.toLocaleString()}
                </p>
                <p className="text-sm text-muted-foreground truncate">{result.merchant || "Unknown merchant"}</p>
              </div>
            </div>

            {/* Field grid */}
            <div className="grid grid-cols-2 gap-2">
              {[
                {
                  label: "Category",
                  value: result.category || "—",
                  icon: CATEGORY_ICON[result.category?.toLowerCase()] ?? <Tag className="h-3.5 w-3.5" />,
                },
                { label: "Payment", value: result.payment_method || "—", icon: <CreditCard className="h-3.5 w-3.5" /> },
                { label: "Date", value: result.date || "—", icon: null },
                { label: "Type", value: result.expense_type || "spend", icon: null },
              ].map(({ label, value, icon }) => (
                <div key={label} className="rounded-lg border border-border bg-card p-2.5">
                  <p className="text-[10px] text-muted-foreground uppercase tracking-wide mb-1">{label}</p>
                  <p className="text-sm font-medium capitalize flex items-center gap-1 truncate">
                    {icon}
                    {value}
                  </p>
                </div>
              ))}
            </div>

            {/* Optional fields */}
            {(result.transaction_id || result.invoice_number || result.tax_amount > 0 || result.notes) && (
              <div className="space-y-1.5">
                {result.transaction_id && (
                  <div className="flex justify-between items-center text-sm px-1">
                    <span className="text-muted-foreground">Transaction ID</span>
                    <span className="font-mono text-xs font-medium truncate max-w-[60%]">{result.transaction_id}</span>
                  </div>
                )}
                {result.invoice_number && (
                  <div className="flex justify-between items-center text-sm px-1">
                    <span className="text-muted-foreground">Invoice No.</span>
                    <span className="font-mono text-xs font-medium">{result.invoice_number}</span>
                  </div>
                )}
                {result.tax_amount > 0 && (
                  <div className="flex justify-between items-center text-sm px-1">
                    <span className="text-muted-foreground">Tax / GST</span>
                    <span className="text-sm font-medium">{currencySymbol}{result.tax_amount}</span>
                  </div>
                )}
                {result.notes && (
                  <div className="rounded-lg border border-border bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
                    {result.notes}
                  </div>
                )}
              </div>
            )}

            {/* CTA buttons */}
            <div className="flex gap-2 pt-1">
              <button
                onClick={handleReset}
                className="flex items-center gap-1.5 rounded-xl border border-border px-4 py-2.5 text-sm font-medium hover:bg-muted transition-colors"
              >
                <RotateCcw className="h-3.5 w-3.5" />
                Rescan
              </button>
              <button
                onClick={handleConfirm}
                className="flex flex-1 items-center justify-center gap-2 rounded-xl bg-primary px-4 py-2.5 text-sm font-semibold text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                <CheckCircle2 className="h-4 w-4" />
                Confirm &amp; Fill Form
              </button>
            </div>
          </motion.div>
        );

      // ── Error ─────────────────────────────────────────────────────────────
      case "error":
        return (
          <motion.div key="error" initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} className="flex flex-col items-center gap-5 py-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircle className="h-8 w-8 text-destructive" />
            </div>
            <div className="text-center space-y-1">
              <p className="font-semibold">Something went wrong</p>
              <p className="text-sm text-muted-foreground max-w-xs">{error}</p>
            </div>
            <div className="flex gap-3">
              <button onClick={onClose} className="rounded-xl border border-border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors">
                Close
              </button>
              <button
                onClick={handleReset}
                className="flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                <RotateCcw className="h-4 w-4" />
                Try Again
              </button>
            </div>
          </motion.div>
        );
    }
  };

  // ── Modal shell ────────────────────────────────────────────────────────────

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
              if (phase !== "processing") onClose();
            }}
            className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
          />

          {/* Panel */}
          <motion.div
            key="panel"
            initial={{ opacity: 0, scale: 0.95, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 20 }}
            transition={{ type: "spring", stiffness: 350, damping: 30 }}
            className="fixed inset-x-4 bottom-4 z-50 mx-auto max-w-md rounded-2xl border border-border bg-card shadow-2xl overflow-hidden sm:inset-x-auto sm:left-1/2 sm:-translate-x-1/2 sm:w-full"
          >
            {/* Drag handle */}
            <div className="flex justify-center pt-3 pb-1">
              <div className="h-1 w-8 rounded-full bg-border" />
            </div>

            {/* Header bar */}
            <div className="flex items-center justify-between px-5 pb-3">
              <div className="flex items-center gap-2">
                <Upload className="h-4 w-4 text-primary" />
                <span className="text-sm font-semibold">Scan Receipt</span>
              </div>
              <button
                onClick={onClose}
                disabled={phase === "processing"}
                className="flex h-7 w-7 items-center justify-center rounded-lg hover:bg-muted text-muted-foreground transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <div className="border-t border-border" />

            {/* Body */}
            <div className="px-5 py-5 max-h-[75vh] overflow-y-auto">
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
