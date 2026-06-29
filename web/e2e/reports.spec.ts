import { test, expect, freshUser, registerAndLogin } from './_fixtures';

test.describe('Navigation + reports', () => {
  test('dashboard loads for new user without crashing', async ({ page }) => {
    const user = await freshUser();
    await registerAndLogin(page, user);
    await page.goto('/');
    // Empty-state copy or stats — at minimum the main element is there.
    await expect(page.locator('main')).toBeVisible();
  });

  test('reports renders period selector and loading or empty state', async ({ page }) => {
    const user = await freshUser();
    await registerAndLogin(page, user);
    await page.goto('/reports');
    await expect(page.getByRole('heading', { name: /reportes/i })).toBeVisible();
    await expect(page.getByRole('group', { name: /período/i })).toBeVisible();
  });

  test('404 page renders for unknown route', async ({ page }) => {
    await page.goto('/this-route-does-not-exist');
    // Without auth this may redirect to login; with auth it shows the error page.
    const url = page.url();
    if (/auth\/login/.test(url)) {
      await expect(page.getByLabel(/email/i)).toBeVisible();
    } else {
      await expect(page.getByText(/no encontramos esa página/i)).toBeVisible();
    }
  });
});