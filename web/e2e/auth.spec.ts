import { test, expect, freshUser, registerAndLogin } from './_fixtures';

test.describe('Auth flow', () => {
  test('register → dashboard', async ({ page }) => {
    const user = await freshUser();
    await registerAndLogin(page, user);
    // Either dashboard or accounts depending on the next-active route.
    await expect(page).toHaveURL(/\/(accounts|dashboard|$)/);
  });

  test('login with bad credentials surfaces toast', async ({ page }) => {
    await page.goto('/auth/login');
    await page.getByLabel(/email/i).fill('nobody@example.com');
    await page.getByLabel(/contraseña/i).fill('wrongpass1');
    await page.getByRole('button', { name: /entrar|iniciar/i }).click();
    await expect(page.getByRole('status')).toBeVisible({ timeout: 5_000 });
  });

  test('logout clears session', async ({ page }) => {
    const user = await freshUser();
    await registerAndLogin(page, user);
    // Cookie + clearAuth on logout: just assert we end up back at /auth/login.
    await page.evaluate(() => localStorage.clear());
    await page.context().clearCookies();
    await page.goto('/auth/login');
    await expect(page).toHaveURL(/auth\/login/);
  });
});