import { apiFetch } from './client';
import { UserSchema, type User } from '$lib/schemas/auth';

export interface AuthResponse {
  user: User;
  access_token: string;
}

export async function login(input: { email: string; password: string }): Promise<AuthResponse> {
  const response = await apiFetch<AuthResponse>('/auth/login', {
    method: 'POST',
    body: input
  });
  response.user = UserSchema.parse(response.user);
  return response;
}

export async function register(input: {
  email: string;
  password: string;
  display_name?: string;
}): Promise<AuthResponse> {
  const response = await apiFetch<AuthResponse>('/auth/register', {
    method: 'POST',
    body: input
  });
  response.user = UserSchema.parse(response.user);
  return response;
}

export async function logout(): Promise<void> {
  await apiFetch<void>('/auth/logout', { method: 'POST' });
}

export async function me(): Promise<User> {
  const response = await apiFetch<{ user: User }>('/auth/me');
  return UserSchema.parse(response.user);
}

export async function refresh(): Promise<AuthResponse> {
  const response = await apiFetch<AuthResponse>('/auth/refresh', { method: 'POST' });
  response.user = UserSchema.parse(response.user);
  return response;
}