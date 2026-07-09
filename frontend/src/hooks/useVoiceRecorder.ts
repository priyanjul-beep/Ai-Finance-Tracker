/**
 * useVoiceRecorder
 *
 * Encapsulates all browser-level audio recording logic:
 *  - Requests microphone permission
 *  - Manages MediaRecorder lifecycle (start / pause / resume / stop / cancel)
 *  - Drives a real-time waveform via AudioContext AnalyserNode
 *  - Tracks recording duration
 *
 * The hook is intentionally decoupled from any API calls so it can be reused
 * across different voice-input surfaces.
 */
"use client";

import { useCallback, useEffect, useRef, useState } from "react";

// ─── Types ────────────────────────────────────────────────────────────────────

export type RecordingPhase =
  | "idle"
  | "requesting"   // asking for mic permission
  | "recording"
  | "paused"
  | "stopped";     // recording finished, blob ready

export interface VoiceRecorderState {
  phase: RecordingPhase;
  /** Elapsed recording time in seconds (paused intervals excluded). */
  duration: number;
  /** 32 amplitude samples (0-255) refreshed at ~30fps during recording. */
  waveform: number[];
  /** Human-readable error string, null when no error. */
  error: string | null;
}

export interface VoiceRecorderActions {
  /** Request permission and begin recording. */
  start: () => Promise<void>;
  pause: () => void;
  resume: () => void;
  /**
   * Stop recording and return the captured Blob.
   * Returns null if there was nothing recorded.
   */
  stop: () => Promise<Blob | null>;
  /** Discard the current recording and reset to idle. */
  cancel: () => void;
  /** Alias for cancel – used after a result has been processed. */
  reset: () => void;
}

export type UseVoiceRecorder = VoiceRecorderState & VoiceRecorderActions;

// ─── Hook ─────────────────────────────────────────────────────────────────────

const WAVEFORM_BARS = 40;
const SILENT_WAVEFORM = new Array(WAVEFORM_BARS).fill(0);

