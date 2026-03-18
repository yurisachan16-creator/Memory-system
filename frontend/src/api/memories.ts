import { client, unwrap } from "./client";
import type {
  CreateMemoryPayload,
  DeleteMemoryResult,
  ListMemoriesParams,
  ListMemoriesResult,
  Memory,
  SearchMemoriesParams,
  SearchMemoriesResult,
  SummaryPayload,
  UpdateMemoryPayload
} from "../types/memory";

export async function listMemories(params: ListMemoriesParams) {
  return unwrap<ListMemoriesResult>(client.get("/memories", { params }));
}

export async function createMemory(payload: CreateMemoryPayload) {
  return unwrap<Memory>(client.post("/memories", payload));
}

export async function updateMemory(id: number, payload: UpdateMemoryPayload) {
  return unwrap<Memory>(client.put(`/memories/${id}`, payload));
}

export async function deleteMemory(id: number, userId: string) {
  return unwrap<DeleteMemoryResult>(client.delete(`/memories/${id}`, { params: { user_id: userId } }));
}

export async function searchMemories(params: SearchMemoriesParams) {
  return unwrap<SearchMemoriesResult>(client.get("/memories/search", { params }));
}

export async function getSummary(userId: string) {
  return unwrap<SummaryPayload>(client.get("/memories/summary", { params: { user_id: userId } }));
}
