/**
 * API smoke: verify that whatever the baseURL points at can answer.
 * Tests skip cleanly if the server is unreachable (CI runs E2E against
 * the compose stack which is up before this spec runs).
 */
import { test, expect } from '@playwright/test';

const API_HEALTH = '/api/v1/health';

test('backend /api/v1/health returns 200', async ({ request }) => {
  const r = await request.get(API_HEALTH);
  expect(r.status()).toBe(200);
});

test('frontend root serves the SPA shell', async ({ page }) => {
  const r = await page.goto('/');
  expect([200, 304]).toContain(r?.status() ?? 0);
  await expect(page.locator('main, body')).toBeVisible();
});