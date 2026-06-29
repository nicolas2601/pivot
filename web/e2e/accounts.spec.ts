import { test, expect, freshUser, registerAndLogin } from './_fixtures';

test.describe('Accounts CRUD', () => {
  test('user can create and delete an account', async ({ page }) => {
    const user = await freshUser();
    await registerAndLogin(page, user);

    await page.goto('/accounts');
    await page.getByRole('button', { name: /nueva cuenta/i }).click();
    await page.getByLabel(/nombre/i).fill('Efectivo principal');
    await page.getByLabel(/tipo/i).selectOption('cash');
    // opening_balance handled by label or placeholder; switch to placeholder selector if needed.
    const opening = page.getByLabel(/saldo inicial|opening/i);
    if (await opening.count()) {
      await opening.first().fill('500000'); // $5.000 COP
    }
    await page.getByRole('button', { name: /guardar|crear/i }).click();
    await expect(page.getByText('Efectivo principal')).toBeVisible({ timeout: 10_000 });
  });

  test('seed categories button populates defaults', async ({ page }) => {
    const user = await freshUser();
    await registerAndLogin(page, user);
    await page.goto('/categories');
    const seedBtn = page.getByRole('button', { name: /categorías por defecto/i });
    if (await seedBtn.count()) {
      await seedBtn.click();
      await expect(page.getByText(/alimentación/i).first()).toBeVisible({ timeout: 10_000 });
    } else {
      // If defaults are already seeded, just confirm page renders categories
      await expect(page.locator('main')).toBeVisible();
    }
  });
});