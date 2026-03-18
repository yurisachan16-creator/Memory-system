import axios from "axios";

export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

const client = axios.create({
  baseURL: "/api/v1",
  timeout: 10000
});

function extractErrorMessage(error: unknown) {
  if (axios.isAxiosError(error)) {
    const envelope = error.response?.data as Partial<ApiEnvelope<unknown>> | undefined;
    return envelope?.message || error.message || "Request failed";
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Request failed";
}

export async function unwrap<T>(request: Promise<{ data: ApiEnvelope<T> }>): Promise<T> {
  try {
    const response = await request;
    if (response.data.code !== 0) {
      throw new Error(response.data.message || "Request failed");
    }
    return response.data.data;
  } catch (error) {
    throw new Error(extractErrorMessage(error));
  }
}

export { client };
