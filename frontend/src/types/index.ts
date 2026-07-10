// All TypeScript interfaces mirroring backend DTOs.

// ─── Common ──────────────────────────────────────────────────────────────────

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface APIError {
  error: string;
  code?: string;
  details?: Record<string, string>;
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

export interface User {
  id: string;
  email: string;
  name: string;
  profile_picture?: string;
  is_email_verified: boolean;
  timezone: string;
  currency: string;
  preferred_language: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_at: string;
}

export interface SignupRequest {
  name: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface UpdateUserRequest {
  name?: string;
  timezone?: string;
  currency?: string;
  preferred_language?: string;
}

// ─── Expenses ─────────────────────────────────────────────────────────────────

export interface Tag {
  id: string;
  name: string;
  color: string;
}

export interface Expense {
  id: string;
  user_id: string;
  amount: number;
  currency: string;
  category: string;
  merchant: string;
  description: string;
  notes?: string;
  date: string;
  expense_type: "spend" | "refund" | "split";
  payment_method: string;
  image_url?: string;
  is_duplicate: boolean;
  is_favorite: boolean;
  tags: Tag[];
  created_at: string;
  updated_at: string;
}

export interface CreateExpenseRequest {
  amount: number;
  currency?: string;
  category: string;
  merchant?: string;
  description?: string;
  notes?: string;
  date: string;
  expense_type?: string;
  payment_method?: string;
  tag_ids?: string[];
}

export interface UpdateExpenseRequest extends Partial<CreateExpenseRequest> {}

export interface AIExpenseParseRequest {
  text?: string;
  image_url?: string;
}

export interface AIExpenseParsed {
  amount: number;
  currency: string;
  merchant: string;
  category: string;
  description: string;
  date: string;
  payment_method: string;
  confidence: number;
}

/** Mirrors backend dto.AIVoiceParseResponse */
export interface AIVoiceParseResponse {
  transcript: string;
  amount: number;
  merchant: string;
  category: string;
  /** Raw string: "today" | "yesterday" | "YYYY-MM-DD" */
  date: string;
  notes: string;
  expense_type: string;
  payment_method: string;
  confidence: number;
  cached: boolean;
}

/** Mirrors backend dto.AIReceiptScanResponse */
export interface AIReceiptScanResponse {
  amount: number;
  merchant: string;
  category: string;
  payment_method: string;
  expense_type: string;
  currency: string;
  /** YYYY-MM-DD */
  date: string;
  transaction_id: string;
  notes: string;
  tax_amount: number;
  invoice_number: string;
  confidence: number;
  raw_text?: string;
  cached: boolean;
}

// ─── Income ──────────────────────────────────────────────────────────────────

export interface Income {
  id: string;
  user_id: string;
  amount: number;
  currency: string;
  source: string;
  category: string;
  description: string;
  notes?: string;
  date: string;
  payment_method: string;
  is_taxable: boolean;
  tax_amount: number;
  created_at: string;
  updated_at: string;
}

export interface CreateIncomeRequest {
  amount: number;
  currency?: string;
  source: string;
  category?: string;
  description?: string;
  notes?: string;
  date: string;
  payment_method?: string;
  is_taxable?: boolean;
  tax_amount?: number;
}

export interface UpdateIncomeRequest extends Partial<CreateIncomeRequest> {}

// ─── Budget ──────────────────────────────────────────────────────────────────

export interface Budget {
  id: string;
  user_id: string;
  category: string;
  amount: number;
  currency: string;
  period: string;
  month: number;
  year: number;
  alert_at: number;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface BudgetStatus extends Budget {
  spent: number;
  remaining: number;
  percent: number;
  status: "on-track" | "warning" | "over-budget";
}

export interface CreateBudgetRequest {
  category: string;
  amount: number;
  currency?: string;
  period?: string;
  month?: number;
  year: number;
  alert_at?: number;
  description?: string;
}

// ─── Subscription ─────────────────────────────────────────────────────────────

export interface Subscription {
  id: string;
  user_id: string;
  name: string;
  amount: number;
  currency: string;
  billing_cycle: string;
  next_billing_date: string;
  category: string;
  payment_method: string;
  notes?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateSubscriptionRequest {
  name: string;
  amount: number;
  currency?: string;
  billing_cycle?: string;
  next_billing_date: string;
  category?: string;
  payment_method?: string;
  notes?: string;
}

// ─── Goal ─────────────────────────────────────────────────────────────────────

export interface Goal {
  id: string;
  user_id: string;
  name: string;
  description?: string;
  target_amount: number;
  current_amount: number;
  currency: string;
  category?: string;
  target_date: string;
  priority: number;
  status: "active" | "completed" | "paused";
  progress_percent: number;
  days_remaining: number;
  monthly_savings_needed: number;
  created_at: string;
  updated_at: string;
}

export interface CreateGoalRequest {
  name: string;
  description?: string;
  target_amount: number;
  current_amount?: number;
  currency?: string;
  category?: string;
  target_date: string;
  priority?: number;
}

// ─── Analytics ───────────────────────────────────────────────────────────────

export interface CategorySpend {
  category: string;
  amount: number;
}

export interface MerchantSpend {
  merchant: string;
  amount: number;
}

export interface PredictionData {
  next_month_prediction: number;
  confidence: number;
  by_category: CategorySpend[];
  trend: "increasing" | "decreasing" | "stable";
  message: string;
}

export interface Dashboard {
  total_balance: number;
  total_income: number;
  total_expenses: number;
  savings_rate: number;
  monthly_spend: number;
  weekly_spend: number;
  spend_by_category: CategorySpend[];
  spend_by_merchant: MerchantSpend[];
  recent_expenses: Expense[];
  upcoming_subscriptions: Subscription[];
  predictions: PredictionData;
}

export interface MonthlyReport {
  month: number;
  year: number;
  total_income: number;
  total_expenses: number;
  net_savings: number;
  savings_rate: number;
  top_categories: CategorySpend[];
  top_merchants: MerchantSpend[];
  budget_summary: BudgetStatus[];
  ai_recommendations: string[];
}

export interface FinancialHealthScore {
  score: number;
  income_score: number;
  savings_score: number;
  expense_ratio: number;
  budget_health: number;
  debt_health: number;
  subscription_health: number;
  insights: string[];
  updated_at: string;
}

// ─── Notification ────────────────────────────────────────────────────────────

export type NotificationType =
  | "welcome"
  | "budget_alert"
  | "budget_exceeded"
  | "budget_warning"
  | "monthly_summary"
  | "weekly_summary"
  | "expense_reminder"
  | "goal_achievement"
  | "general";

export type NotificationPriority = "low" | "medium" | "high" | "critical";

export interface Notification {
  id: string;
  user_id: string;
  title: string;
  message: string;
  type: NotificationType | string;
  priority: NotificationPriority | string;
  is_read: boolean;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface NotificationListResponse {
  notifications: Notification[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
  unread_count: number;
}

// ─── Filters / params ────────────────────────────────────────────────────────

export interface ExpenseFilters {
  [key: string]: unknown;
  page?: number;
  limit?: number;
  category?: string;
  merchant?: string;
  from?: string;
  to?: string;
  search?: string;
  sort?: "date" | "amount";
  order?: "asc" | "desc";
}

export interface IncomeFilters {
  [key: string]: unknown;
  page?: number;
  limit?: number;
  from?: string;
  to?: string;
}

// ─── Chart data ──────────────────────────────────────────────────────────────

export interface ChartDataPoint {
  name: string;
  value: number;
  color?: string;
}

export interface TimeSeriesPoint {
  date: string;
  income: number;
  expenses: number;
  savings: number;
}
