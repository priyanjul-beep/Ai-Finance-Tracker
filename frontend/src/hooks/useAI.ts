"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";
import { expenseService } from "@/services/expense.service";
import type { AIExpenseParsed, CreateExpenseRequest } from "@/types";

export function useAIExpenseParse() {
  const [parsed, setParsed] = useState<AIExpenseParsed | null>(null);

  const mutation = useMutation({
    mutationFn: ({
      text,
      imageUrl,
    }: {
      text?: string;
      imageUrl?: string;
    }) => expenseService.parseWithAI(text, imageUrl),
    onSuccess: (data) => {
      setParsed(data);
      toast.success("Expense parsed by AI!");
    },
    onError: () => {
      toast.error("AI parsing failed. Please enter manually.");
    },
  });

  const reset = () => setParsed(null);

  return {
    parsed,
    parseText: (text: string) => mutation.mutate({ text }),
    parseImage: (imageUrl: string) => mutation.mutate({ imageUrl }),
    isLoading: mutation.isPending,
    reset,

    // Convert parsed result to CreateExpenseRequest
    toCreateRequest: (): Partial<CreateExpenseRequest> | null => {
      if (!parsed) return null;
      return {
        amount: parsed.amount,
        currency: parsed.currency,
        merchant: parsed.merchant,
        category: parsed.category,
        description: parsed.description,
        date: parsed.date,
        payment_method: parsed.payment_method,
      };
    },
  };
}
