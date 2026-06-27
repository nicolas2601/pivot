import { browser } from '$app/environment';
import { goto } from '$app/navigation';

const TOKEN_KEY = 'access_token';

export function getAccessToken(): string | null {
  if (!browser) return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setAccessToken(token: string): void {
  if (!browser) return;
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearAccessToken(): void {
  if (!browser) return;
  localStorage.removeItem(TOKEN_KEY);
}

/**
 * Patches global fetch to add Authorization header and handle 401 by trying
 * a refresh once, then redirecting to /auth/login if refresh fails.
 *
 * Must be called once from onMount in root +layout.svelte.
 */
export function installAuthInterceptor(): void {
  if (!browser) return;
  if ((window as unknown as { __authInterceptorInstalled?: boolean }).__authInterceptorInstalled) {
    return;
  }
  (window as unknown as { __authInterceptorInstalled?: boolean }).__authInterceptorInstalled = true;

  const originalFetch = window.fetch.bind(window);

  window.fetch = async (input: RequestInfo | URL, init: RequestInit = {}) => {
    const token = getAccessToken();
    const headers = new Headers(init.headers);
    if (token && !headers.has('Authorization')) {
      headers.set('Authorization', `Bearer ${token}`);
    }

    const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;

    let response = await originalFetch(input, { ...init, headers });

    if (
      response.status === 401 &&
      token &&
      !url.includes('/auth/login') &&
      !url.includes('/auth/register') &&
      !url.includes('/auth/refresh')
    ) {
      // Try refreshing once
      const refreshResp = await originalFetch('/api/v1/auth/refresh', {
        method: 'POST',
        credentials: 'include'
      });
      if (refreshResp.ok) {
        const data = await refreshResp.json();
        setAccessToken(data.access_token);
        const retryHeaders = new Headers(init.headers);
        retryHeaders.set('Authorization', `Bearer ${data.access_token}`);
        response = await originalFetch(input, { ...init, headers: retryHeaders });
      } else {
        clearAccessToken();
        await goto('/auth/login');
      }
    }

    return response;
  };
}