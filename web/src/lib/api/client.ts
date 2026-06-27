import { PUBLIC_API_URL } from '$env/static/public';
import { ApiException, type ApiError } from '$lib/utils/api-error';

const BASE_URL = PUBLIC_API_URL;

export interface ApiOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
}

export async function apiFetch<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const { body, headers, ...rest } = options;

  const response = await fetch(`${BASE_URL}${path}`, {
    ...rest,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...headers
    },
    body: body !== undefined ? JSON.stringify(body) : undefined
  });

  if (!response.ok) {
    let apiError: ApiError;
    try {
      apiError = await response.json();
    } catch {
      apiError = { code: 'NETWORK_ERROR', message: 'Error de red' };
    }
    throw new ApiException(response.status, apiError);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}