export function useVoiceRecorder(): UseVoiceRecorder {
  const [phase,    setPhase]    = useState<RecordingPhase>("idle");
  const [duration, setDuration] = useState(0);
  const [waveform, setWaveform] = useState<number[]>(SILENT_WAVEFORM);
  const [error,    setError]    = useState<string | null>(null);

  // Refs — never cause re-renders
  const recorderRef  = useRef<MediaRecorder | null>(null);
  const chunksRef    = useRef<Blob[]>([]);
  const streamRef    = useRef<MediaStream | null>(null);
  const analyserRef  = useRef<AnalyserNode | null>(null);
  const audioCtxRef  = useRef<AudioContext | null>(null);
  const timerRef     = useRef<ReturnType<typeof setInterval> | null>(null);
  const rafRef       = useRef<number | null>(null);
  const stopResolve  = useRef<((b: Blob | null) => void) | null>(null);

  // ── Internal helpers ────────────────────────────────────────────────────────

  /** Tear down all browser resources without touching React state. */
  const releaseResources = useCallback(() => {
    if (timerRef.current)  clearInterval(timerRef.current);
    if (rafRef.current)    cancelAnimationFrame(rafRef.current);
    if (audioCtxRef.current) {
      audioCtxRef.current.close().catch(() => {});
      audioCtxRef.current = null;
    }
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((t) => t.stop());
      streamRef.current = null;
    }
    analyserRef.current = null;
    recorderRef.current = null;
    chunksRef.current   = [];
    timerRef.current    = null;
    rafRef.current      = null;
  }, []);

  /** Animate the waveform by sampling the AnalyserNode frequency data. */
  const animateWaveform = useCallback(() => {
    if (!analyserRef.current) return;
    const data = new Uint8Array(analyserRef.current.frequencyBinCount);
    analyserRef.current.getByteFrequencyData(data);

    const step = Math.floor(data.length / WAVEFORM_BARS);
    const bars = Array.from({ length: WAVEFORM_BARS }, (_, i) => data[i * step] ?? 0);
    setWaveform(bars);
    rafRef.current = requestAnimationFrame(animateWaveform);
  }, []);

  /** Start the duration clock. */
  const startTimer = useCallback(() => {
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = setInterval(() => setDuration((d) => d + 1), 1_000);
  }, []);

  /** Stop the duration clock without resetting the value. */
  const stopTimer = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // ── Cleanup on unmount ──────────────────────────────────────────────────────

  useEffect(() => () => releaseResources(), [releaseResources]);

  // ── Public API ───────────────────────────────────────────────────────────────

  const start = useCallback(async () => {
    setError(null);
    setPhase("requesting");

    // ── 1. Request microphone ─────────────────────────────────────────────────
    let stream: MediaStream;
    try {
      stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          sampleRate: 16_000,
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
        },
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      let userMsg: string;
      if (msg.includes("Permission denied") || msg.includes("NotAllowedError") || msg.includes("PermissionDeniedError")) {
        userMsg = "Microphone access was denied. Please allow microphone permission in your browser settings and try again.";
      } else if (msg.includes("NotFoundError") || msg.includes("DevicesNotFoundError")) {
        userMsg = "No microphone detected. Please connect a microphone and try again.";
      } else if (msg.includes("NotReadableError") || msg.includes("TrackStartError")) {
        userMsg = "Your microphone is in use by another application. Please close it and try again.";
      } else {
        userMsg = "Could not access the microphone. Please try again.";
      }
      setError(userMsg);
      setPhase("idle");
      return;
    }

    streamRef.current = stream;

    // ── 2. Wire up AudioContext for waveform ──────────────────────────────────
    try {
      const audioCtx = new AudioContext();
      audioCtxRef.current = audioCtx;
      const source   = audioCtx.createMediaStreamSource(stream);
      const analyser = audioCtx.createAnalyser();
      analyser.fftSize = 128; // 64 frequency bins → enough for 40-bar waveform
      analyser.smoothingTimeConstant = 0.8;
      source.connect(analyser);
      analyserRef.current = analyser;
    } catch {
      // Waveform is purely cosmetic; continue without it
    }

    // ── 3. Choose MIME type ───────────────────────────────────────────────────
    const preferredTypes = [
      "audio/webm;codecs=opus",
      "audio/webm",
      "audio/ogg;codecs=opus",
      "audio/mp4",
    ];
    const mimeType = preferredTypes.find((t) => MediaRecorder.isTypeSupported(t)) ?? "";

    // ── 4. Set up MediaRecorder ───────────────────────────────────────────────
    const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined);
    recorderRef.current = recorder;
    chunksRef.current   = [];

    recorder.ondataavailable = (e) => {
      if (e.data && e.data.size > 0) chunksRef.current.push(e.data);
    };

    recorder.start(100); // emit data every 100 ms for lower latency
    setPhase("recording");
    setDuration(0);
    startTimer();
    animateWaveform();
  }, [animateWaveform, startTimer]);

  const pause = useCallback(() => {
    if (recorderRef.current?.state === "recording") {
      recorderRef.current.pause();
      stopTimer();
      if (rafRef.current) { cancelAnimationFrame(rafRef.current); rafRef.current = null; }
      setWaveform(SILENT_WAVEFORM);
      setPhase("paused");
    }
  }, [stopTimer]);

  const resume = useCallback(() => {
    if (recorderRef.current?.state === "paused") {
      recorderRef.current.resume();
      startTimer();
      animateWaveform();
      setPhase("recording");
    }
  }, [startTimer, animateWaveform]);

  const stop = useCallback((): Promise<Blob | null> => {
    return new Promise((resolve) => {
      stopTimer();
      if (rafRef.current) { cancelAnimationFrame(rafRef.current); rafRef.current = null; }
      setWaveform(SILENT_WAVEFORM);

      const recorder = recorderRef.current;
      if (!recorder || recorder.state === "inactive") {
        releaseResources();
        setPhase("stopped");
        resolve(null);
        return;
      }

      // Store resolve so onstop can call it after the final chunk arrives
      stopResolve.current = resolve;

      recorder.onstop = () => {
        const blob = chunksRef.current.length > 0
          ? new Blob(chunksRef.current, { type: recorder.mimeType || "audio/webm" })
          : null;
        releaseResources();
        setPhase("stopped");
        stopResolve.current?.(blob);
        stopResolve.current = null;
      };

      recorder.stop();
    });
  }, [stopTimer, releaseResources]);

  const cancel = useCallback(() => {
    stopTimer();
    if (rafRef.current) cancelAnimationFrame(rafRef.current);
    releaseResources();
    setPhase("idle");
    setDuration(0);
    setWaveform(SILENT_WAVEFORM);
    setError(null);
    stopResolve.current?.(null);
    stopResolve.current = null;
  }, [stopTimer, releaseResources]);

  const reset = cancel; // semantic alias

  return { phase, duration, waveform, error, start, pause, resume, stop, cancel, reset };
}
