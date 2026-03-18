export type Category = "preference" | "identity" | "goal" | "context";
export type Source = "chat" | "manual" | "system";

export interface Memory {
  id: number;
  user_id: string;
  content: string;
  category: Category;
  source: Source;
  importance: number;
  content_hash: string;
  is_deleted: boolean;
  created_at: string;
  updated_at: string;
}

export interface ListMemoriesResult {
  items: Memory[];
  total: number;
  page: number;
  page_size: number;
}

export interface ListMemoriesParams {
  user_id: string;
  category?: string;
  sort_by?: "created_at" | "importance";
  order?: "asc" | "desc";
  page?: number;
  page_size?: number;
}

export interface CreateMemoryPayload {
  user_id: string;
  content: string;
  category: Category;
  source: Source;
  importance: number;
}

export interface UpdateMemoryPayload {
  user_id: string;
  content: string;
  category: Category;
  importance: number;
}

export interface DeleteMemoryResult {
  id: number;
}

export interface SearchItem {
  memory: Memory;
  relevance_score: number;
  recency_score: number;
  final_score: number;
  matched_terms: string[];
}

export interface SearchMemoriesParams {
  user_id: string;
  query: string;
  limit?: number;
}

export interface SearchMemoriesResult {
  items: SearchItem[];
  count: number;
  cached: boolean;
}

export interface SummaryResult {
  preferences: Memory[];
  goals: Memory[];
  background: Memory[];
  recent: Memory[];
}

export interface SummaryPayload {
  summary: SummaryResult;
  cached: boolean;
}

export const CATEGORY_OPTIONS: Array<{ label: string; value: Category }> = [
  { label: "Preference", value: "preference" },
  { label: "Identity", value: "identity" },
  { label: "Goal", value: "goal" },
  { label: "Context", value: "context" }
];

export const SOURCE_OPTIONS: Array<{ label: string; value: Source }> = [
  { label: "Chat", value: "chat" },
  { label: "Manual", value: "manual" },
  { label: "System", value: "system" }
];

export const categoryLabelMap: Record<Category, string> = {
  preference: "Preference",
  identity: "Identity",
  goal: "Goal",
  context: "Context"
};

export const sourceLabelMap: Record<Source, string> = {
  chat: "Chat",
  manual: "Manual",
  system: "System"
};

export function formatDateTime(value: string) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("en-US", {
    year: "numeric",
    month: "short",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}
